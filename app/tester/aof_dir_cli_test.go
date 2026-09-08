package tester

// TestAOFDirCLIRespectsConfiguredDir is a higher-fidelity regression test for
// Issue 1 than a unit test can be: it builds and runs the actual `cachedb`
// binary the way a real user would, with --dir/--appendonly, and asserts that
// the resulting persistence files land where --dir says they should - not
// relative to the process's working directory.
//
// Before the fix, main() never assigned aof.AOF.Dir, so the AOF directory was
// always created relative to the process's CWD regardless of --dir, while the
// RDB file correctly honored --dir. This test builds the binary fresh, runs it
// from a CWD that is deliberately different from the configured --dir, and
// fails if any AOF artifacts appear under the CWD instead of under --dir.

import (
	"bufio"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"
)

func TestAOFDirCLIRespectsConfiguredDir(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine test file location")
	}
	// app/tester/aof_dir_cli_test.go -> app/
	appDir := filepath.Join(filepath.Dir(thisFile), "..")

	binPath := filepath.Join(t.TempDir(), "cachedb_under_test")
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = appDir
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("failed to build cachedb binary: %v\n%s", err, out)
	}

	dataDir := t.TempDir()
	cwd := t.TempDir() // deliberately NOT the same as dataDir

	port := getFreePort(t)

	cmd := exec.Command(binPath,
		"--port", strconv.Itoa(port),
		"--dir", dataDir,
		"--appendonly", "yes",
		"--appendfilename", "appendonly.aof",
	)
	cmd.Dir = cwd

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start cachedb: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	var conn net.Conn
	var err error
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		conn, err = net.Dial("tcp", addr)
		if err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("cachedb did not start listening on %s: %v", addr, err)
	}
	defer conn.Close()

	if _, err := conn.Write([]byte("*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n")); err != nil {
		t.Fatalf("failed to send SET: %v", err)
	}
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("failed to read SET response: %v", err)
	}
	if line != "+OK\r\n" {
		t.Fatalf("expected +OK, got %q", line)
	}

	time.Sleep(150 * time.Millisecond) // let the AOF write land on disk

	// The RDB file must exist under the configured --dir.
	if _, err := os.Stat(filepath.Join(dataDir, "rdbfile.db")); err != nil {
		t.Fatalf("expected RDB file under configured --dir %s: %v", dataDir, err)
	}

	// The AOF directory must ALSO exist under the configured --dir.
	aofDirUnderConfiguredDir := filepath.Join(dataDir, "appendonly")
	if _, err := os.Stat(aofDirUnderConfiguredDir); err != nil {
		t.Fatalf("expected AOF directory under configured --dir %s, got error: %v "+
			"(this is the Issue 1 regression: --dir must control the AOF directory too)",
			aofDirUnderConfiguredDir, err)
	}

	// Nothing persistence-related should have leaked into the process's CWD.
	if _, err := os.Stat(filepath.Join(cwd, "appendonly")); err == nil {
		t.Fatalf("found an AOF directory under the process CWD (%s) - --dir was not respected", cwd)
	}
	if _, err := os.Stat(filepath.Join(cwd, "rdbfile.db")); err == nil {
		t.Fatalf("found an RDB file under the process CWD (%s) - --dir was not respected", cwd)
	}
}
