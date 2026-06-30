package tester

import (
	"fmt"
	"sync"
	"testing"
)

func lpop_test(t *testing.T) {
	stage47_LPopBasic(t)
	stage48_LPopOrder(t)
	stage49_LPopUntilEmpty(t)
	stage50_LPopMissingKey(t)
	stage51_LPopWrongType(t)
	stage52_LPopWrongArguments(t)
	stage53_ConcurrentLPop(t)
}


func stage47_LPopBasic(t *testing.T) {
	stage("STAGE 47: LPOP BASIC")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*4\r\n$5\r\nLPUSH\r\n$9\r\nlpletters\r\n$1\r\nb\r\n$1\r\na\r\n")

	resp := send(conn, "*2\r\n$4\r\nLPOP\r\n$9\r\nlpletters\r\n")

	expected := "$1\r\na\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("basic LPOP OK")
}

func stage48_LPopOrder(t *testing.T) {
	stage("STAGE 48: LPOP ORDER")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*5\r\n$5\r\nLPUSH\r\n$7\r\nlpqueue\r\n$1\r\nc\r\n$1\r\nb\r\n$1\r\na\r\n")

	expected := []string{
		"$1\r\na\r\n",
		"$1\r\nb\r\n",
		"$1\r\nc\r\n",
	}

	for _, want := range expected {
		resp := send(conn, "*2\r\n$4\r\nLPOP\r\n$7\r\nlpqueue\r\n")

		if resp != want {
			failf(t, "expected %q got %q", want, resp)
		}
	}

	pass("LPOP preserves FIFO removal from the left")
}

func stage49_LPopUntilEmpty(t *testing.T) {
	stage("STAGE 49: LPOP UNTIL EMPTY")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*3\r\n$5\r\nLPUSH\r\n$6\r\nlplist\r\n$5\r\nhello\r\n")

	resp := send(conn, "*2\r\n$4\r\nLPOP\r\n$6\r\nlplist\r\n")
	if resp != "$5\r\nhello\r\n" {
		failf(t, "expected hello got %q", resp)
	}

	resp = send(conn, "*2\r\n$4\r\nLPOP\r\n$6\r\nlplist\r\n")
	if resp != "$-1\r\n" {
		failf(t, "expected nil bulk string got %q", resp)
	}

	pass("empty list handled correctly")
}

func stage50_LPopMissingKey(t *testing.T) {
	stage("STAGE 50: LPOP MISSING KEY")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn, "*2\r\n$4\r\nLPOP\r\n$7\r\nunknown\r\n")

	expected := "$-1\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("missing key handled")
}

func stage51_LPopWrongType(t *testing.T) {
	stage("STAGE 51: LPOP WRONG TYPE")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n")

	resp := send(conn, "*2\r\n$4\r\nLPOP\r\n$3\r\nkey\r\n")

	if len(resp) == 0 || resp[0] != '-' {
		failf(t, "expected error response got %q", resp)
	}

	pass("wrong type handled")
}

func stage52_LPopWrongArguments(t *testing.T) {
	stage("STAGE 52: LPOP WRONG ARGUMENTS")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn, "*1\r\n$4\r\nLPOP\r\n")

	if len(resp) == 0 || resp[0] != '-' {
		failf(t, "expected error response got %q", resp)
	}

	pass("wrong argument count handled")
}

func stage53_ConcurrentLPop(t *testing.T) {
	stage("STAGE 53: CONCURRENT LPOP")

	conn := dial(t)
	defer conn.Close()

	// preload 100 elements
	for i := 0; i < 100; i++ {
		send(conn, fmt.Sprintf("*3\r\n$5\r\nLPUSH\r\n$4\r\nwork\r\n$2\r\n%02d\r\n", i))
	}
	conn.Close()

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			c := dial(t)
			defer c.Close()

			resp := send(c, "*2\r\n$4\r\nLPOP\r\n$4\r\nwork\r\n")

			if len(resp) == 0 {
				failf(t, "expected bulk string reply")
			}
		}()
	}

	wg.Wait()

	pass("100 concurrent LPOP clients OK")
}