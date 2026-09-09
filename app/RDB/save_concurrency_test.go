package rdb

// Regression test for Phase 3 Priority 1: SaveRDB (via WriteReSizeDB and
// writeDatabase) ranged over replconfig.Database/Expiry with NO lock at all,
// while every live command (SET/DEL/INCR/LPUSH/...) mutates those same maps
// under DatabaseMutex/ExpiryMutex. Concurrent unsynchronized map iteration
// alongside a map write is not just a data race - Go's runtime can raise an
// unrecoverable "fatal error: concurrent map iteration and map write" that
// crashes the whole process, not just the goroutine. This test drives SAVE
// concurrently with writes under `go test -race` to prove it's actually
// reachable from real server code paths, not merely theoretical.

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"CacheDB/app/config"
	"CacheDB/app/storage"
)

func newTestReplConfig() *config.SERVER {
	return &config.SERVER{
		Database: make(map[string]storage.Data),
		Expiry:   make(map[string]time.Time),
	}
}

func TestSaveRDB_DoesNotRaceWithConcurrentWrites(t *testing.T) {
	replconfig := newTestReplConfig()

	for i := 0; i < 50; i++ {
		replconfig.Database["key"+strconv.Itoa(i)] = storage.Data{
			Type:  storage.STRING,
			Value: []byte("value"),
		}
	}

	dir := t.TempDir()
	path := dir + "/dump.rdb"

	var wg sync.WaitGroup
	stop := make(chan struct{})

	// Writer goroutines mutating Database/Expiry exactly the way live
	// commands do (DatabaseMutex/ExpiryMutex around the map access).
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			n := 0
			for {
				select {
				case <-stop:
					return
				default:
				}
				key := "writer" + strconv.Itoa(id) + "-" + strconv.Itoa(n)

				replconfig.DatabaseMutex.Lock()
				replconfig.Database[key] = storage.Data{Type: storage.STRING, Value: []byte("v")}
				replconfig.DatabaseMutex.Unlock()

				replconfig.ExpiryMutex.Lock()
				replconfig.Expiry[key] = time.Now().Add(time.Hour)
				replconfig.ExpiryMutex.Unlock()

				replconfig.DatabaseMutex.Lock()
				delete(replconfig.Database, key)
				replconfig.DatabaseMutex.Unlock()

				n++
			}
		}(i)
	}

	// Repeatedly SAVE while the writers above are mutating the maps.
	for i := 0; i < 20; i++ {
		if err := SaveRDB(path, replconfig); err != nil {
			t.Fatalf("SaveRDB failed: %v", err)
		}
	}

	close(stop)
	wg.Wait()
}
