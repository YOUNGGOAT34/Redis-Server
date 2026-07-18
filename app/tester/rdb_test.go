package tester


import (
    "os"
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

    rdb.ReadRDBFile(config)
}