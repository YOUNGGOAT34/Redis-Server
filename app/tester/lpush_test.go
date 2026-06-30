package tester

import (
	"fmt"
	"sync"
	"testing"
)

func lpush_test(t *testing.T) {
	stage41_LPushBasic(t)
	stage42_LPushMultipleValues(t)
	stage43_LPushAppend(t)
	stage44_LPushLongString(t)
	stage45_LPushManyElements(t)
	stage46_ConcurrentLPush(t)
	stage47_LPushWrongNumberArguments(t)
	stage48_LPushWrongType(t)
}


func stage41_LPushBasic(t *testing.T) {
	stage("STAGE 41: LPUSH BASIC")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn,
		"*3\r\n$5\r\nLPUSH\r\n$5\r\nllist\r\n$5\r\nhello\r\n")

	expected := ":1\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("basic LPUSH OK")
}


func stage42_LPushMultipleValues(t *testing.T) {
	stage("STAGE 42: LPUSH MULTIPLE VALUES")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn,
		"*5\r\n$5\r\nLPUSH\r\n$6\r\nl1list\r\n$3\r\none\r\n$3\r\ntwo\r\n$5\r\nthree\r\n")

	expected := ":3\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("multiple values inserted OK")
}



func stage43_LPushAppend(t *testing.T) {
	stage("STAGE 43: LPUSH APPEND")

	conn := dial(t)
	defer conn.Close()

	tests := []struct {
		req      string
		expected string
	}{
		{"*3\r\n$5\r\nLPUSH\r\n$6\r\nlqueue\r\n$1\r\na\r\n", ":1\r\n"},
		{"*3\r\n$5\r\nLPUSH\r\n$6\r\nlqueue\r\n$1\r\nb\r\n", ":2\r\n"},
		{"*3\r\n$5\r\nLPUSH\r\n$6\r\nlqueue\r\n$1\r\nc\r\n", ":3\r\n"},
	}

	for _, tc := range tests {
		resp := send(conn, tc.req)

		if resp != tc.expected {
			failf(t, "expected %q got %q", tc.expected, resp)
		}
	}

	pass("append to existing list OK")
}



func stage44_LPushLongString(t *testing.T) {
	stage("STAGE 44: LONG LPUSH")

	conn := dial(t)
	defer conn.Close()

	msg := "abcdefghijklmnopqrstuvwxyz123456789"

	req := fmt.Sprintf(
		"*3\r\n$5\r\nLPUSH\r\n$6\r\nl2list\r\n$%d\r\n%s\r\n",
		len(msg),
		msg,
	)

	resp := send(conn, req)

	expected := ":1\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("long element inserted OK")
}


func stage45_LPushManyElements(t *testing.T) {
	stage("STAGE 45: MANY ELEMENTS")

	conn := dial(t)
	defer conn.Close()

	req := "*12\r\n" +
		"$5\r\nLPUSH\r\n" +
		"$5\r\nlnums\r\n" +
		"$1\r\n1\r\n" +
		"$1\r\n2\r\n" +
		"$1\r\n3\r\n" +
		"$1\r\n4\r\n" +
		"$1\r\n5\r\n" +
		"$1\r\n6\r\n" +
		"$1\r\n7\r\n" +
		"$1\r\n8\r\n" +
		"$1\r\n9\r\n" +
		"$2\r\n10\r\n"

	resp := send(conn, req)

	expected := ":10\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("many elements inserted OK")
}



func stage46_ConcurrentLPush(t *testing.T) {
	stage("STAGE 46: CONCURRENT LPUSH")

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			conn := dial(t)
			defer conn.Close()

			msg := fmt.Sprintf("item-%d", id)

			req := fmt.Sprintf(
				"*3\r\n$5\r\nLPUSH\r\n$6\r\nqueue1\r\n$%d\r\n%s\r\n",
				len(msg),
				msg,
			)

			resp := send(conn, req)

			if len(resp) == 0 || resp[0] != ':' {
				failf(t, "client %d expected integer reply got %q", id, resp)
			}
		}(i)
	}

	wg.Wait()

	pass("100 concurrent LPUSH clients OK")
}


func stage47_LPushWrongNumberArguments(t *testing.T) {
	stage("STAGE 47: LPUSH WRONG ARGUMENTS")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn, "*1\r\n$5\r\nLPUSH\r\n")

	if len(resp) == 0 || resp[0] != '-' {
		failf(t, "expected error reply got %q", resp)
	}

	pass("wrong argument count handled")
}


func stage48_LPushWrongType(t *testing.T) {
	stage("STAGE 48: LPUSH WRONG TYPE")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n")

	resp := send(conn,
		"*3\r\n$5\r\nLPUSH\r\n$3\r\nkey\r\n$5\r\nhello\r\n")

	if len(resp) == 0 || resp[0] != '-' {
		failf(t, "expected WRONGTYPE error got %q", resp)
	}

	pass("wrong type handled OK")
}