package tester

import "testing"

func multi_test(t *testing.T) {
	stage99_MultiBasic(t)
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