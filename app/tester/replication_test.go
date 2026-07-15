package tester

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"CacheDB/app/RESP"
	"CacheDB/app/server"
)

// ===========================================================================
// Helpers
// ===========================================================================

func getFreePort(t *testing.T) int {
	t.Helper()
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to resolve address: %v", err)
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		t.Fatalf("failed to bind ephemeral port: %v", err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

// encodeCommand builds a RESP array command from plain string arguments.
func encodeCommand(args ...string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "*%d\r\n", len(args))
	for _, a := range args {
		fmt.Fprintf(&b, "$%d\r\n%s\r\n", len(a), a)
	}
	return b.String()
}

// readRESP reads exactly one complete RESP reply from reader and returns its raw wire bytes.
func readRESP(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read reply header: %w", err)
	}
	if len(line) == 0 {
		return "", fmt.Errorf("empty reply header")
	}

	switch line[0] {
	case '+', '-', ':':
		return line, nil

	case '$':
		var size int
		if _, err := fmt.Sscanf(line, "$%d\r\n", &size); err != nil {
			return "", fmt.Errorf("failed to parse bulk length from %q: %w", line, err)
		}
		if size < 0 {
			return line, nil // null bulk string
		}
		payload := make([]byte, size+2) // +2 for trailing CRLF
		if _, err := io.ReadFull(reader, payload); err != nil {
			return "", fmt.Errorf("failed to read bulk payload (%d bytes): %w", size, err)
		}
		return line + string(payload), nil

	case '*':
		var count int
		if _, err := fmt.Sscanf(line, "*%d\r\n", &count); err != nil {
			return "", fmt.Errorf("failed to parse array length from %q: %w", line, err)
		}
		full := line
		for i := 0; i < count; i++ {
			elem, err := readRESP(reader)
			if err != nil {
				return "", fmt.Errorf("failed to read array element %d: %w", i, err)
			}
			full += elem
		}
		return full, nil

	default:
		return "", fmt.Errorf("unrecognized RESP type byte %q in line %q", line[0], line)
	}
}

// testConn wraps a connection with a persistent buffered reader so multiple
// commands can be sent over the same connection and replies read in order.
type testConn struct {
	conn   net.Conn
	reader *bufio.Reader
}

// Renamed from dial to dialWithReader to avoid conflict with dial in main_test.go
func dialWithReader(t *testing.T, port int) *testConn {
	t.Helper()
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("failed to connect to server on port %d: %v", port, err)
	}
	return &testConn{conn: conn, reader: bufio.NewReader(conn)}
}

func (c *testConn) send(t *testing.T, args ...string) string {
	t.Helper()
	if _, err := c.conn.Write([]byte(encodeCommand(args...))); err != nil {
		t.Fatalf("failed to write command %v: %v", args, err)
	}
	resp, err := readRESP(c.reader)
	if err != nil {
		t.Fatalf("failed to read response for command %v: %v", args, err)
	}
	return resp
}

func (c *testConn) close() { c.conn.Close() }

// sendRawCommand keeps the old single-shot signature: dial, write one raw payload, close.
func sendRawCommand(t *testing.T, port int, rawPayload string) string {
	t.Helper()
	tc := dialWithReader(t, port)
	defer tc.close()
	if _, err := tc.conn.Write([]byte(rawPayload)); err != nil {
		t.Fatalf("failed to write raw command payload: %v", err)
	}
	resp, err := readRESP(tc.reader)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	return resp
}

func startMaster(t *testing.T) (*RESP.SERVER, int) {
	t.Helper()
	port := getFreePort(t)
	cfg := &RESP.SERVER{
		Role:         "master",
		PORT:         port,
		MASTERREPLID: "8371b4fb115b71c4a0413b1db346e45071511224",
		REPLICAS:     make([]*RESP.REPLICA, 0),
	}
	go server.StartServer(cfg)
	time.Sleep(100 * time.Millisecond) // give the TCP listener time to bind
	return cfg, port
}

func startReplica(t *testing.T, masterPort int) (*RESP.SERVER, int) {
	t.Helper()
	port := getFreePort(t)
	cfg := &RESP.SERVER{
		Role:       "slave",
		PORT:       port,
		MasterHost: "127.0.0.1",
		MasterPort: masterPort,
	}
	go server.StartServer(cfg)
	return cfg, port
}

