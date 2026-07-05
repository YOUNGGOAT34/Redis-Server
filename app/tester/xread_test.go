package tester

import (
	"testing"
	"time"
)

func xread_test(t *testing.T) {
	// stage82_XReadBasic(t)
	// stage83_XReadAfterID(t)
	// stage84_XReadLatestOnly(t)
	// stage85_XReadMultipleStreams(t)
	// stage86_XReadMissingStream(t)
	// stage87_XReadEmptyResult(t)
	stage88_XReadBlockWakeup(t)
}

func stage82_XReadBasic(t *testing.T) {
	stage("STAGE 82: XREAD BASIC")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*5\r\n$4\r\nXADD\r\n$10\r\nstream-82r\r\n$3\r\n1-0\r\n$1\r\na\r\n$1\r\n1\r\n")
	send(conn, "*5\r\n$4\r\nXADD\r\n$10\r\nstream-82r\r\n$3\r\n2-0\r\n$1\r\nb\r\n$1\r\n2\r\n")

	resp := send(conn,
		"*4\r\n" +
			"$5\r\nXREAD\r\n" +
			"$7\r\nSTREAMS\r\n" +
			"$10\r\nstream-82r\r\n" +
			"$3\r\n0-0\r\n")
    
	expected :=
		"*1\r\n" +
			"*2\r\n" +
			"$10\r\nstream-82r\r\n" +
			"*2\r\n" +
			"*2\r\n" +
			"$3\r\n1-0\r\n" +
			"*2\r\n$1\r\na\r\n$1\r\n1\r\n" +
			"*2\r\n" +
			"$3\r\n2-0\r\n" +
			"*2\r\n$1\r\nb\r\n$1\r\n2\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("basic XREAD OK")
}

func stage83_XReadAfterID(t *testing.T) {
	stage("STAGE 83: XREAD AFTER ID")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*5\r\n$4\r\nXADD\r\n$9\r\nstream-83\r\n$3\r\n1-0\r\n$1\r\na\r\n$1\r\n1\r\n")
	send(conn, "*5\r\n$4\r\nXADD\r\n$9\r\nstream-83\r\n$3\r\n2-0\r\n$1\r\nb\r\n$1\r\n2\r\n")
	send(conn, "*5\r\n$4\r\nXADD\r\n$9\r\nstream-83\r\n$3\r\n3-0\r\n$1\r\nc\r\n$1\r\n3\r\n")

	resp := send(conn,
		"*4\r\n" +
			"$5\r\nXREAD\r\n" +
			"$7\r\nSTREAMS\r\n" +
			"$9\r\nstream-83\r\n" +
			"$3\r\n1-0\r\n")

	expected :=
		"*1\r\n" +
			"*2\r\n" +
			"$9\r\nstream-83\r\n" +
			"*2\r\n" +
			"*2\r\n$3\r\n2-0\r\n*2\r\n$1\r\nb\r\n$1\r\n2\r\n" +
			"*2\r\n$3\r\n3-0\r\n*2\r\n$1\r\nc\r\n$1\r\n3\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("entries after ID returned OK")
}

func stage84_XReadLatestOnly(t *testing.T) {
	stage("STAGE 84: XREAD LATEST ENTRY")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*5\r\n$4\r\nXADD\r\n$9\r\nstream-84\r\n$3\r\n5-0\r\n$1\r\nx\r\n$1\r\n9\r\n")

	resp := send(conn,
		"*4\r\n" +
			"$5\r\nXREAD\r\n" +
			"$7\r\nSTREAMS\r\n" +
			"$9\r\nstream-84\r\n" +
			"$3\r\n4-0\r\n")

	expected :=
		"*1\r\n" +
			"*2\r\n" +
			"$9\r\nstream-84\r\n" +
			"*1\r\n" +
			"*2\r\n$3\r\n5-0\r\n*2\r\n$1\r\nx\r\n$1\r\n9\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("latest entry returned OK")
}

func stage85_XReadMultipleStreams(t *testing.T) {
	stage("STAGE 85: XREAD MULTIPLE STREAMS")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*5\r\n$4\r\nXADD\r\n$2\r\ns1\r\n$3\r\n1-0\r\n$1\r\na\r\n$1\r\n1\r\n")
	send(conn, "*5\r\n$4\r\nXADD\r\n$2\r\ns2\r\n$3\r\n1-0\r\n$1\r\nb\r\n$1\r\n2\r\n")

	resp := send(conn,
		"*6\r\n" +
			"$5\r\nXREAD\r\n" +
			"$7\r\nSTREAMS\r\n" +
			"$2\r\ns1\r\n" +
			"$2\r\ns2\r\n" +
			"$3\r\n0-0\r\n" +
			"$3\r\n0-0\r\n")

	if len(resp) == 0 || resp[0] != '*' {
		failf(t, "expected RESP array got %q", resp)
	}

	pass("multiple streams supported")
}

func stage86_XReadMissingStream(t *testing.T) {
	stage("STAGE 86: XREAD MISSING STREAM")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn,
		"*4\r\n" +
			"$5\r\nXREAD\r\n" +
			"$7\r\nSTREAMS\r\n" +
			"$7\r\nmissing\r\n" +
			"$3\r\n0-0\r\n")

	expected := "$-1\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("missing stream handled")
}

func stage87_XReadEmptyResult(t *testing.T) {
	stage("STAGE 87: XREAD EMPTY RESULT")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*5\r\n$4\r\nXADD\r\n$9\r\nstream-87\r\n$3\r\n1-0\r\n$1\r\na\r\n$1\r\n1\r\n")

	resp := send(conn,
		"*4\r\n" +
			"$5\r\nXREAD\r\n" +
			"$7\r\nSTREAMS\r\n" +
			"$9\r\nstream-87\r\n" +
			"$3\r\n1-0\r\n")

	expected := "$-1\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("empty read handled")
}


func stage88_XReadBlockWakeup(t *testing.T) {
	stage("STAGE 88: XREAD BLOCK WAKEUP")

	reader := dial(t)
	writer := dial(t)

	defer reader.Close()
	defer writer.Close()

	done := make(chan string)

	go func() {
		resp := send(reader,
			"*6\r\n" +
				"$5\r\nXREAD\r\n" +
				"$5\r\nBLOCK\r\n" +
				"$4\r\n5000\r\n" +
				"$7\r\nSTREAMS\r\n" +
				"$9\r\nstream-88\r\n" +
				"$3\r\n0-0\r\n")

		done <- resp
	}()

	time.Sleep(100 * time.Millisecond)

	send(writer,
		"*5\r\n" +
			"$4\r\nXADD\r\n" +
			"$9\r\nstream-88\r\n" +
			"$3\r\n1-0\r\n" +
			"$1\r\na\r\n" +
			"$1\r\n1\r\n")

	resp := <-done

	expected :=
		"*1\r\n" +
			"*2\r\n" +
			"$9\r\nstream-88\r\n" +
			"*1\r\n" +
			"*2\r\n" +
			"$3\r\n1-0\r\n" +
			"*2\r\n" +
			"$1\r\na\r\n" +
			"$1\r\n1\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("blocking XREAD wakes after XADD")
}