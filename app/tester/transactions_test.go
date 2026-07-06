package tester

import "testing"

func transaction_test(t *testing.T) {
	
	stage100_ExecWithoutMulti(t)
	stage101_MultiQueuesCommands(t)
	stage102_ExecRunsQueuedCommands(t)
	stage103_ExecEmptyQueue(t)
	stage106_ExecWrongArguments(t)
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