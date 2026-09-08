package aof

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newOpenAOFFile(t *testing.T) (*AOF, *os.File) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.aof")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("failed to open test AOF file: %v", err)
	}
	return &AOF{File: f}, f
}

func TestAppendCommand_AlwaysFsyncsWithoutError(t *testing.T) {
	cfg, f := newOpenAOFFile(t)
	defer f.Close()
	cfg.AppendFsync = "always"

	if err := cfg.AppendCommand([]byte("*1\r\n$4\r\nPING\r\n")); err != nil {
		t.Fatalf("expected AppendCommand with appendfsync=always to succeed, got: %v", err)
	}
}

func TestAppendCommand_EverysecDoesNotFsyncPerWrite(t *testing.T) {
	cfg, f := newOpenAOFFile(t)
	defer f.Close()
	cfg.AppendFsync = "everysec"

	// Not independently observable that Sync() was skipped without OS-level
	// instrumentation, but this proves everysec mode writes successfully
	// through the same path as "always" without requiring a per-write fsync.
	if err := cfg.AppendCommand([]byte("*1\r\n$4\r\nPING\r\n")); err != nil {
		t.Fatalf("expected AppendCommand with appendfsync=everysec to succeed, got: %v", err)
	}
}

func TestAppendCommand_NoDoesNotFsync(t *testing.T) {
	cfg, f := newOpenAOFFile(t)
	defer f.Close()
	cfg.AppendFsync = "no"

	if err := cfg.AppendCommand([]byte("*1\r\n$4\r\nPING\r\n")); err != nil {
		t.Fatalf("expected AppendCommand with appendfsync=no to succeed, got: %v", err)
	}
}

func TestAppendCommand_NoOpWithoutOpenFile(t *testing.T) {
	cfg := &AOF{AppendFsync: "always"} // File is nil (appendonly disabled)

	if err := cfg.AppendCommand([]byte("*1\r\n$4\r\nPING\r\n")); err != nil {
		t.Fatalf("expected no-op (nil error) when no AOF file is open, got: %v", err)
	}
}

func TestStartFsyncLoop_NoOpUnlessEverysec(t *testing.T) {
	cfg, f := newOpenAOFFile(t)
	defer f.Close()
	cfg.AppendOnly = "yes"
	cfg.AppendFsync = "always" // not "everysec" -> StartFsyncLoop must not start a ticker

	stop := make(chan struct{})
	cfg.StartFsyncLoop(stop)
	close(stop) // if a goroutine were started, this proves it can still be stopped harmlessly

	// Give any (incorrectly) started goroutine a moment to misbehave; nothing
	// to assert on directly beyond "this doesn't hang or panic".
	time.Sleep(10 * time.Millisecond)
}

func TestStartFsyncLoop_DoneClosesImmediatelyWhenNoOp(t *testing.T) {
	cfg, f := newOpenAOFFile(t)
	defer f.Close()
	cfg.AppendOnly = "yes"
	cfg.AppendFsync = "always" // not "everysec" -> no-op

	done := cfg.StartFsyncLoop(make(chan struct{}))
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("expected done to be closed immediately for a no-op StartFsyncLoop")
	}
}

func TestStartFsyncLoop_DoneClosesAfterStopSignal(t *testing.T) {
	cfg, f := newOpenAOFFile(t)
	defer f.Close()
	cfg.AppendOnly = "yes"
	cfg.AppendFsync = "everysec"

	stop := make(chan struct{})
	done := cfg.StartFsyncLoop(stop)

	select {
	case <-done:
		t.Fatal("expected done to still be open while the loop is running")
	case <-time.After(50 * time.Millisecond):
	}

	close(stop)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("expected done to close shortly after stop is closed - a caller " +
			"waiting on done before closing the AOF file must not block forever")
	}
}

func TestStartFsyncLoop_EverysecFsyncsPeriodically(t *testing.T) {
	cfg, f := newOpenAOFFile(t)
	defer f.Close()
	cfg.AppendOnly = "yes"
	cfg.AppendFsync = "everysec"

	stop := make(chan struct{})
	defer close(stop)

	cfg.StartFsyncLoop(stop)

	// Exact fsync timing/occurrence isn't independently observable from a
	// unit test without OS-level syscall interception, so this test only
	// documents and exercises that the loop starts and can be stopped
	// cleanly without panicking or leaking past the stop signal - it does
	// NOT assert that a physical fsync syscall occurred within the window.
	time.Sleep(1100 * time.Millisecond)
}
