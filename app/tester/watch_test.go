package tester

import "testing"


func watch_test(t *testing.T) {

   //  stage200_WatchBasic(t)
   //  stage201_WatchWrongArguments(t)
   //  stage202_WatchMultipleKeys(t)
   //  stage203_WatchDuplicateKeys(t)

   //  stage204_WatchExecSuccess(t)
   //  stage205_WatchAbortOnExternalWrite(t)
   //  stage206_WatchAbortMultipleKeys(t)

   //  stage207_SameClientWriteAllowed(t)
   //  stage208_WatchClearedAfterExec(t)
   //  stage209_WatchClearedAfterDiscard(t)

   //  stage210_WatchIsolation(t)

   //  stage211_WatchInsideMultiRejected(t)

   //  stage212_MultipleClientsWatchingSameKey(t)
   //  stage213_OnlyOneWatcherInvalidated(t)

   //  stage214_WatchThenDeleteKey(t)
   //  stage215_WatchNewKeyThenCreateIt(t)

   //  stage216_ReWatchExistingKey(t)

   //  stage217_WatchReadOnlyCommands(t)

   //  stage218_ExecWithoutModification(t)

   //  stage219_AnyWatchedKeyInvalidates(t)

    stage220_UnwatchBasic(t)
    stage221_UnwatchWrongArguments(t)
    stage222_UnwatchWithoutWatch(t)
   
}


func stage200_WatchBasic(t *testing.T) {
	stage("STAGE 200: WATCH BASIC")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$1\r\nx\r\n")

	expected := "+OK\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("WATCH accepted")
}

func stage201_WatchWrongArguments(t *testing.T) {
	stage("STAGE 201: WATCH WRONG ARGUMENTS")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn,
		"*1\r\n"+
			"$5\r\nWATCH\r\n")

	expected := "-Wrong number of arguments for 'WATCH' command\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("WATCH validates argument count")
}

func stage202_WatchMultipleKeys(t *testing.T) {
	stage("STAGE 202: WATCH MULTIPLE KEYS")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn,
		"*4\r\n"+
			"$5\r\nWATCH\r\n"+
			"$1\r\na\r\n"+
			"$1\r\nb\r\n"+
			"$1\r\nc\r\n")

	expected := "+OK\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("multiple keys watched")
}

func stage203_WatchDuplicateKeys(t *testing.T) {
	stage("STAGE 203: WATCH DUPLICATE KEYS")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn,
		"*4\r\n"+
			"$5\r\nWATCH\r\n"+
			"$1\r\nx\r\n"+
			"$1\r\nx\r\n"+
			"$1\r\nx\r\n")

	expected := "+OK\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("duplicate WATCH succeeds")
}


