package tester

import "testing"

func transaction_test(t *testing.T) {
	
	stage100_ExecWithoutMulti(t)
	stage101_MultiQueuesCommands(t)
	stage102_ExecRunsQueuedCommands(t)
	stage103_ExecEmptyQueue(t)
	stage106_ExecWrongArguments(t)
	stage108_DiscardBasic(t)
	stage109_DiscardWithoutMulti(t)
	stage110_DiscardClearsQueue(t)
	stage111_DiscardWrongArguments(t) 
	stage112_ExecAfterDiscard(t) 
	stage113_TransactionRuntimeError(t)
	stage114_TransactionMultipleErrors(t)
	stage115_TransactionContinuesAfterFailure(t) 
}



func stage100_ExecWithoutMulti(t *testing.T) {
	stage("STAGE 100: EXEC WITHOUT MULTI")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected := "-ERR EXEC without MULTI\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("EXEC without MULTI rejected")
}


func stage101_MultiQueuesCommands(t *testing.T) {
	stage("STAGE 101: MULTI QUEUES COMMANDS")

	conn := dial(t)
	defer conn.Close()

	send(conn,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	resp := send(conn,
		"*3\r\n"+
			"$3\r\nSET\r\n"+
			"$1\r\nx\r\n"+
			"$1\r\n1\r\n")

	expected := "+QUEUED\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("commands are queued")
}

func stage102_ExecRunsQueuedCommands(t *testing.T) {
	stage("STAGE 102: EXEC RUNS QUEUED COMMANDS")

	conn := dial(t)
	defer conn.Close()

	send(conn,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn,
		"*3\r\n"+
			"$3\r\nSET\r\n"+
			"$1\r\nx\r\n"+
			"$2\r\n42\r\n")

	send(conn,
		"*2\r\n"+
			"$4\r\nINCR\r\n"+
			"$1\r\nx\r\n")

	resp := send(conn,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected :=
		"*2\r\n" +
			"+OK\r\n" +
			":43\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("EXEC executes queued commands")
}


func stage103_ExecEmptyQueue(t *testing.T) {
	stage("STAGE 103: EXEC EMPTY QUEUE")

	conn := dial(t)
	defer conn.Close()

	send(conn,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	resp := send(conn,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected := "*0\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("empty transaction returns empty array")
}


func stage106_ExecWrongArguments(t *testing.T) {
	stage("STAGE 106: EXEC WRONG ARGUMENTS")

	conn := dial(t)
	defer conn.Close()

	send(conn,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	resp := send(conn,
		"*2\r\n"+
			"$4\r\nEXEC\r\n"+
			"$1\r\nx\r\n")

	expected := "-Wrong number of arguments for 'EXEC' command\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("EXEC validates argument count")
}


func stage108_DiscardBasic(t *testing.T) {
	stage("STAGE 108: DISCARD BASIC")

	conn := dial(t)
	defer conn.Close()

	send(conn,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	resp := send(conn,
		"*1\r\n"+
			"$7\r\nDISCARD\r\n")

	expected := "+OK\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("DISCARD exits transaction")
}



func stage109_DiscardWithoutMulti(t *testing.T) {
	stage("STAGE 109: DISCARD WITHOUT MULTI")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn,
		"*1\r\n"+
			"$7\r\nDISCARD\r\n")

	expected := "-ERR DISCARD without MULTI\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("DISCARD without MULTI rejected")
}


func stage110_DiscardClearsQueue(t *testing.T) {
	stage("STAGE 110: DISCARD CLEARS QUEUE")

	conn := dial(t)
	defer conn.Close()

	send(conn,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn,
		"*3\r\n"+
			"$3\r\nSET\r\n"+
			"$2\r\nxx\r\n"+
			"$2\r\n42\r\n")

	send(conn,
		"*1\r\n"+
			"$7\r\nDISCARD\r\n")

	resp := send(conn,
		"*2\r\n"+
			"$3\r\nGET\r\n"+
			"$2\r\nxx\r\n")

	expected := "$-1\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("queued commands discarded")
}

func stage111_DiscardWrongArguments(t *testing.T) {
	stage("STAGE 111: DISCARD WRONG ARGUMENTS")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn,
		"*2\r\n"+
			"$7\r\nDISCARD\r\n"+
			"$1\r\nx\r\n")

	expected := "-Wrong number of arguments for 'DISCARD' command\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("DISCARD validates argument count")
}

func stage112_ExecAfterDiscard(t *testing.T) {
	stage("STAGE 112: EXEC AFTER DISCARD")

	conn := dial(t)
	defer conn.Close()

	send(conn,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn,
		"*3\r\n"+
			"$3\r\nSET\r\n"+
			"$1\r\nx\r\n"+
			"$1\r\n1\r\n")

	send(conn,
		"*1\r\n"+
			"$7\r\nDISCARD\r\n")

	resp := send(conn,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected := "-ERR EXEC without MULTI\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("EXEC rejected after DISCARD")
}



func stage113_TransactionRuntimeError(t *testing.T) {
	stage("STAGE 113: TRANSACTION RUNTIME ERROR")

	conn := dial(t)
	defer conn.Close()

	send(conn,
		"*3\r\n"+
			"$3\r\nSET\r\n"+
			"$3\r\nxxx\r\n"+
			"$5\r\nhello\r\n")

	send(conn,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn,
		"*2\r\n"+
			"$4\r\nINCR\r\n"+
			"$3\r\nxxx\r\n")

	send(conn,
		"*2\r\n"+
			"$3\r\nGET\r\n"+
			"$3\r\nxxx\r\n")

	resp := send(conn,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected :=
		"*2\r\n" +
			"-ERR value is not an integer or out of range\r\n" +
			"$5\r\nhello\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("runtime errors don't abort transaction")
}



func stage114_TransactionMultipleErrors(t *testing.T) {
	stage("STAGE 114: MULTIPLE RUNTIME ERRORS")

	conn := dial(t)
	defer conn.Close()

	send(conn,
		"*3\r\n"+
			"$3\r\nSET\r\n"+
			"$5\r\nxxxxx\r\n"+
			"$5\r\nhello\r\n")

	send(conn,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn,
		"*2\r\n"+
			"$4\r\nINCR\r\n"+
			"$5\r\nxxxxx\r\n")

	send(conn,
		"*2\r\n"+
			"$4\r\nINCR\r\n"+
			"$5\r\nxxxxx\r\n")

	resp := send(conn,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected :=
		"*2\r\n" +
			"-ERR value is not an integer or out of range\r\n" +
			"-ERR value is not an integer or out of range\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("multiple runtime errors returned")
}

func stage115_TransactionContinuesAfterFailure(t *testing.T) {
	stage("STAGE 115: CONTINUE AFTER FAILURE")

	conn := dial(t)
	defer conn.Close()

	send(conn,
		"*3\r\n"+
			"$3\r\nSET\r\n"+
			"$4\r\nxxxx\r\n"+
			"$5\r\nhello\r\n")

	send(conn,
		"*3\r\n"+
			"$3\r\nSET\r\n"+
			"$1\r\ny\r\n"+
			"$1\r\n5\r\n")

	send(conn,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	send(conn,
		"*2\r\n"+
			"$4\r\nINCR\r\n"+
			"$4\r\nxxxx\r\n")

	send(conn,
		"*2\r\n"+
			"$4\r\nINCR\r\n"+
			"$1\r\ny\r\n")

	send(conn,
		"*2\r\n"+
			"$3\r\nGET\r\n"+
			"$1\r\ny\r\n")

	resp := send(conn,
		"*1\r\n"+
			"$4\r\nEXEC\r\n")

	expected :=
		"*3\r\n" +
			"-ERR value is not an integer or out of range\r\n" +
			":6\r\n" +
			"$1\r\n6\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("transaction continues after failures")
}