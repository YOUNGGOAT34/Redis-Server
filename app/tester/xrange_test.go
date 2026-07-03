package tester

import (
	"sync"
	"testing"
)


func xrange_test(t *testing.T) {
	// stage77_XRangeBasic(t)
	// stage78_XRangeSingleEntry(t)
	// stage79_XRangePartialRange(t)
	// stage80_XRangeNoMatches(t)
	// stage81_XRangeMissingKey(t)
	// stage82_XRangeWrongType(t)
	// stage83_XRangeWrongArguments(t)
	stage84_XRangeConcurrentReads(t)
}

// ------------------------------------------------------------
func stage77_XRangeBasic(t *testing.T) {
	stage("STAGE 77: XRANGE BASIC")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*5\r\n$4\r\nXADD\r\n$9\r\nstream-77\r\n$3\r\n1-0\r\n$1\r\na\r\n$1\r\n1\r\n")
	send(conn, "*5\r\n$4\r\nXADD\r\n$9\r\nstream-77\r\n$3\r\n2-0\r\n$1\r\nb\r\n$1\r\n2\r\n")
	send(conn, "*5\r\n$4\r\nXADD\r\n$9\r\nstream-77\r\n$3\r\n3-0\r\n$1\r\nc\r\n$1\r\n3\r\n")

	resp := send(conn,
		"*4\r\n" +
			"$6\r\nXRANGE\r\n" +
			"$9\r\nstream-77\r\n" +
			"$3\r\n1-0\r\n" +
			"$3\r\n3-0\r\n")

	expected :=
		"*3\r\n" +
			"*2\r\n" +
			"$3\r\n1-0\r\n" +
			"*2\r\n" +
			"$1\r\na\r\n" +
			"$1\r\n1\r\n" +
			"*2\r\n" +
			"$3\r\n2-0\r\n" +
			"*2\r\n" +
			"$1\r\nb\r\n" +
			"$1\r\n2\r\n" +
			"*2\r\n" +
			"$3\r\n3-0\r\n" +
			"*2\r\n" +
			"$1\r\nc\r\n" +
			"$1\r\n3\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("basic XRANGE OK")
}

func stage78_XRangeSingleEntry(t *testing.T) {
	stage("STAGE 78: XRANGE SINGLE ENTRY")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*5\r\n$4\r\nXADD\r\n$9\r\nstream-78\r\n$3\r\n5-0\r\n$4\r\nname\r\n$4\r\ngoat\r\n")

	resp := send(conn, "*4\r\n$6\r\nXRANGE\r\n$9\r\nstream-78\r\n$3\r\n5-0\r\n$3\r\n5-0\r\n")

	expected :=
	"*1\r\n" +
	"*2\r\n" +
	"$3\r\n5-0\r\n" +
	"*2\r\n" +
	"$4\r\nname\r\n" +
	"$4\r\ngoat\r\n"


	if resp != expected {
			failf(t, "expected %q got %q", expected, resp)
		}

	pass("single entry returned")
}

func stage79_XRangePartialRange(t *testing.T) {
	stage("STAGE 79: XRANGE PARTIAL RANGE")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*5\r\n$4\r\nXADD\r\n$9\r\nstream-79\r\n$3\r\n1-0\r\n$1\r\na\r\n$1\r\n1\r\n")
	send(conn, "*5\r\n$4\r\nXADD\r\n$9\r\nstream-79\r\n$3\r\n2-0\r\n$1\r\nb\r\n$1\r\n2\r\n")
	send(conn, "*5\r\n$4\r\nXADD\r\n$9\r\nstream-79\r\n$3\r\n3-0\r\n$1\r\nc\r\n$1\r\n3\r\n")

	resp := send(conn, "*4\r\n$6\r\nXRANGE\r\n$9\r\nstream-79\r\n$3\r\n2-0\r\n$3\r\n3-0\r\n")

		expected :=
		"*2\r\n" +
		"*2\r\n" +
		"$3\r\n2-0\r\n" +
		"*2\r\n" +
		"$1\r\nb\r\n" +
		"$1\r\n2\r\n" +
		"*2\r\n" +
		"$3\r\n3-0\r\n" +
		"*2\r\n" +
		"$1\r\nc\r\n" +
		"$1\r\n3\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("partial range OK")
}

func stage80_XRangeNoMatches(t *testing.T) {
	stage("STAGE 80: XRANGE EMPTY RANGE")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*5\r\n$4\r\nXADD\r\n$9\r\nstream-80\r\n$3\r\n1-0\r\n$1\r\na\r\n$1\r\n1\r\n")

	resp := send(conn, "*4\r\n$6\r\nXRANGE\r\n$9\r\nstream-80\r\n$3\r\n9-0\r\n$4\r\n10-0\r\n")

	expected := "*0\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("empty range handled")
}

func stage81_XRangeMissingKey(t *testing.T) {
	stage("STAGE 81: XRANGE MISSING KEY")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn, "*4\r\n$6\r\nXRANGE\r\n$9\r\nstream-81\r\n$1\r\n-\r\n$1\r\n+\r\n")

	expected := "*0\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("missing stream handled")
}

func stage82_XRangeWrongType(t *testing.T) {
	stage("STAGE 82: XRANGE WRONG TYPE")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*3\r\n$3\r\nSET\r\n$9\r\nstream-82\r\n$5\r\nhello\r\n")

	resp := send(conn, "*4\r\n$6\r\nXRANGE\r\n$9\r\nstream-82\r\n$1\r\n-\r\n$1\r\n+\r\n")

	if len(resp) == 0 || resp[0] != '-' {
		failf(t, "expected WRONGTYPE error got %q", resp)
	}

	pass("wrong type handled")
}

func stage83_XRangeWrongArguments(t *testing.T) {
	stage("STAGE 83: XRANGE WRONG ARGUMENTS")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn, "*1\r\n$6\r\nXRANGE\r\n")

	if len(resp) == 0 || resp[0] != '-' {
		failf(t, "expected error got %q", resp)
	}

	pass("wrong arguments handled")
}

func stage84_XRangeConcurrentReads(t *testing.T) {
	stage("STAGE 84: CONCURRENT XRANGE")

	conn := dial(t)

	send(conn, "*5\r\n$4\r\nXADD\r\n$9\r\nstream-84\r\n$3\r\n1-0\r\n$1\r\na\r\n$1\r\n1\r\n")
	send(conn, "*5\r\n$4\r\nXADD\r\n$9\r\nstream-84\r\n$3\r\n2-0\r\n$1\r\nb\r\n$1\r\n2\r\n")
	conn.Close()

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			conn := dial(t)
			defer conn.Close()

			resp := send(conn, "*4\r\n$6\r\nXRANGE\r\n$9\r\nstream-84\r\n$1\r\n-\r\n$3\r\n2-0\r\n")

			if len(resp) == 0 || resp[0] != '*' {
				failf(t, "expected RESP array got %q", resp)
			}
		}()
	}

	wg.Wait()

	pass("100 concurrent XRANGE clients OK")
}