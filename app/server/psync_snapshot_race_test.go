package server

// Regression test for Phase 3 Priority 7 (the "critical" snapshot/streaming
// boundary race).
//
// Before this fix, the PSYNC handler did: SaveRDB() [acquiring and releasing
// DatabaseMutex/ExpiryMutex internally] -> read file -> send to replica ->
// THEN separately lock ReplicasMutex to append the replica. Between "SaveRDB
// releases its locks" and "replica appended to REPLICAS", a live write could
// complete (mutate Database) and reach PropagateCommands - which only checks
// REPLICAS under ReplicasMutex. Since the new replica wasn't registered yet,
// PropagateCommands would never send it that write, AND the write happened
// AFTER the snapshot was already taken, so it wouldn't be in the RDB either.
// The write is lost for that replica, permanently and silently.
//
// The fix holds DatabaseMutex+ExpiryMutex+ReplicasMutex together across BOTH
// the snapshot AND the replica-registration append (snapshotAndRegisterReplica),
// unlocking ReplicasMutex, then ExpiryMutex, then DatabaseMutex last (LIFO
// defers), with the REPLICAS append happening before any of the three
// unlocks. That gives a deterministic guarantee via Go's mutex happens-before
// semantics: by the time ANY goroutine successfully acquires DatabaseMutex
// again, the REPLICAS append must have already happened (append -> Replicas-
// Mutex.Unlock -> [program order, same goroutine] -> DatabaseMutex.Unlock ->
// [mutex happens-before] -> the next Lock() succeeding). This test checks
// exactly that invariant - len(REPLICAS) as observed by a concurrent writer
// the instant it acquires DatabaseMutex - rather than relying on a separate
// "done" channel's close ordering, which would itself race independently of
// the underlying mutex behavior.

import (
	"net"
	"testing"
	"time"

	"CacheDB/app/config"
	"CacheDB/app/storage"
)

func TestSnapshotAndRegisterReplica_RegistersBeforeReleasingDatabaseLock(t *testing.T) {
	replConfig := &config.SERVER{}
	replConfig.Database = make(map[string]storage.Data)
	replConfig.Expiry = make(map[string]time.Time)
	replConfig.REPLICAS = make([]*config.REPLICA, 0)

	rdbConfig := newTestRDBConfig(t)

	// A big-ish database makes the snapshot's map iteration take long enough
	// that the writer goroutine below reliably attempts its Lock() call
	// WHILE snapshotAndRegisterReplica is still running, rather than both
	// goroutines finishing before either can overlap with the other.
	for i := 0; i < 20000; i++ {
		key := "key" + time.Now().Format("150405.000000000") + string(rune(i))
		replConfig.Database[key] = storage.Data{Type: storage.STRING, Value: []byte("value")}
	}

	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	snapshotStarted := make(chan struct{})
	snapshotDone := make(chan struct{})
	go func() {
		close(snapshotStarted)
		defer close(snapshotDone)
		if err := snapshotAndRegisterReplica(rdbConfig, replConfig, serverConn); err != nil {
			t.Errorf("snapshotAndRegisterReplica failed: %v", err)
		}
	}()
	<-snapshotStarted

	replicaCountWhenWriterGotLock := -1
	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		// Exactly what a live write command does: Lock, mutate, Unlock.
		replConfig.DatabaseMutex.Lock()
		replConfig.ReplicasMutex.RLock()
		replicaCountWhenWriterGotLock = len(replConfig.REPLICAS)
		replConfig.ReplicasMutex.RUnlock()
		replConfig.Database["concurrent-write-key"] = storage.Data{Type: storage.STRING, Value: []byte("v")}
		replConfig.DatabaseMutex.Unlock()
	}()

	<-snapshotDone
	<-writerDone

	if replicaCountWhenWriterGotLock != 1 {
		t.Fatalf("expected the replica to already be registered (count=1) by the time a concurrent "+
			"write acquired DatabaseMutex, got count=%d - this means a write could land in the gap "+
			"between snapshot and replica registration and be silently lost to that replica "+
			"(the Priority 7 regression)", replicaCountWhenWriterGotLock)
	}

	replConfig.ReplicasMutex.RLock()
	finalCount := len(replConfig.REPLICAS)
	replConfig.ReplicasMutex.RUnlock()
	if finalCount != 1 {
		t.Fatalf("expected exactly 1 registered replica after snapshotAndRegisterReplica, got %d", finalCount)
	}
}
