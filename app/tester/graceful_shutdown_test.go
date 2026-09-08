package tester

// Regression tests for Issue 4: graceful shutdown.
//
// Before the fix, CacheDB had no signal.Notify/context handling at all, so
// SIGTERM (what `docker stop` sends) killed the process immediately with no
// chance to close the AOF file, replica connections, or the listener
// cleanly. These tests exercise the real compiled binary and real OS
// signals, since graceful shutdown is fundamentally about process-level
// behavior that a pure in-process unit test can't observe.

import (
	"net"
	"os/exec"
	"strconv"
	"syscall"
	"testing"
	"time"
)

func buildCacheDBCommand(binPath string, port int, dataDir string, appendonly string) *exec.Cmd {
	return exec.Command(binPath,
		"--port", strconv.Itoa(port),
		"--dir", dataDir,
		"--appendonly", appendonly,
		"--appendfilename", "appendonly.aof",
	)
}

func TestSIGTERMExitsCleanly(t *testing.T) {
	binPath := mustBuildCacheDB(t)
	dataDir := t.TempDir()
	port := getFreePort(t)

	cmd := buildCacheDBCommand(binPath, port, dataDir, "yes")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start cachedb: %v", err)
	}

	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	conn := dialWithRetry(t, addr)
	conn.Close() // just confirming it's actually listening before we signal it

	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("failed to send SIGTERM: %v", err)
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected clean exit (status 0) after SIGTERM, got error: %v", err)
		}
	case <-time.After(5 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("process did not exit within 5s of receiving SIGTERM (this is the Issue 4 regression: no signal handling means SIGTERM behavior is left entirely to the Go runtime default, and a hang here means shutdown logic is blocking forever)")
	}
}

func TestSIGTERMDrainsInFlightClientBeforeExiting(t *testing.T) {
	binPath := mustBuildCacheDB(t)
	dataDir := t.TempDir()
	port := getFreePort(t)

	cmd := buildCacheDBCommand(binPath, port, dataDir, "no")
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

	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("failed to send SIGTERM: %v", err)
	}

	// The server should NOT exit immediately while our connection is still
	// open - give it a moment and confirm it's still alive, proving shutdown
	// waits for in-flight clients rather than force-killing them.
	time.Sleep(300 * time.Millisecond)
	if err := cmd.Process.Signal(syscall.Signal(0)); err != nil {
		t.Fatalf("expected server to still be draining with an open client connection, but it already exited: %v", err)
	}

	conn.Close()

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected clean exit once the client disconnected, got: %v", err)
		}
	case <-time.After(5 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("process did not exit after the draining client disconnected")
	}
}
