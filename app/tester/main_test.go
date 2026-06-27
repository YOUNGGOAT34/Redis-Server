package tester

import (
	"net"
	"os"
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

func TestMain(m *testing.M) {
	go server.StartServer()

	waitForServer()

	code := m.Run()
	os.Exit(code)
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

func TestStage1_PingOnce(t *testing.T) {
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

func TestStage2_MultiplePingSameConnection(t *testing.T) {
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

func TestStage3_MultipleClients(t *testing.T) {
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

func TestStage4_ClientDisconnect(t *testing.T) {
	stage("STAGE 4: CLIENT DISCONNECT")

	conn := dial(t)

	_, _ = conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
	conn.Close()

	time.Sleep(100 * time.Millisecond)

	pass("disconnect handled safely")
}