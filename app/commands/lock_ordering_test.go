package commands

// Regression test for Phase 3 Priority 1/2 (lock-ordering audit).
//
// GetCommand's lazy-expiry path used to acquire ExpiryMutex.Lock() first,
// then (nested, while still holding it) DatabaseMutex.Lock() to delete the
// expired key's data. Elsewhere (RDB.SaveRDB/SaveRDBLocked, and the PSYNC
// handler's snapshotAndRegisterReplica in server.go) the SAME two mutexes
// are acquired in the OPPOSITE order: DatabaseMutex first, then
// ExpiryMutex. Two goroutines acquiring the same pair of locks in opposite
// orders is a textbook AB-BA deadlock: if goroutine A (GET on an expired
// key) holds ExpiryMutex and blocks waiting for DatabaseMutex, while
// goroutine B (a concurrent SAVE/PSYNC) holds DatabaseMutex and blocks
// waiting for ExpiryMutex, neither can ever proceed - the whole server
// hangs permanently.
//
// This test calls the REAL GetCommand (not a hand-copied reproduction of
// its old internal pattern) so it tracks the actual source going forward.
// Since we can't pause execution mid-function inside GetCommand itself
// without adding a test-only hook, the forced interleaving instead comes
// from goroutine B holding DatabaseMutex.RLock() for a fixed short window
// that comfortably covers the time GetCommand needs to reach its own
// internal lock acquisitions.

import (
	"testing"
	"time"

	"CacheDB/app/config"
	"CacheDB/app/storage"
)

func TestGetExpiredKey_DoesNotDeadlockWithDatabaseThenExpiryLockOrder(t *testing.T) {
	replconfig := &config.SERVER{
		Database: make(map[string]storage.Data),
		Expiry:   make(map[string]time.Time),
	}

	replconfig.Database["key"] = storage.Data{Type: storage.STRING, Value: []byte("v")}
	replconfig.Expiry["key"] = time.Now().Add(-time.Hour) // already expired

	bHoldsDatabase := make(chan struct{})
	getDone := make(chan struct{})
	bDone := make(chan struct{})

	// Goroutine A: a real client calling GET on an already-expired key -
	// exactly what triggers GetCommand's lazy-expiry deletion path.
	go func() {
		defer close(getDone)
		<-bHoldsDatabase // make sure B already holds DatabaseMutex.RLock() first
		GetCommand([][]byte{[]byte("key")}, replconfig)
	}()

	// Goroutine B: reproduces RDB.SaveRDB/SaveRDBLocked's exact order
	// (DatabaseMutex first, then ExpiryMutex), holding DatabaseMutex.RLock()
	// for a short fixed window so A's GetCommand call has time to reach its
	// own internal lock acquisitions while B is still holding Database.
	go func() {
		defer close(bDone)
		replconfig.DatabaseMutex.RLock()
		close(bHoldsDatabase)
		time.Sleep(200 * time.Millisecond)

		replconfig.ExpiryMutex.RLock() // mirrors SaveRDB's snapshot lock order
		replconfig.ExpiryMutex.RUnlock()
		replconfig.DatabaseMutex.RUnlock()
	}()

	select {
	case <-getDone:
	case <-time.After(3 * time.Second):
		t.Fatal("deadlock detected: GetCommand's ExpiryMutex->DatabaseMutex lock order and " +
			"SaveRDB's DatabaseMutex->ExpiryMutex lock order deadlock against each other when " +
			"they interleave (a live GET on an expired key racing a concurrent SAVE/PSYNC)")
	}

	select {
	case <-bDone:
	case <-time.After(3 * time.Second):
		t.Fatal("goroutine B never completed - likely deadlocked waiting on ExpiryMutex")
	}
}
