package tester

import "testing"


func incr_test(t *testing.T) {
	stage93_IncrNewKey(t)
	stage94_IncrExistingKey(t)
	stage95_IncrWrongType(t)
	stage96_IncrNonInteger(t)
	stage97_IncrWrongArguments(t)
	stage98_IncrOverflow(t)
}


func stage93_IncrNewKey(t *testing.T) {
	stage("STAGE 93: INCR NEW KEY")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn,
		"*2\r\n" +
			"$4\r\nINCR\r\n" +
			"$7\r\ncounter\r\n")

	expected := ":1\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("missing key initialized to 1")
}


func stage94_IncrExistingKey(t *testing.T) {
	stage("STAGE 94: INCR EXISTING")

	conn := dial(t)
	defer conn.Close()

	send(conn,
		"*3\r\n" +
			"$3\r\nSET\r\n" +
			"$7\r\ncounter\r\n" +
			"$2\r\n41\r\n")

	resp := send(conn,
		"*2\r\n" +
			"$4\r\nINCR\r\n" +
			"$7\r\ncounter\r\n")

	expected := ":42\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("existing integer incremented")
}


func stage95_IncrWrongType(t *testing.T) {
	stage("STAGE 95: INCR WRONG TYPE")

	conn := dial(t)
	defer conn.Close()

	send(conn,
		"*3\r\n" +
			"$5\r\nLPUSH\r\n" +
			"$4\r\nlist\r\n" +
			"$1\r\na\r\n")

	resp := send(conn,
		"*2\r\n" +
			"$4\r\nINCR\r\n" +
			"$4\r\nlist\r\n")

	expected := "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("wrong type rejected")
}



func stage96_IncrNonInteger(t *testing.T) {
	stage("STAGE 96: INCR NON INTEGER")

	conn := dial(t)
	defer conn.Close()

	send(conn,
		"*3\r\n" +
			"$3\r\nSET\r\n" +
			"$4\r\nname\r\n" +
			"$4\r\ngoat\r\n")

	resp := send(conn,
		"*2\r\n" +
			"$4\r\nINCR\r\n" +
			"$4\r\nname\r\n")

	expected := "-ERR value is not an integer or out of range\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("non-integer rejected")
}



func stage97_IncrWrongArguments(t *testing.T) {
	stage("STAGE 97: INCR WRONG ARGUMENTS")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn,
		"*1\r\n" +
			"$4\r\nINCR\r\n")

	expected := "-Wrong number of arguments for 'INCR' command\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("wrong argument count detected")
}


func stage98_IncrOverflow(t *testing.T) {
	stage("STAGE 98: INCR OVERFLOW")

	conn := dial(t)
	defer conn.Close()

	send(conn,
		"*3\r\n" +
			"$3\r\nSET\r\n" +
			"$1\r\nx\r\n" +
			"$19\r\n9223372036854775807\r\n")

	resp := send(conn,
		"*2\r\n" +
			"$4\r\nINCR\r\n" +
			"$1\r\nx\r\n")

	expected := "-ERR increment or decrement would overflow\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("integer overflow detected")
}