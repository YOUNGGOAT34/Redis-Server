package tester

import (
	"fmt"
	"sync"
	"testing"
)

func stage15_GetBasic(t *testing.T) {
	stage("STAGE 15: GET BASIC")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$4\r\ngoat\r\n")

	resp := send(conn, "*2\r\n$3\r\nGET\r\n$4\r\nname\r\n")

	expected := "$4\r\ngoat\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("basic GET OK")
}

func stage16_GetMissingKey(t *testing.T) {
	stage("STAGE 16: GET MISSING KEY")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn, "*2\r\n$3\r\nGET\r\n$7\r\nmissing\r\n")

	expected := "$-1\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("missing key handled OK")
}


func stage17_GetOverwrite(t *testing.T) {
	stage("STAGE 17: GET AFTER OVERWRITE")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$4\r\ngoat\r\n")
	send(conn, "*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$4\r\nlion\r\n")

	resp := send(conn, "*2\r\n$3\r\nGET\r\n$4\r\nname\r\n")

	expected := "$4\r\nlion\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("overwrite reflected in GET OK")
}


func stage18_GetMultipleKeys(t *testing.T) {
	stage("STAGE 18: MULTIPLE GETS")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$4\r\ngoat\r\n")
	send(conn, "*3\r\n$3\r\nSET\r\n$3\r\nage\r\n$2\r\n20\r\n")
	send(conn, "*3\r\n$3\r\nSET\r\n$4\r\nlang\r\n$2\r\nGo\r\n")

	tests := []struct {
		key      string
		expected string
	}{
		{"name", "$4\r\ngoat\r\n"},
		{"age", "$2\r\n20\r\n"},
		{"lang", "$2\r\nGo\r\n"},
	}

	for _, tc := range tests {
		req := fmt.Sprintf("*2\r\n$3\r\nGET\r\n$%d\r\n%s\r\n", len(tc.key), tc.key)

		resp := send(conn, req)

		if resp != tc.expected {
			failf(t, "GET %q expected %q got %q", tc.key, tc.expected, resp)
		}
	}

	pass("multiple GETs OK")
}


func stage19_GetLongValue(t *testing.T) {
	stage("STAGE 19: LONG VALUE")

	conn := dial(t)
	defer conn.Close()

	value := "abcdefghijklmnopqrstuvwxyz123456789"

	req := fmt.Sprintf(
		"*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$%d\r\n%s\r\n",
		len(value),
		value,
	)

	send(conn, req)

	resp := send(conn, "*2\r\n$3\r\nGET\r\n$4\r\nname\r\n")

	expected := fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("long value retrieved OK")
}


func stage20_ConcurrentGets(t *testing.T) {
	stage("STAGE 20: CONCURRENT GET")

	conn := dial(t)

	send(conn, "*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$4\r\ngoat\r\n")
	conn.Close()

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			conn := dial(t)
			defer conn.Close()

			resp := send(conn, "*2\r\n$3\r\nGET\r\n$4\r\nname\r\n")

			expected := "$4\r\ngoat\r\n"

			if resp != expected {
				failf(t, "expected %q got %q", expected, resp)
			}
		}()
	}

	wg.Wait()

	pass("100 concurrent GET clients OK")
}