package tester

import (
	"fmt"
	"sync"
	"testing"
)


func echo_test(t *testing.T){
	 stage5_EchoBasic(t)
    stage6_EchoLongString(t)
    stage7_EchoMultipleRequestsSameConnection(t)
    stage8_PingThenEcho(t)
    stage9_ConcurrentEcho(t)
}


func stage5_EchoBasic(t *testing.T) {
	stage("STAGE 5: ECHO BASIC")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn, "*2\r\n$4\r\nECHo\r\n$3\r\nhey\r\n")

	expected := "$3\r\nhey\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("basic ECHO OK")
}

func stage6_EchoLongString(t *testing.T) {
	stage("STAGE 6: LONG ECHO")

	conn := dial(t)
	defer conn.Close()

	msg := "abcdefghijklmnopqrstuvwxyz12345678"

	req := fmt.Sprintf("*2\r\n$4\r\nECHO\r\n$%d\r\n%s\r\n", len(msg), msg)
	expected := fmt.Sprintf("$%d\r\n%s\r\n", len(msg), msg)

	resp := send(conn, req)

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("multi-digit bulk string length OK")
}

func stage7_EchoMultipleRequestsSameConnection(t *testing.T) {
	stage("STAGE 7: MULTIPLE ECHO (SAME CONNECTION)")

	conn := dial(t)
	defer conn.Close()

	tests := []string{
		"hello",
		"redis",
		"codecrafters",
		"golang",
	}

	for _, msg := range tests {
		req := fmt.Sprintf("*2\r\n$4\r\nECHO\r\n$%d\r\n%s\r\n", len(msg), msg)
		expected := fmt.Sprintf("$%d\r\n%s\r\n", len(msg), msg)

		resp := send(conn, req)

		if resp != expected {
			failf(t, "expected %q got %q", expected, resp)
		}
	}

	pass("multiple ECHOs on same connection OK")
}

func stage8_PingThenEcho(t *testing.T) {
	stage("STAGE 8: PING + ECHO")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn, "*1\r\n$4\r\nPING\r\n")
	if resp != "+PONG\r\n" {
		failf(t, "expected +PONG got %q", resp)
	}

	resp = send(conn, "*2\r\n$4\r\nECHO\r\n$5\r\nhello\r\n")
	if resp != "$5\r\nhello\r\n" {
		failf(t, "expected bulk string got %q", resp)
	}

	resp = send(conn, "*1\r\n$4\r\nPING\r\n")
	if resp != "+PONG\r\n" {
		failf(t, "expected +PONG got %q", resp)
	}

	pass("PING and ECHO coexist OK")
}

func stage9_ConcurrentEcho(t *testing.T) {
	stage("STAGE 9: CONCURRENT ECHO")

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			conn := dial(t)
			defer conn.Close()

			msg := fmt.Sprintf("client-%d", id)

			req := fmt.Sprintf("*2\r\n$4\r\nECHO\r\n$%d\r\n%s\r\n", len(msg), msg)
			expected := fmt.Sprintf("$%d\r\n%s\r\n", len(msg), msg)

			resp := send(conn, req)

			if resp != expected {
				failf(t, "client %d expected %q got %q", id, expected, resp)
			}
		}(i)
	}

	wg.Wait()

	pass("100 concurrent clients OK")
}
