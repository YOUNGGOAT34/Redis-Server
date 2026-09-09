package server

// Regression tests for Issue 5: before the fix, StartServer's replica branch
// called syncWithMaster once and panic()'d on any failure - including the
// completely ordinary case of the master not being up yet. These tests
// exercise connectToMaster() directly (the extracted retry wrapper) with a
// hand-rolled fake master that speaks just enough of the handshake protocol
// to let syncWithMaster succeed, so retry/backoff timing can be tested
// deterministically and fast.

import (
	"bufio"
	"fmt"
	"net"
	"testing"
	"time"

	rdb "CacheDB/app/RDB"
	"CacheDB/app/config"
)

func newTestRDBConfig(t *testing.T) *rdb.RDB {
	t.Helper()
	return &rdb.RDB{
		Dir:        t.TempDir(),
		DbFileName: "dump.rdb",
	}
}

// fakeMaster speaks the minimal replica handshake: PING, REPLCONF x2, PSYNC,
// then a FULLRESYNC + a syntactically valid (empty) RDB payload.
func fakeMasterHandshake(t *testing.T, conn net.Conn) {
	t.Helper()
	reader := bufio.NewReader(conn)

	readCommand := func() {
		// crude RESP array reader: just consume until it stops looking like
		// more protocol is buffered by reading a line and, for arrays,
		// the declared number of bulk strings.
		line, err := reader.ReadString('\n')
		if err != nil || len(line) == 0 || line[0] != '*' {
			return
		}
		var n int
		fmt.Sscanf(line, "*%d", &n)
		for i := 0; i < n; i++ {
			lenLine, _ := reader.ReadString('\n')
			var l int
			fmt.Sscanf(lenLine, "$%d", &l)
			buf := make([]byte, l+2) // +2 for trailing \r\n
			_, _ = reader.Read(buf)
		}
	}

	readCommand() // PING
	conn.Write([]byte("+PONG\r\n"))

	readCommand() // REPLCONF listening-port
	conn.Write([]byte("+OK\r\n"))

	readCommand() // REPLCONF capa
	conn.Write([]byte("+OK\r\n"))

	readCommand() // PSYNC
	conn.Write([]byte("+FULLRESYNC 8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb 0\r\n"))

	// Minimal valid-enough RDB payload: receiveFullResync just reads a
	// length-prefixed blob ($<n>\r\n<n bytes>), it doesn't validate contents
	// in the handshake path itself.
	payload := []byte("REDIS0011\xff")
	conn.Write([]byte(fmt.Sprintf("$%d\r\n", len(payload))))
	conn.Write(payload)
}

func startFakeMasterNow(t *testing.T) (addr string, cleanup func()) {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start fake master listener: %v", err)
	}

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go fakeMasterHandshake(t, conn)
		}
	}()

	return l.Addr().String(), func() { l.Close() }
}

func TestConnectToMaster_SucceedsImmediatelyWhenMasterIsUp(t *testing.T) {
	addr, cleanup := startFakeMasterNow(t)
	defer cleanup()

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("failed to split addr: %v", err)
	}
	var port int
	fmt.Sscanf(portStr, "%d", &port)

	replConfig := &config.SERVER{MasterHost: host, MasterPort: port}
	rdbConfig := newTestRDBConfig(t)

	start := time.Now()
	conn, err := connectToMaster(replConfig, rdbConfig, nil)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("expected immediate success, got error: %v", err)
	}
	defer conn.Close()

	if elapsed > 200*time.Millisecond {
		t.Fatalf("expected near-instant connection when master is already up, took %s", elapsed)
	}
}

func TestConnectToMaster_RetriesUntilMasterBecomesAvailable(t *testing.T) {
	origInitial, origMax := initialReplicaRetryDelay, maxReplicaRetryDelay
	initialReplicaRetryDelay = 50 * time.Millisecond
	maxReplicaRetryDelay = 200 * time.Millisecond
	defer func() { initialReplicaRetryDelay, maxReplicaRetryDelay = origInitial, origMax }()

	// Reserve a port but don't listen on it yet - connectToMaster should
	// hit connection-refused and retry until we start listening.
	probe, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to reserve a port: %v", err)
	}
	addr := probe.Addr().String()
	probe.Close() // free it up, but keep the (host, port) to reuse

	host, portStr, _ := net.SplitHostPort(addr)
	var port int
	fmt.Sscanf(portStr, "%d", &port)

	replConfig := &config.SERVER{MasterHost: host, MasterPort: port}
	rdbConfig := newTestRDBConfig(t)

	resultCh := make(chan struct {
		conn net.Conn
		err  error
	}, 1)
	go func() {
		conn, err := connectToMaster(replConfig, rdbConfig, nil)
		resultCh <- struct {
			conn net.Conn
			err  error
		}{conn, err}
	}()

	// Give it a couple of failed attempts before the master appears.
	time.Sleep(250 * time.Millisecond)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("failed to start delayed fake master on %s: %v", addr, err)
	}
	defer l.Close()
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go fakeMasterHandshake(t, conn)
		}
	}()

	select {
	case res := <-resultCh:
		if res.err != nil {
			t.Fatalf("expected connectToMaster to eventually succeed after master became available, got: %v", res.err)
		}
		res.conn.Close()
	case <-time.After(5 * time.Second):
		t.Fatal("connectToMaster did not succeed within 5s of the master becoming available " +
			"(this is the Issue 5 regression: a temporarily unavailable master must not be fatal)")
	}
}

func TestConnectToMaster_StopsRetryingWhenStopSignalFires(t *testing.T) {
	origInitial, origMax := initialReplicaRetryDelay, maxReplicaRetryDelay
	initialReplicaRetryDelay = 50 * time.Millisecond
	maxReplicaRetryDelay = 100 * time.Millisecond
	defer func() { initialReplicaRetryDelay, maxReplicaRetryDelay = origInitial, origMax }()

	// A host that will never accept a connection (nothing listens here).
	probe, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to reserve a port: %v", err)
	}
	addr := probe.Addr().String()
	probe.Close()

	host, portStr, _ := net.SplitHostPort(addr)
	var port int
	fmt.Sscanf(portStr, "%d", &port)

	replConfig := &config.SERVER{MasterHost: host, MasterPort: port}
	rdbConfig := newTestRDBConfig(t)

	stop := make(chan struct{})
	resultCh := make(chan error, 1)
	go func() {
		_, err := connectToMaster(replConfig, rdbConfig, stop)
		resultCh <- err
	}()

	time.Sleep(150 * time.Millisecond) // let it fail and retry a couple times
	close(stop)

	select {
	case err := <-resultCh:
		if err == nil {
			t.Fatal("expected connectToMaster to return an error once stopped, got nil")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("connectToMaster did not stop retrying within 2s of the stop signal " +
			"(this is a busy-loop/hang risk: a permanently misconfigured replica must be able to shut down cleanly)")
	}
}
