package tester

import "testing"

func multi_test(t *testing.T) {
	stage99_MultiBasic(t)
	stage104_NestedMulti(t)
	stage105_MultiWrongArguments(t)
}

func stage99_MultiBasic(t *testing.T) {
	stage("STAGE 99: MULTI")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn,
		"*1\r\n" +
			"$5\r\nMULTI\r\n")

	expected := "+OK\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("MULTI returns OK")
}


func stage104_NestedMulti(t *testing.T) {
	stage("STAGE 104: NESTED MULTI")

	conn := dial(t)
	defer conn.Close()

	send(conn,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	resp := send(conn,
		"*1\r\n"+
			"$5\r\nMULTI\r\n")

	expected := "-ERR MULTI calls cannot be nested\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("nested MULTI rejected")
}


func stage105_MultiWrongArguments(t *testing.T) {
	stage("STAGE 105: MULTI WRONG ARGUMENTS")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn,
		"*2\r\n"+
			"$5\r\nMULTI\r\n"+
			"$1\r\nx\r\n")

	expected := "-Wrong number of arguments for 'MULTI' command\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("MULTI validates argument count")
}