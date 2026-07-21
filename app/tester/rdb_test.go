package tester

import (
	"os"
	"reflect"
	"testing"

	"CacheDB/app/RDB"
)

func TestReadRDB(t *testing.T) {
    filename := "test.rdb"

    err := os.WriteFile(filename, rdb.EmptyRDB, 0644)
    if err != nil {
        t.Fatal(err)
    }

    defer os.Remove(filename)

    config := rdb.RDB{
        Dir:         ".",
        DbFileName:  filename,
    }

    rdb.ReadRDBFile(&config)
}






func TestReadRDBFileWithLists(t *testing.T) {
	// Mock config pointing to directory/file
	rdbConfig := &rdb.RDB{
		Dir:        ".",
		DbFileName: "test_dump.rdb",
	}

	// Read entries
	entries, err := rdb.ReadRDBFile(rdbConfig)
	if err != nil {
		t.Fatalf("Failed to parse test RDB file: %v", err)
	}

	// Verify we parsed 3 items
	if len(entries) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(entries))
	}

	// Quick check mapping
	for _, entry := range entries {
		switch string(entry.Key) {
		case "name":
			if string(entry.Value.([]byte)) != "goat" {
				t.Errorf("Expected 'goat', got %s", entry.Value)
			}
		case "fruits":
			expected := [][]byte{[]byte("apple"), []byte("banana")}
			if !reflect.DeepEqual(entry.Value, expected) {
				t.Errorf("Fruits mismatch! Got %v", entry.Value)
			}
		case "colors":
			expected := [][]byte{[]byte("red"), []byte("green"), []byte("blue")}
			if !reflect.DeepEqual(entry.Value, expected) {
				t.Errorf("Colors mismatch! Got %v", entry.Value)
			}
		}
	}
}