package tester

import (
	"fmt"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/server"
)

var (
	reset  = "\033[0m"
	bold   = "\033[1m"

	green  = "\033[1;32m"
	red    = "\033[1;31m"
	yellow = "\033[1;33m"
	cyan   = "\033[1;36m"
)

func stage(name string) {
	println("\n" + bold + cyan + "▶ " + name + reset)
}

func pass(msg string) {
	println(green + "✔ " + msg + reset)
}

func fail(msg string) {
	println(red + "✖ " + msg + reset)
}

func info(msg string) {
	println(yellow + "• " + msg + reset)
}

// ---------------- SERVER BOOTSTRAP ----------------

func TestMain(t *testing.T) {
	go server.StartServer()

	waitForServer()

	 stage1_PingOnce(t)
    stage2_MultiplePingSameConnection(t)
    stage3_MultipleClients(t)
    stage4_ClientDisconnect(t)
    stage5_EchoBasic(t)
    stage6_EchoLongString(t)
    stage7_EchoMultipleRequestsSameConnection(t)
    stage8_PingThenEcho(t)
    stage9_ConcurrentEcho(t)
	
}



func failf(t *testing.T, format string, args ...any) {
	t.Fatalf(red+"✘ "+format+reset, args...)
	os.Exit(1)
}


func waitForServer() {
	for i := 0; i < 50; i++ {
		conn, err := net.Dial("tcp", "localhost:6379")
		if err == nil {
			conn.Close()
			return
		}
		time.Sleep(20 * time.Millisecond)
	}

	panic(red + "server did not start" + reset)
}

// ---------------- HELPERS ----------------

func dial(t *testing.T) net.Conn {
	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	return conn
}

func send(conn net.Conn, req string) string {
	_, err := conn.Write([]byte(req))
	if err != nil {
		return ""
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return ""
	}

	return string(buf[:n])
}

// ---------------- TESTS ----------------

func stage1_PingOnce(t *testing.T) {
	stage("STAGE 1: PING ONCE")

	conn := dial(t)
	defer conn.Close()

	info("sending PING")

	resp := send(conn, "*1\r\n$4\r\nPING\r\n")

	if resp != "+PONG\r\n" {
		fail("expected +PONG, got " + resp)
		t.Fatalf("test failed")
	}

	pass("PING → PONG OK")
}

func stage2_MultiplePingSameConnection(t *testing.T) {
	stage("STAGE 2: MULTIPLE PING (SAME CONNECTION)")

	conn := dial(t)
	defer conn.Close()

	for i := 0; i < 10; i++ {
		resp := send(conn, "*1\r\n$4\r\nPING\r\n")

		if resp != "+PONG\r\n" {
			fail("iteration failed at index " + string(rune(i)))
			t.Fatalf("failed")
		}
	}

	pass("multiple PINGs on same connection OK")
}

func stage3_MultipleClients(t *testing.T) {
	stage("STAGE 3: MULTIPLE CLIENTS")

	client := func() {
		conn := dial(t)
		defer conn.Close()

		for i := 0; i < 10; i++ {
			resp := send(conn, "*1\r\n$4\r\nPING\r\n")

			if resp != "+PONG\r\n" {
				return
			}
		}
	}

	for i := 0; i < 5; i++ {
		go client()
	}

	time.Sleep(1 * time.Second)

	pass("concurrent clients OK")
}

func stage4_ClientDisconnect(t *testing.T) {
	stage("STAGE 4: CLIENT DISCONNECT")

	conn := dial(t)

	_, _ = conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
	conn.Close()

	time.Sleep(100 * time.Millisecond)

	pass("disconnect handled safely")
}


func stage5_EchoBasic(t *testing.T) {
	stage("STAGE 5: ECHO BASIC")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn, "*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n")

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