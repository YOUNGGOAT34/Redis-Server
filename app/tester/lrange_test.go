package tester

import "testing"

func lrange_test(t *testing.T) {
	stage34_LRangeBasic(t)
	stage35_LRangeSubset(t)
	stage36_LRangeNegativeIndexes(t)
	stage37_LRangeOutOfBounds(t)
	stage38_LRangeMissingKey(t)
	stage39_LRangeSingleElement(t)
	stage40_LRangeWrongType(t)
}

func stage34_LRangeBasic(t *testing.T) {
	stage("STAGE 34: LRANGE BASIC")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*5\r\n$5\r\nRPUSH\r\n$5\r\nlist1\r\n$3\r\none\r\n$3\r\ntwo\r\n$5\r\nthree\r\n")

	resp := send(conn, "*4\r\n$6\r\nLRANGE\r\n$5\r\nlist1\r\n$1\r\n0\r\n$2\r\n-1\r\n")

	expected := "*3\r\n$3\r\none\r\n$3\r\ntwo\r\n$5\r\nthree\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("basic LRANGE OK")
}

func stage35_LRangeSubset(t *testing.T) {
	stage("STAGE 35: LRANGE SUBSET")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*6\r\n$5\r\nRPUSH\r\n$7\r\nletters\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n$1\r\nd\r\n")

	resp := send(conn, "*4\r\n$6\r\nLRANGE\r\n$7\r\nletters\r\n$1\r\n1\r\n$1\r\n2\r\n")

	expected := "*2\r\n$1\r\nb\r\n$1\r\nc\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("subset returned OK")
}

func stage36_LRangeNegativeIndexes(t *testing.T) {
	stage("STAGE 36: LRANGE NEGATIVE INDEXES")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*6\r\n$5\r\nRPUSH\r\n$5\r\nnums1\r\n$1\r\n1\r\n$1\r\n2\r\n$1\r\n3\r\n$1\r\n4\r\n")

	resp := send(conn, "*4\r\n$6\r\nLRANGE\r\n$5\r\nnums1\r\n$2\r\n-2\r\n$2\r\n-1\r\n")

	expected := "*2\r\n$1\r\n3\r\n$1\r\n4\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("negative indexes OK")
}

func stage37_LRangeOutOfBounds(t *testing.T) {
	stage("STAGE 37: LRANGE OUT OF BOUNDS")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*5\r\n$5\r\nRPUSH\r\n$5\r\nnums2\r\n$1\r\nA\r\n$1\r\nB\r\n$1\r\nC\r\n")

	resp := send(conn, "*4\r\n$6\r\nLRANGE\r\n$5\r\nnums2\r\n$1\r\n0\r\n$2\r\n99\r\n")

	expected := "*3\r\n$1\r\nA\r\n$1\r\nB\r\n$1\r\nC\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("end index clipped correctly")
}

func stage38_LRangeMissingKey(t *testing.T) {
	stage("STAGE 38: LRANGE MISSING KEY")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn, "*4\r\n$6\r\nLRANGE\r\n$7\r\nunknown\r\n$1\r\n0\r\n$2\r\n-1\r\n")

	expected := "*0\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("missing key handled OK")
}

func stage39_LRangeSingleElement(t *testing.T) {
	stage("STAGE 39: LRANGE SINGLE ELEMENT")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*5\r\n$5\r\nRPUSH\r\n$4\r\nsolo\r\n$5\r\napple\r\n$6\r\nbanana\r\n$6\r\norange\r\n")

	resp := send(conn, "*4\r\n$6\r\nLRANGE\r\n$4\r\nsolo\r\n$1\r\n1\r\n$1\r\n1\r\n")

	expected := "*1\r\n$6\r\nbanana\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("single element range OK")
}

func stage40_LRangeWrongType(t *testing.T) {
	stage("STAGE 40: LRANGE WRONG TYPE")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n")

	resp := send(conn, "*4\r\n$6\r\nLRANGE\r\n$3\r\nkey\r\n$1\r\n0\r\n$2\r\n-1\r\n")

	if len(resp) == 0 || resp[0] != '-' {
		failf(t, "expected RESP error, got %q", resp)
	}

	pass("wrong type handled OK")
}