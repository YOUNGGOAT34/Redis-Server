package tester

import (
	"testing"
	"time"
)


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




