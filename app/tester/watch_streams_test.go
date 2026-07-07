package tester


import (
   "fmt"
   "testing"
)

func stream_xadd_watch_test(t *testing.T) {
   stage400_XAddInvalidatesWatch(t)
   stage401_XAddCreatesWatchedKey(t)
}

func stage400_XAddInvalidatesWatch(t *testing.T) {
   stage("STAGE 400: XADD INVALIDATES WATCH ON EXISTING STREAM")

   conn1, conn2 := dial(t), dial(t)
   defer conn1.Close()
   defer conn2.Close()

   key := "stream_watch_existing"
   
   // 1. Create the stream first so it exists
   send(conn1, fmt.Sprintf("*5\r\n$4\r\nXADD\r\n$%d\r\n%s\r\n$1\r\n*\r\n$4\r\ninit\r\n$1\r\n1\r\n", len(key), key))

   // 2. Client 1 watches the stream
   send(conn1, fmt.Sprintf("*2\r\n$5\r\nWATCH\r\n$%d\r\n%s\r\n", len(key), key))

   // 3. Client 2 appends another entry to the stream
   send(conn2, fmt.Sprintf("*5\r\n$4\r\nXADD\r\n$%d\r\n%s\r\n$1\r\n*\r\n$4\r\nname\r\n$4\r\nopen\r\n", len(key), key))

   // 4. Client 1 tries to execute a transaction
   send(conn1, "*1\r\n$5\r\nMULTI\r\n")
   send(conn1, "*1\r\n$4\r\nPING\r\n")
   resp := send(conn1, "*1\r\n$4\r\nEXEC\r\n")

   expected := "*-1\r\n" // Transaction must abort
   if resp != expected {
      failf(t, "expected %q got %q", expected, resp)
   }
   pass("XADD on existing stream invalidates WATCH")
}

func stage401_XAddCreatesWatchedKey(t *testing.T) {
   stage("STAGE 401: XADD CREATES WATCHED KEY")

   conn1, conn2 := dial(t), dial(t)
   defer conn1.Close()
   defer conn2.Close()

   key := "stream_missing_tracked"

   // 1. Watch a key that doesn't exist yet
   send(conn1, fmt.Sprintf("*2\r\n$5\r\nWATCH\r\n$%d\r\n%s\r\n", len(key), key))

   // 2. Client 2 creates it via XADD
   send(conn2, fmt.Sprintf("*5\r\n$4\r\nXADD\r\n$%d\r\n%s\r\n$1\r\n*\r\n$1\r\nx\r\n$1\r\ny\r\n", len(key), key))

   // 3. Client 1 tries to execute a transaction
   send(conn1, "*1\r\n$5\r\nMULTI\r\n")
   send(conn1, "*1\r\n$4\r\nPING\r\n")
   resp := send(conn1, "*1\r\n$4\r\nEXEC\r\n")

   expected := "*-1\r\n" // Transaction must abort
   if resp != expected {
      failf(t, "expected %q got %q", expected, resp)
   }
   pass("XADD creating a watched key aborts transaction")
}