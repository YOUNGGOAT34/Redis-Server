package tester

import (
	"fmt"
	"sync"
	"testing"
	"time"
)


func stage21_SetEX(t *testing.T) {
	stage("STAGE 21: SET EX")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn, "*5\r\n$3\r\nSET\r\n$4\r\nname\r\n$4\r\ngoat\r\n$2\r\nEX\r\n$1\r\n2\r\n")

	if resp != "+OK\r\n" {
		failf(t, "expected +OK got %q", resp)
	}

	resp = send(conn, "*2\r\n$3\r\nGET\r\n$4\r\nname\r\n")

	expected := "$4\r\ngoat\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("SET EX stores value before expiry")
}


func stage22_EXExpires(t *testing.T) {
	stage("STAGE 22: EX EXPIRY")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*5\r\n$3\r\nSET\r\n$4\r\nname\r\n$4\r\ngoat\r\n$2\r\nEX\r\n$1\r\n1\r\n")

	time.Sleep(1100 * time.Millisecond)

	resp := send(conn, "*2\r\n$3\r\nGET\r\n$4\r\nname\r\n")

	expected := "$-1\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("EX expiration OK")
}


func stage23_SetPX(t *testing.T) {
	stage("STAGE 23: SET PX")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*5\r\n$3\r\nSET\r\n$4\r\nname\r\n$4\r\ngoat\r\n$2\r\nPX\r\n$3\r\n500\r\n")

	resp := send(conn, "*2\r\n$3\r\nGET\r\n$4\r\nname\r\n")

	expected := "$4\r\ngoat\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("SET PX stores value before expiry")
}


func stage24_PXExpires(t *testing.T) {
	stage("STAGE 24: PX EXPIRY")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*5\r\n$3\r\nSET\r\n$4\r\nname\r\n$4\r\ngoat\r\n$2\r\nPX\r\n$3\r\n500\r\n")

	time.Sleep(600 * time.Millisecond)

	resp := send(conn, "*2\r\n$3\r\nGET\r\n$4\r\nname\r\n")

	expected := "$-1\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("PX expiration OK")
}


func stage25_OverwriteRemovesOldExpiry(t *testing.T) {
	stage("STAGE 25: OVERWRITE EXPIRY")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*5\r\n$3\r\nSET\r\n$4\r\nname\r\n$4\r\ngoat\r\n$2\r\nEX\r\n$1\r\n1\r\n")

	send(conn, "*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$4\r\nlion\r\n")
   
	time.Sleep(1100 * time.Millisecond)

	resp := send(conn, "*2\r\n$3\r\nGET\r\n$4\r\nname\r\n")

	expected := "$4\r\nlion\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("overwrite clears previous expiry")
}


func stage26_ConcurrentExpiry(t *testing.T) {
	stage("STAGE 26: CONCURRENT EXPIRY")

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			conn := dial(t)
			defer conn.Close()

			key := fmt.Sprintf("key-%d", id)

			req := fmt.Sprintf(
				"*5\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$5\r\nvalue\r\n$2\r\nPX\r\n$3\r\n200\r\n",
				len(key),
				key,
			)

			resp := send(conn, req)

			if resp != "+OK\r\n" {
				failf(t, "client %d expected +OK got %q", id, resp)
			}
		}(i)
	}

	wg.Wait()

	time.Sleep(300 * time.Millisecond)

	pass("concurrent expiry setup OK")
}