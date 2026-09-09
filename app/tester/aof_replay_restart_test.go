package tester

// TestAOFReplayRestoresDataAcrossRestart is a regression test for Issue 2.
//
// Before the fix, replayAOF() dispatched every replayed command through
// dispatchCommands() using a bare &storage.Client{} (User == nil), and
// dispatchCommands() rejects any non-AUTH command with NOAUTH when
// client.User == nil. Since the default ACL user was also seeded AFTER
// replayAOF() ran in StartServer(), every single replayed command failed
// authentication and was silently dropped - AOF replay never actually
// restored any data.
//
// This test starts a real server, writes a key (which lands in the AOF),
// stops the process, restarts it against the same --dir, and asserts the
// key is still present.

import (
	"bufio"
	"net"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"
)

func mustBuildCacheDB(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine test file location")
	}
	appDir := filepath.Join(filepath.Dir(thisFile), "..")

	binPath := filepath.Join(t.TempDir(), "cachedb_under_test_replay")
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = appDir
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("failed to build cachedb binary: %v\n%s", err, out)
	}
	return binPath
}

func dialWithRetry(t *testing.T, addr string) net.Conn {
	t.Helper()
	var conn net.Conn
	var err error
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		conn, err = net.Dial("tcp", addr)
		if err == nil {
			return conn
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("could not connect to %s: %v", addr, err)
	return nil
}

func sendAndReadLine(t *testing.T, conn net.Conn, reader *bufio.Reader, req string) string {
	t.Helper()
	if _, err := conn.Write([]byte(req)); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	return line
}

func TestAOFReplayRestoresDataAcrossRestart(t *testing.T) {
	binPath := mustBuildCacheDB(t)
	dataDir := t.TempDir()

	startAndSet := func(port int) {
		cmd := exec.Command(binPath,
			"--port", strconv.Itoa(port),
			"--dir", dataDir,
			"--appendonly", "yes",
			"--appendfilename", "appendonly.aof",
		)
		if err := cmd.Start(); err != nil {
			t.Fatalf("failed to start cachedb: %v", err)
		}
		defer func() {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}()

		addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
		conn := dialWithRetry(t, addr)
		defer conn.Close()
		reader := bufio.NewReader(conn)

		line := sendAndReadLine(t, conn, reader, "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n")
		if line != "+OK\r\n" {
			t.Fatalf("expected +OK for SET, got %q", line)
		}

		time.Sleep(150 * time.Millisecond) // let the AOF write flush to disk
	}

	port1 := getFreePort(t)
	startAndSet(port1)

	// Restart against the SAME --dir: this must replay the AOF and restore "foo".
	port2 := getFreePort(t)
	cmd := exec.Command(binPath,
		"--port", strconv.Itoa(port2),
		"--dir", dataDir,
		"--appendonly", "yes",
		"--appendfilename", "appendonly.aof",
	)
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start cachedb (restart): %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(port2))
	conn := dialWithRetry(t, addr)
	defer conn.Close()
	reader := bufio.NewReader(conn)

	if _, err := conn.Write([]byte("*2\r\n$3\r\nGET\r\n$3\r\nfoo\r\n")); err != nil {
		t.Fatalf("failed to send GET: %v", err)
	}
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("failed to read GET status line: %v", err)
	}

	if line == "$-1\r\n" {
		t.Fatalf("AOF replay did not restore 'foo' after restart - got nil " +
			"(this is the Issue 2 regression: replayed commands must not be rejected as NOAUTH)")
	}
	if line != "$3\r\n" {
		t.Fatalf("expected bulk string header $3, got %q", line)
	}
	valueLine, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("failed to read GET value: %v", err)
	}
	if valueLine != "bar\r\n" {
		t.Fatalf("expected replayed value 'bar', got %q", valueLine)
	}
}
