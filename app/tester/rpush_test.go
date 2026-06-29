package tester

import (
	"fmt"
	"sync"
	"testing"
)





func rpush_test(t *testing.T){
	  stage27_RPushBasic(t)
	  stage28_RPushMultipleValues(t)
	  stage29_RPushAppend(t)
	  stage30_RPushLongString(t)
	  stage31_RPushManyElements(t)
	  stage32_ConcurrentRPush(t)
	  stage33_RPushWrongNumberArguments(t)
}

func stage27_RPushBasic(t *testing.T) {
	stage("STAGE 27: RPUSH BASIC")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn, "*3\r\n$5\r\nRPUSH\r\n$6\r\nmylist\r\n$5\r\nhello\r\n")

	expected := ":1\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("basic RPUSH OK")
}


func stage28_RPushMultipleValues(t *testing.T) {
	stage("STAGE 28: RPUSH MULTIPLE VALUES")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn,
		"*5\r\n$5\r\nRPUSH\r\n$6\r\nmylist\r\n$3\r\none\r\n$3\r\ntwo\r\n$5\r\nthree\r\n")

	expected := ":4\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("multiple values inserted OK")
}


func stage29_RPushAppend(t *testing.T) {
	stage("STAGE 29: RPUSH APPEND")

	conn := dial(t)
	defer conn.Close()

	tests := []struct {
		req      string
		expected string
	}{
		{"*3\r\n$5\r\nRPUSH\r\n$5\r\nqueue\r\n$1\r\na\r\n", ":1\r\n"},
		{"*3\r\n$5\r\nRPUSH\r\n$5\r\nqueue\r\n$1\r\nb\r\n", ":2\r\n"},
		{"*3\r\n$5\r\nRPUSH\r\n$5\r\nqueue\r\n$1\r\nc\r\n", ":3\r\n"},
	}

	for _, tc := range tests {
		resp := send(conn, tc.req)

		if resp != tc.expected {
			failf(t, "expected %q got %q", tc.expected, resp)
		}
	}

	pass("append to existing list OK")
}


func stage30_RPushLongString(t *testing.T) {
	stage("STAGE 30: LONG RPUSH")

	conn := dial(t)
	defer conn.Close()

	msg := "abcdefghijklmnopqrstuvwxyz123456789"

	req := fmt.Sprintf(
		"*3\r\n$5\r\nRPUSH\r\n$4\r\nlist\r\n$%d\r\n%s\r\n",
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


func stage31_RPushManyElements(t *testing.T) {
	stage("STAGE 31: MANY ELEMENTS")

	conn := dial(t)
	defer conn.Close()

	req := "*12\r\n" +
		"$5\r\nRPUSH\r\n" +
		"$4\r\nnums\r\n" +
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



func stage32_ConcurrentRPush(t *testing.T) {
	stage("STAGE 32: CONCURRENT RPUSH")

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			conn := dial(t)
			defer conn.Close()

			msg := fmt.Sprintf("item-%d", id)

			req := fmt.Sprintf(
				"*3\r\n$5\r\nRPUSH\r\n$5\r\nqueue\r\n$%d\r\n%s\r\n",
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

	pass("100 concurrent RPUSH clients OK")
}


func stage33_RPushWrongNumberArguments(t *testing.T) {
	stage("STAGE 33: RPUSH WRONG ARGUMENTS")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn, "*1\r\n$5\r\nRPUSH\r\n")

	if len(resp) == 0 {
		failf(t, "expected an error response")
	}

	pass("wrong argument count handled")
}