func waitForReplicaCount(cfg *RESP.SERVER, n int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		cfg.ReplicasMutex.RLock()
		count := len(cfg.REPLICAS)
		cfg.ReplicasMutex.RUnlock()
		if count >= n {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return false
}

func waitForValue(t *testing.T, port int, key, expected string, timeout time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp := sendRawCommand(t, port, encodeCommand("GET", key))
		if strings.Contains(resp, expected) {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return false
}

// fakeReplica speaks just enough of the handshake protocol to register as a replica.
func fakeReplica(t *testing.T, masterPort int) *testConn {
	t.Helper()
	tc := dialWithReader(t, masterPort)

	steps := []struct {
		payload      string
		expectPrefix string
	}{
		{"*1\r\n$4\r\nPING\r\n", "+PONG"},
		{"*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$5\r\n55555\r\n", "+OK"},
		{"*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n", "+OK"},
		{"*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n", "+FULLRESYNC"},
	}

	for _, step := range steps {
		if _, err := tc.conn.Write([]byte(step.payload)); err != nil {
			t.Fatalf("fake replica handshake write failed: %v", err)
		}
		resp, err := readRESP(tc.reader)
		if err != nil {
			t.Fatalf("fake replica handshake read failed: %v", err)
		}
		if !strings.HasPrefix(resp, step.expectPrefix) {
			t.Fatalf("fake replica handshake: expected prefix %q, got %q", step.expectPrefix, resp)
		}
	}

	return tc
}

// ===========================================================================
// TEST 1: Full Handshake Integration & Info Command
// ===========================================================================
func TestIntegration_ReplicaHandshake(t *testing.T) {
	masterConfig, masterPort := startMaster(t)
	_, _ = startReplica(t, masterPort)

	replicaRegistered := waitForReplicaCount(masterConfig, 1, 1*time.Second)

	infoResp := sendRawCommand(t, masterPort, encodeCommand("INFO", "replication"))
	if !strings.Contains(infoResp, "role: master") {
		t.Errorf("expected 'role: master' in INFO block, got: %q", infoResp)
	}

	if !replicaRegistered && !strings.Contains(infoResp, "connected_slaves:1") {
		t.Error("replica never registered in master's REPLICAS slice, and INFO doesn't report connected_slaves:1 either")
	}
}

// ===========================================================================
// TEST 2: Write Command Propagation
// ===========================================================================
func TestIntegration_WritePropagation(t *testing.T) {
	_, masterPort := startMaster(t)
	_, replicaPort := startReplica(t, masterPort)

	setResp := sendRawCommand(t, masterPort, encodeCommand("SET", "key", "value"))
	if !strings.Contains(setResp, "OK") {
		t.Fatalf("SET failed on master: %q", setResp)
	}

	if !waitForValue(t, replicaPort, "key", "value", 2*time.Second) {
		t.Fatal("replica failed to receive and store the propagated command from master")
	}
}

// ===========================================================================
// TEST 3: WAIT Command Logic
// ===========================================================================
func TestIntegration_WaitCommandLogic(t *testing.T) {
	_, masterPort := startMaster(t)
	_, _ = startReplica(t, masterPort)

	time.Sleep(250 * time.Millisecond) // let the handshake settle

	sendRawCommand(t, masterPort, encodeCommand("INCR", "count"))
	time.Sleep(200 * time.Millisecond)

	waitResp := sendRawCommand(t, masterPort, encodeCommand("WAIT", "1", "2000"))

	if waitResp != ":1\r\n" && waitResp != ":0\r\n" {
		t.Errorf("expected WAIT to return exactly ':0\\r\\n' or ':1\\r\\n', got: %q", waitResp)
	}
}

// ===========================================================================
// TEST 4: WAIT with zero connected replicas returns immediately
// ===========================================================================
func TestIntegration_WaitNoReplicasReturnsImmediately(t *testing.T) {
	_, masterPort := startMaster(t)

	start := time.Now()
	resp := sendRawCommand(t, masterPort, encodeCommand("WAIT", "1", "2000"))
	elapsed := time.Since(start)

	if resp != ":0\r\n" {
		t.Errorf("expected ':0\\r\\n' with no replicas connected, got: %q", resp)
	}
	if elapsed > 200*time.Millisecond {
		t.Errorf("WAIT with zero replicas should return immediately, took %v (timeout was 2000ms)", elapsed)
	}
}

// ===========================================================================
// TEST 5: WAIT actually respects its timeout when a replica never ACKs
// ===========================================================================
func TestIntegration_WaitTimeoutRespected(t *testing.T) {
	masterConfig, masterPort := startMaster(t)

	stuck := fakeReplica(t, masterPort)
	defer stuck.close()

	if !waitForReplicaCount(masterConfig, 1, 1*time.Second) {
		t.Fatal("fake replica never registered in master's REPLICAS slice")
	}

	const timeoutMs = 300
	start := time.Now()
	resp := sendRawCommand(t, masterPort, encodeCommand("WAIT", "1", fmt.Sprintf("%d", timeoutMs)))
	elapsed := time.Since(start)

	if resp != ":0\r\n" {
		t.Errorf("expected ':0\\r\\n' since the fake replica never ACKs, got: %q", resp)
	}
	if elapsed < timeoutMs*time.Millisecond {
		t.Errorf("WAIT returned early (%v) despite requiring 1 replica ack that never came; should have waited out the %dms timeout", elapsed, timeoutMs)
	}
	if elapsed > (timeoutMs+250)*time.Millisecond {
		t.Errorf("WAIT overran its timeout by too much: waited %v for a %dms timeout", elapsed, timeoutMs)
	}
}

// ===========================================================================
// TEST 6: Multiple replicas all receive propagated writes
// ===========================================================================
func TestIntegration_MultipleReplicas(t *testing.T) {
	masterConfig, masterPort := startMaster(t)
	_, replicaPort1 := startReplica(t, masterPort)
	_, replicaPort2 := startReplica(t, masterPort)

	if !waitForReplicaCount(masterConfig, 2, 2*time.Second) {
		t.Fatal("both replicas never registered with the master")
	}

	setResp := sendRawCommand(t, masterPort, encodeCommand("SET", "shared", "fanout"))
	if !strings.Contains(setResp, "OK") {
		t.Fatalf("SET failed on master: %q", setResp)
	}

	if !waitForValue(t, replicaPort1, "shared", "fanout", 2*time.Second) {
		t.Error("replica 1 never received the propagated write")
	}
	if !waitForValue(t, replicaPort2, "shared", "fanout", 2*time.Second) {
		t.Error("replica 2 never received the propagated write")
	}
}

// ===========================================================================
// TEST 7: Ordered propagation — sequential writes must apply in order
// ===========================================================================
func TestIntegration_OrderedPropagation(t *testing.T) {
	_, masterPort := startMaster(t)
	_, replicaPort := startReplica(t, masterPort)

	master := dialWithReader(t, masterPort)
	defer master.close()

	for _, v := range []string{"v1", "v2", "v3"} {
		resp := master.send(t, "SET", "seq", v)
		if !strings.Contains(resp, "OK") {
			t.Fatalf("SET seq=%s failed: %q", v, resp)
		}
	}

	if !waitForValue(t, replicaPort, "seq", "v3", 2*time.Second) {
		t.Fatal("replica never converged to the final value 'v3' — writes may have been reordered or dropped")
	}

	time.Sleep(100 * time.Millisecond)
	finalResp := sendRawCommand(t, replicaPort, encodeCommand("GET", "seq"))
	if !strings.Contains(finalResp, "v3") {
		t.Errorf("replica value for 'seq' regressed after settling: %q", finalResp)
	}
}

// ===========================================================================
// TEST 8: Replica offset advances after a write + WAIT round trip
// ===========================================================================
func TestIntegration_ReplicaOffsetTracking(t *testing.T) {
	masterConfig, masterPort := startMaster(t)
	_, _ = startReplica(t, masterPort)

	if !waitForReplicaCount(masterConfig, 1, 1*time.Second) {
		t.Fatal("replica never registered with the master")
	}

	sendRawCommand(t, masterPort, encodeCommand("SET", "k", "v"))
	sendRawCommand(t, masterPort, encodeCommand("WAIT", "1", "2000"))

	masterConfig.ReplicasMutex.RLock()
	defer masterConfig.ReplicasMutex.RUnlock()
	if len(masterConfig.REPLICAS) == 0 {
		t.Fatal("replica disappeared from REPLICAS slice")
	}
	offset := masterConfig.REPLICAS[0].Offset.Load()
	if offset <= 0 {
		t.Errorf("expected replica offset to advance above 0 after SET+WAIT, got %d", offset)
	}
}

// ===========================================================================
// TEST 9 (KNOWN BUG REGRESSION): pipelined commands on one connection
// ===========================================================================
func TestIntegration_PipelinedCommands_KnownBug(t *testing.T) {
	_, masterPort := startMaster(t)

	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", masterPort))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	pipelined := encodeCommand("SET", "pkey", "pval") + encodeCommand("GET", "pkey")
	if _, err := conn.Write([]byte(pipelined)); err != nil {
		t.Fatalf("failed to write pipelined commands: %v", err)
	}

	reader := bufio.NewReader(conn)

	firstResp, err := readRESP(reader)
	if err != nil {
		t.Fatalf("failed to read response to pipelined SET: %v", err)
	}
	if !strings.Contains(firstResp, "OK") {
		t.Fatalf("expected OK for pipelined SET, got: %q", firstResp)
	}

	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	secondResp, err := readRESP(reader)
	if err != nil {
		t.Fatalf("KNOWN BUG: server never responded to the second pipelined command (GET pkey) — "+
			"handleClient drops unconsumed bytes from the read buffer between loop iterations. "+
			"Underlying read error: %v", err)
	}
	if !strings.Contains(secondResp, "pval") {
		t.Fatalf("KNOWN BUG: expected pipelined GET to return 'pval', got: %q — "+
			"the second command in the pipeline was likely misparsed or ignored", secondResp)
	}
}