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

func TestMain(t *testing.T) {
	go server.StartServer()

	waitForServer()
    
	//  pingtest(t)
   //  echo_test(t)
	//  set_test(t)
	//  get_test(t)
	//  expiry_test(t)
	//  rpush_test(t)
	//  lrange_test(t)
	// lpush_test(t)
	// llen_test(t)
	// lpop_test(t)
	// blpop_test(t)
	// type_test(t)
	xadd_test(t)
	
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


