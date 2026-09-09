package tester

// Regression test for Issue 7 (persistence/replication interaction).
//
// The master's PSYNC handler streams whatever RDB file is currently on disk
// (os.ReadFile(rdbConfig.Dir + "/" + rdbConfig.DbFileName)) rather than
// generating a fresh snapshot of current in-memory state. Since CacheDB only
// writes the RDB file at startup (if missing) or on an explicit SAVE, any
// write that happens on the master BEFORE a replica connects - but without an
// intervening SAVE - is invisible to that replica's full resync: the
// on-disk snapshot it receives predates the write, and the master only
// streams commands issued AFTER the replica's PSYNC completes.
//
// This test proves it: write a key to the master with no SAVE in between,
// then connect a replica and confirm the key is actually present.

import (
	"testing"
	"time"
)

func TestPSYNCStreamsCurrentState_NotStaleOnDiskSnapshot(t *testing.T) {
	masterCfg, masterPort := startMaster(t)

	// Write BEFORE any replica exists and with no SAVE in between - this is
	// exactly the case where "whatever is on disk" and "current in-memory
	// state" diverge.
	resp := sendRawCommand(t, masterPort, encodeCommand("SET", "pre-existing", "value"))
	if resp != "+OK\r\n" {
		t.Fatalf("expected +OK for SET, got %q", resp)
	}

	replicaCfg, replicaPort := startReplica(t, masterPort)
	_ = masterCfg
	_ = replicaCfg

	// Give the handshake + RDB transfer time to complete before we query.
	time.Sleep(300 * time.Millisecond)

	got := sendRawCommand(t, replicaPort, encodeCommand("GET", "pre-existing"))
	if got == "$-1\r\n" {
		t.Fatalf("replica's full resync did not include a write made before it connected " +
			"(this is the Issue 7 regression: PSYNC must snapshot current in-memory state, " +
			"not stream a stale on-disk RDB file)")
	}
	if got != "$5\r\nvalue\r\n" {
		t.Fatalf("expected replica to have 'value' for 'pre-existing', got %q", got)
	}
}