func stage204_WatchExecSuccess(t *testing.T) {
	stage("STAGE 204: WATCH EXEC SUCCESS")

	conn := dial(t)
	defer conn.Close()

	send(conn,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$1\r\nx\r\n")

	send(conn,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn,
		"*3\r\n"+
			"$3\r\nSET\r\n"+
			"$1\r\nx\r\n"+
			"$2\r\n42\r\n")

	resp := send(conn,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected :=
		"*1\r\n" +
			"+OK\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("transaction succeeds when watched keys unchanged")
}


func stage205_WatchAbortOnExternalWrite(t *testing.T) {
	stage("STAGE 205: WATCH ABORT ON EXTERNAL WRITE")

	conn1 := dial(t)
	conn2 := dial(t)

	defer conn1.Close()
	defer conn2.Close()

	send(conn1,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$1\r\nx\r\n")

	send(conn2,
		"*3\r\n"+
			"$3\r\nSET\r\n"+
			"$1\r\nx\r\n"+
			"$1\r\n9\r\n")

	send(conn1,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn1,
		"*2\r\n"+
			"$3\r\nGET\r\n"+
			"$1\r\nx\r\n")

	resp := send(conn1,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected := "*-1\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("transaction aborted after watched key modified")
}


func stage206_WatchAbortMultipleKeys(t *testing.T) {
	stage("STAGE 206: WATCH MULTIPLE KEYS INVALIDATION")

	conn1 := dial(t)
	conn2 := dial(t)

	defer conn1.Close()
	defer conn2.Close()

	send(conn1,
		"*4\r\n"+
			"$5\r\nWATCH\r\n"+
			"$1\r\na\r\n"+
			"$1\r\nb\r\n"+
			"$1\r\nc\r\n")

	send(conn2,
		"*3\r\n"+
			"$3\r\nSET\r\n"+
			"$1\r\nb\r\n"+
			"$2\r\n99\r\n")

	send(conn1,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn1,
		"*2\r\n"+
			"$3\r\nGET\r\n"+
			"$1\r\na\r\n")

	resp := send(conn1,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected := "*-1\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("any watched key invalidates transaction")
}

func stage207_SameClientWriteAllowed(t *testing.T) {
	stage("STAGE 207: SAME CLIENT WRITE DOES NOT INVALIDATE")

	conn := dial(t)
	defer conn.Close()

	send(conn,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$1\r\nx\r\n")

	send(conn,
		"*3\r\n"+
			"$3\r\nSET\r\n"+
			"$1\r\nx\r\n"+
			"$2\r\n55\r\n")

	send(conn,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn,
		"*2\r\n"+
			"$3\r\nGET\r\n"+
			"$1\r\nx\r\n")

	resp := send(conn,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected :=
		"*1\r\n" +
			"$2\r\n55\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("same client writes do not invalidate WATCH")
}


func stage208_WatchClearedAfterExec(t *testing.T) {
	stage("STAGE 208: WATCH CLEARED AFTER EXEC")

	conn1 := dial(t)
	conn2 := dial(t)

	defer conn1.Close()
	defer conn2.Close()

	send(conn1,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$1\r\nx\r\n")

	send(conn1,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn1,
		"*1\r\n"+
			"$4\r\nPING\r\n")

	resp := send(conn1,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected := "*1\r\n+PONG\r\n"

	if resp != expected {
		
		failf(t, "expected %q got %q", expected, resp)
	}

	// // WATCH should now be cleared.
	send(conn2,
		"*3\r\n"+
			"$3\r\nSET\r\n"+
			"$1\r\nx\r\n"+
			"$1\r\n9\r\n")

	send(conn1,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn1,
		"*1\r\n"+
			"$4\r\nPING\r\n")

	resp = send(conn1,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected = "*1\r\n+PONG\r\n"

	if resp != expected {
		failf(t, "WATCH wasn't cleared after EXEC")
	}

	pass("EXEC clears WATCH state")
}


func stage209_WatchClearedAfterDiscard(t *testing.T) {
	stage("STAGE 209: WATCH CLEARED AFTER DISCARD")

	conn1 := dial(t)
	conn2 := dial(t)

	defer conn1.Close()
	defer conn2.Close()

	send(conn1,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$1\r\nx\r\n")

	send(conn1,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn1,
		"*1\r\n"+
			"$7\r\nDISCARD\r\n")

	send(conn2,
		"*3\r\n"+
			"$3\r\nSET\r\n"+
			"$1\r\nx\r\n"+
			"$1\r\n8\r\n")

	send(conn1,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn1,
		"*1\r\n"+
			"$4\r\nPING\r\n")

	resp := send(conn1,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected := "*1\r\n+PONG\r\n"

	if resp != expected {
		failf(t, "WATCH should be cleared after DISCARD")
	}

	pass("DISCARD clears WATCH")
}



func stage210_WatchIsolation(t *testing.T) {
	stage("STAGE 210: WATCH CLIENT ISOLATION")

	conn1 := dial(t)
	conn2 := dial(t)
	conn3 := dial(t)

	defer conn1.Close()
	defer conn2.Close()
	defer conn3.Close()

	send(conn1,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$1\r\na\r\n")

	send(conn2,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$1\r\nb\r\n")

	send(conn3,
		"*3\r\n"+
			"$3\r\nSET\r\n"+
			"$1\r\na\r\n"+
			"$1\r\n1\r\n")

	send(conn2,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn2,
		"*1\r\n"+
			"$4\r\nPING\r\n")

	resp := send(conn2,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected := "*1\r\n+PONG\r\n"

	if resp != expected {
		failf(t, "client2 should not be affected")
	}

	send(conn1,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn1,
		"*1\r\n"+
			"$4\r\nPING\r\n")

	resp = send(conn1,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected = "*-1\r\n"

	if resp != expected {
		failf(t, "client1 should abort")
	}

	pass("WATCH state isolated per client")
}


func stage211_WatchInsideMultiRejected(t *testing.T) {
	stage("STAGE 211: WATCH INSIDE MULTI")

	conn := dial(t)
	defer conn.Close()

	send(conn,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	resp := send(conn,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$1\r\nx\r\n")

	expected := "-ERR WATCH inside MULTI is not allowed\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("WATCH rejected inside MULTI")
}


func stage212_MultipleClientsWatchingSameKey(t *testing.T) {
	stage("STAGE 212: MULTIPLE CLIENTS WATCH SAME KEY")

	conn1 := dial(t)
	conn2 := dial(t)
	conn3 := dial(t)

	defer conn1.Close()
	defer conn2.Close()
	defer conn3.Close()

	send(conn1,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$1\r\nx\r\n")

	send(conn2,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$1\r\nx\r\n")

	send(conn3,
		"*3\r\n"+
			"$3\r\nSET\r\n"+
			"$1\r\nx\r\n"+
			"$1\r\n9\r\n")

	send(conn1,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn1,
		"*1\r\n"+
			"$4\r\nPING\r\n")

	send(conn2,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn2,
		"*1\r\n"+
			"$4\r\nPING\r\n")

	resp1 := send(conn1,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	resp2 := send(conn2,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected := "*-1\r\n"

	if resp1 != expected {
		failf(t, "client1 expected abort got %q", resp1)
	}

	if resp2 != expected {
		failf(t, "client2 expected abort got %q", resp2)
	}

	pass("all watchers invalidated")
}


func stage213_OnlyOneWatcherInvalidated(t *testing.T) {
	stage("STAGE 213: ONLY CORRECT WATCHER INVALIDATED")

	conn1 := dial(t)
	conn2 := dial(t)
	conn3 := dial(t)

	defer conn1.Close()
	defer conn2.Close()
	defer conn3.Close()

	send(conn1,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$1\r\na\r\n")

	send(conn2,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$1\r\nb\r\n")

	send(conn3,
		"*3\r\n"+
			"$3\r\nSET\r\n"+
			"$1\r\na\r\n"+
			"$1\r\n5\r\n")

	send(conn1,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn1,
		"*1\r\n"+
			"$4\r\nPING\r\n")

	send(conn2,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn2,
		"*1\r\n"+
			"$4\r\nPING\r\n")

	resp1 := send(conn1,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	resp2 := send(conn2,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	if resp1 != "*-1\r\n" {
		failf(t, "client1 expected abort")
	}

	if resp2 != "*1\r\n+PONG\r\n" {
		failf(t, "client2 should succeed")
	}

	pass("only affected watcher aborts")
}


// func stage214_WatchThenDeleteKey(t *testing.T) {
// 	stage("STAGE 214: DELETE INVALIDATES WATCH")

// 	conn1 := dial(t)
// 	conn2 := dial(t)

// 	defer conn1.Close()
// 	defer conn2.Close()

// 	send(conn1,
// 		"*2\r\n"+
// 			"$5\r\nWATCH\r\n"+
// 			"$1\r\nx\r\n")

// 	send(conn2,
// 		"*2\r\n"+
// 			"$3\r\nDEL\r\n"+
// 			"$1\r\nx\r\n")

// 	send(conn1,
// 		"*1\r\n"+
// 			"$5\r\nMULTI\r\n")

// 	send(conn1,
// 		"*1\r\n"+
// 			"$4\r\nPING\r\n")

// 	resp := send(conn1,
// 		"*1\r\n"+
// 			"$4\r\nEXEC\r\n")

// 	expected := "*-1\r\n"

// 	if resp != expected {
// 		failf(t, "expected abort got %q", resp)
// 	}

// 	pass("DEL invalidates WATCH")
// }



func stage215_WatchNewKeyThenCreateIt(t *testing.T) {
	stage("STAGE 215: CREATE WATCHED KEY")

	conn1 := dial(t)
	conn2 := dial(t)

	defer conn1.Close()
	defer conn2.Close()

	send(conn1,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$7\r\nnew_key\r\n")

	send(conn2,
		"*3\r\n"+
			"$3\r\nSET\r\n"+
			"$7\r\nnew_key\r\n"+
			"$5\r\nhello\r\n")

	send(conn1,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn1,
		"*1\r\n"+
			"$4\r\nPING\r\n")

	resp := send(conn1,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected := "*-1\r\n"

	if resp != expected {
		failf(t, "expected abort got %q", resp)
	}

	pass("creating watched key invalidates transaction")
}




func stage216_ReWatchExistingKey(t *testing.T) {
	stage("STAGE 216: REPEATED WATCH")

	conn1 := dial(t)
	conn2 := dial(t)

	defer conn1.Close()
	defer conn2.Close()

	send(conn1,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$1\r\nx\r\n")

	resp := send(conn1,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$1\r\nx\r\n")

	expected := "+OK\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	send(conn2,
		"*3\r\n"+
			"$3\r\nSET\r\n"+
			"$1\r\nx\r\n"+
			"$1\r\n7\r\n")

	send(conn1,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn1,
		"*1\r\n"+
			"$4\r\nPING\r\n")

	resp = send(conn1,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected = "*-1\r\n"

	if resp != expected {
		failf(t, "expected transaction abort got %q", resp)
	}

	pass("repeated WATCH behaves correctly")
}


func stage217_WatchReadOnlyCommands(t *testing.T) {
	stage("STAGE 217: READ COMMANDS DO NOT INVALIDATE")

	conn1 := dial(t)
	conn2 := dial(t)

	defer conn1.Close()
	defer conn2.Close()

	send(conn1,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$1\r\nx\r\n")

	send(conn2,
		"*2\r\n"+
			"$3\r\nGET\r\n"+
			"$1\r\nx\r\n")

	send(conn1,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn1,
		"*1\r\n"+
			"$4\r\nPING\r\n")

	resp := send(conn1,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected := "*1\r\n+PONG\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("read-only commands don't invalidate WATCH")
}


func stage218_ExecWithoutModification(t *testing.T) {
	stage("STAGE 218: EXEC WITHOUT MODIFICATION")

	conn := dial(t)
	defer conn.Close()

	send(conn,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$1\r\nk\r\n")

	send(conn,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn,
		"*1\r\n"+
			"$4\r\nPING\r\n")

	resp := send(conn,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected := "*1\r\n+PONG\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("transaction succeeds when watched keys unchanged")
}


func stage219_AnyWatchedKeyInvalidates(t *testing.T) {
	stage("STAGE 219: ANY WATCHED KEY INVALIDATES")

	conn1 := dial(t)
	conn2 := dial(t)

	defer conn1.Close()
	defer conn2.Close()

	send(conn1,
		"*4\r\n"+
			"$5\r\nWATCH\r\n"+
			"$1\r\na\r\n"+
			"$1\r\nb\r\n"+
			"$1\r\nc\r\n")

	send(conn2,
		"*3\r\n"+
			"$3\r\nSET\r\n"+
			"$1\r\nc\r\n"+
			"$1\r\n1\r\n")

	send(conn1,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn1,
		"*1\r\n"+
			"$4\r\nPING\r\n")

	resp := send(conn1,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected := "*-1\r\n"

	if resp != expected {
		failf(t, "expected abort got %q", resp)
	}

	pass("single modified watched key aborts transaction")
}




func stage220_UnwatchBasic(t *testing.T) {
	stage("STAGE 220: UNWATCH BASIC")

	conn := dial(t)
	defer conn.Close()

	send(conn,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$1\r\nx\r\n")

	resp := send(conn,
		"*1\r\n"+
			"$7\r\nUNWATCH\r\n")

	expected := "+OK\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("UNWATCH clears watched keys")
}

func stage221_UnwatchWrongArguments(t *testing.T) {
	stage("STAGE 221: UNWATCH WRONG ARGUMENTS")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn,
		"*2\r\n"+
			"$7\r\nUNWATCH\r\n"+
			"$1\r\nx\r\n")

	expected := "-Wrong number of arguments for 'UNWATCH' command\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("UNWATCH validates argument count")
}

func stage222_UnwatchWithoutWatch(t *testing.T) {
	stage("STAGE 222: UNWATCH WITHOUT WATCH")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn,
		"*1\r\n"+
			"$7\r\nUNWATCH\r\n")

	expected := "+OK\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("UNWATCH succeeds even with no watched keys")
}

func stage223_UnwatchPreventsAbort(t *testing.T) {
	stage("STAGE 223: UNWATCH CLEARS WATCHES")

	conn1 := dial(t)
	conn2 := dial(t)

	defer conn1.Close()
	defer conn2.Close()

	send(conn1,
		"*2\r\n"+
			"$5\r\nWATCH\r\n"+
			"$1\r\nx\r\n")

	send(conn1,
		"*1\r\n"+
			"$7\r\nUNWATCH\r\n")

	send(conn2,
		"*3\r\n"+
			"$3\r\nSET\r\n"+
			"$1\r\nx\r\n"+
			"$1\r\n5\r\n")

	send(conn1,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn1,
		"*1\r\n"+
			"$4\r\nPING\r\n")

	resp := send(conn1,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected := "*1\r\n+PONG\r\n"

	if resp != expected {
		failf(t, "UNWATCH failed to clear watched keys")
	}

	pass("UNWATCH removes watch state")
}