package server

// Regression test for Phase 3 Priority 8.
//
// client.Dirty is written from another client's goroutine by markDirty
// (commands/utilities.go, called by every write command) under
// storage.WatchedKeysMutex. execCommand's read of client.Dirty, and
// clearWatches' write of client.Dirty = false, previously did not take that
// same lock - a real, reachable data race between a client running
// WATCH+MULTI+EXEC and another client concurrently writing the watched key.
// This test drives that exact scenario through the real execCommand/
// watchCommand/clearWatches code paths under `go test -race`.

import (
	"sync"
	"testing"

	"CacheDB/app/AOF"
	"CacheDB/app/commands"
	"CacheDB/app/config"
	"CacheDB/app/storage"
)

func newAuthenticatedTestUser() *storage.User {
	user := &storage.User{
		Name:               "default",
		Passwords:          make([][32]byte, 0),
		Flags:              storage.UserFlags{NoPass: true, Enabled: true},
		CommandPermissions: storage.AllCommands,
	}
	commands.GrantOrRevokePermission(user, []byte("@ADMIN"), true)
	return user
}

func TestWatchExec_DoesNotRaceWithConcurrentWriteToWatchedKey(t *testing.T) {
	storage.WatchedKeysMutex.Lock()
	storage.WatchedKeys = make(map[string]map[*storage.Client]struct{})
	storage.WatchedKeysMutex.Unlock()

	replConfig := &config.SERVER{
		Database: make(map[string]storage.Data),
	}
	aofConfig := &aof.AOF{}

	// High iteration count: the actual unsynchronized window per iteration
	// (a single field read/write) is extremely narrow, so this needs many
	// iterations to reliably force genuine goroutine interleaving.
	const iterations = 20000

	var wg sync.WaitGroup
	wg.Add(2)

	// "Client A": WATCH a key, queue a command, run EXEC repeatedly.
	go func() {
		defer wg.Done()
		clientA := &storage.Client{KeysWatched: make(map[string]struct{}), User: newAuthenticatedTestUser()}

		for i := 0; i < iterations; i++ {
			watchCommand([][]byte{[]byte("shared-key")}, clientA)
			clientA.InTransaction = true
			clientA.Queue = []storage.Command{{Args: [][]byte{[]byte("PING")}}}
			execCommand(nil, clientA, replConfig, aofConfig)
		}
	}()

	// "Client B": writes the same watched key concurrently, exactly what
	// SET does (dispatch through the real command handler so markDirty is
	// invoked exactly as it would be for a live client).
	go func() {
		defer wg.Done()
		clientB := &storage.Client{User: newAuthenticatedTestUser()}
		for i := 0; i < iterations; i++ {
			dispatchCommands(clientB, [][]byte{[]byte("SET"), []byte("shared-key"), []byte("v")}, replConfig, nil, aofConfig)
		}
	}()

	wg.Wait()
}
