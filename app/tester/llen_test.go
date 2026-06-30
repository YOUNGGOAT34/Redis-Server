package tester

import (
	"sync"
	"testing"
)

func llen_test(t *testing.T) {
	stage49_LLenEmptyList(t)
	stage50_LLenSingleElement(t)
	stage51_LLenMultipleElements(t)
	stage52_LLenAfterLPush(t)
	stage53_LLenAfterRPush(t)
	stage54_LLenMissingKey(t)
	stage55_LLenWrongType(t)
	stage56_ConcurrentLLen(t)
}

func stage49_LLenEmptyList(t *testing.T) {
	stage("STAGE 49: LLEN EMPTY LIST")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*2\r\n$5\r\nRPUSH\r\n$5\r\nempty\r\n")
	resp := send(conn, "*2\r\n$4\r\nLLEN\r\n$5\r\nempty\r\n")

	expected := ":0\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("empty list length OK")
}

func stage50_LLenSingleElement(t *testing.T) {
	stage("STAGE 50: LLEN SINGLE ELEMENT")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*3\r\n$7\r\nRPUSH\r\n$4\r\nlelist\r\n$5\r\nhello\r\n")

	resp := send(conn, "*2\r\n$4\r\nLLEN\r\n$4\r\nlist\r\n")

	expected := ":1\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("single element length OK")
}

func stage51_LLenMultipleElements(t *testing.T) {
	stage("STAGE 51: LLEN MULTIPLE ELEMENTS")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*6\r\n$7\r\nRPUSH\r\n$5\r\nlenums1\r\n$1\r\n1\r\n$1\r\n2\r\n$1\r\n3\r\n$1\r\n4\r\n")

	resp := send(conn, "*2\r\n$4\r\nLLEN\r\n$5\r\nnums1\r\n")

	expected := ":4\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("multiple elements length OK")
}

func stage52_LLenAfterLPush(t *testing.T) {
	stage("STAGE 52: LLEN AFTER LPUSH")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*5\r\n$5\r\nLPUSH\r\n$5\r\nstack\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n")

	resp := send(conn, "*2\r\n$4\r\nLLEN\r\n$5\r\nstack\r\n")

	expected := ":3\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("LPUSH length OK")
}

func stage53_LLenAfterRPush(t *testing.T) {
	stage("STAGE 53: LLEN AFTER RPUSH")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*5\r\n$5\r\nRPUSH\r\n$7\r\nlequeue\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n")

	resp := send(conn, "*2\r\n$4\r\nLLEN\r\n$7\r\nlequeue\r\n")

	expected := ":3\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("RPUSH length OK")
}

func stage54_LLenMissingKey(t *testing.T) {
	stage("STAGE 54: LLEN MISSING KEY")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn, "*2\r\n$4\r\nLLEN\r\n$7\r\nunknown\r\n")

	expected := ":0\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("missing key handled OK")
}

func stage55_LLenWrongType(t *testing.T) {
	stage("STAGE 55: LLEN WRONG TYPE")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n")

	resp := send(conn, "*2\r\n$4\r\nLLEN\r\n$3\r\nkey\r\n")

	if len(resp) == 0 || resp[0] != '-' {
		failf(t, "expected WRONGTYPE error got %q", resp)
	}

	pass("wrong type handled OK")
}

func stage56_ConcurrentLLen(t *testing.T) {
	stage("STAGE 56: CONCURRENT LLEN")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*6\r\n$5\r\nRPUSH\r\n$8\r\nle2queue\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n$1\r\nd\r\n")

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			c := dial(t)
			defer c.Close()

			resp := send(c, "*2\r\n$4\r\nLLEN\r\n$8\r\nle2queue\r\n")

			if resp != ":4\r\n" {
				failf(t, "expected %q got %q", ":4\r\n", resp)
			}
		}()
	}

	wg.Wait()

	pass("100 concurrent LLEN clients OK")
}