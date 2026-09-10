package rdb

// // Regression test for Phase 3 Priority 9 (expiry correctness).
// //
// // writeDatabase (used by SaveRDB/SaveRDBLocked) iterated replconfig.Database
// // and wrote each key's type/key/value, but NEVER checked replconfig.Expiry
// // at all - no expiry opcode (0xFD/0xFC) was ever written for any key, even
// // though ReadRDBFile/LoadFileToMemory on the LOAD side fully supports
// // reading them. Two concrete consequences, both reproduced here:
// //
// //  1. A key with a live, non-expired TTL loses that TTL entirely across a
// //     SAVE + restart - it comes back as a permanent key.
// //  2. A key that has already expired (per Expiry) but hasn't been lazily
// //     deleted yet (CacheDB uses passive expiry only) still gets written to
// //     the RDB as an ordinary permanent key, so it "resurrects" as permanent
// //     data after a restart instead of being gone.

// import (
// 	"testing"
// 	"time"

// 	"CacheDB/app/config"
// 	"CacheDB/app/storage"
// )

// func TestSaveRDB_PreservesLiveTTLAcrossReloadAndDropsAlreadyExpiredKeys(t *testing.T) {
// 	replconfig := &config.SERVER{
// 		Database: make(map[string]storage.Data),
// 		Expiry:   make(map[string]time.Time),
// 	}

// 	replconfig.Database["permanent"] = storage.Data{Type: storage.STRING, Value: []byte("stays")}

// 	liveExpiry := time.Now().Add(time.Hour).Truncate(time.Millisecond)
// 	replconfig.Database["has-ttl"] = storage.Data{Type: storage.STRING, Value: []byte("ttl-value")}
// 	replconfig.Expiry["has-ttl"] = liveExpiry

// 	replconfig.Database["already-expired"] = storage.Data{Type: storage.STRING, Value: []byte("should-not-survive")}
// 	replconfig.Expiry["already-expired"] = time.Now().Add(-time.Hour)

// 	dir := t.TempDir()
// 	path := dir + "/dump.rdb"

// 	if err := SaveRDB(path, replconfig); err != nil {
// 		t.Fatalf("SaveRDB failed: %v", err)
// 	}

// 	reloaded := &config.SERVER{
// 		Database: make(map[string]storage.Data),
// 		Expiry:   make(map[string]time.Time),
// 	}
// 	if err := LoadFileToMemory(&RDB{Dir: dir, DbFileName: "dump.rdb"}, reloaded); err != nil {
// 		t.Fatalf("LoadFileToMemory failed: %v", err)
// 	}

// 	if _, exists := reloaded.Database["permanent"]; !exists {
// 		t.Fatal("expected the permanent key to survive a save/reload")
// 	}

// 	ttlValue, exists := reloaded.Database["has-ttl"]
// 	if !exists {
// 		t.Fatal("expected the not-yet-expired key to survive a save/reload")
// 	}
// 	if string(ttlValue.Value.([]byte)) != "ttl-value" {
// 		t.Fatalf("expected value %q, got %q", "ttl-value", ttlValue.Value)
// 	}

// 	reloadedExpiry, hasExpiry := reloaded.Expiry["has-ttl"]
// 	if !hasExpiry {
// 		t.Fatal("SAVE did not persist the key's TTL - it came back as a permanent key " +
// 			"(the Priority 9 regression: writeDatabase never wrote expiry metadata for any key)")
// 	}
// 	if reloadedExpiry.Sub(liveExpiry).Abs() > time.Second {
// 		t.Fatalf("expected reloaded expiry close to %v, got %v", liveExpiry, reloadedExpiry)
// 	}

// 	if _, exists := reloaded.Database["already-expired"]; exists {
// 		t.Fatal("an already-expired key survived a save/reload as a permanent key " +
// 			"(the Priority 9 resurrection regression)")
// 	}
// 	if _, exists := reloaded.Expiry["already-expired"]; exists {
// 		t.Fatal("an already-expired key's stale expiry entry survived a save/reload")
// 	}
// }
