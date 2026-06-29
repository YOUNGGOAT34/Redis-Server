package tester

import (
	"fmt"
	"sync"
	"testing"
)

func set_test(t * testing.T){
	 Stage10_SetBasic(t)
	 Stage11_SetOverwrite(t)
	 Stage12_SetMultipleKeys(t)
	 Stage13_SetLongValue(t)
	 Stage14_SetConcurrentClients(t)
}


func Stage10_SetBasic(t *testing.T) {
	stage("STAGE 10: SET BASIC")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn, "*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$4\r\ngoat\r\n")

	expected := "+OK\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("basic SET OK")
}

func Stage11_SetOverwrite(t *testing.T) {
	stage("STAGE 11: SET OVERWRITE")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$4\r\ngoat\r\n")

	resp := send(conn, "*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$4\r\nlion\r\n")

	expected := "+OK\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("existing key overwritten OK")
}

func Stage12_SetMultipleKeys(t *testing.T) {
	stage("STAGE 12: MULTIPLE KEYS")

	conn := dial(t)
	defer conn.Close()

	tests := []struct {
		key   string
		value string
	}{
		{"name", "goat"},
		{"lang", "golang"},
		{"editor", "vim"},
		{"db", "redis"},
	}

	for _, tc := range tests {
		req := fmt.Sprintf(
			"*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
			len(tc.key),
			tc.key,
			len(tc.value),
			tc.value,
		)

		resp := send(conn, req)

		if resp != "+OK\r\n" {
			failf(t, "SET %q failed: got %q", tc.key, resp)
		}
	}

	pass("multiple keys inserted OK")
}

func Stage13_SetLongValue(t *testing.T) {
	stage("STAGE 13: LONG VALUE")

	conn := dial(t)
	defer conn.Close()

	value := "abcdefghijklmnopqrstuvwxyz123456789"

	req := fmt.Sprintf(
		"*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$%d\r\n%s\r\n",
		len(value),
		value,
	)

	resp := send(conn, req)

	if resp != "+OK\r\n" {
		failf(t, "expected +OK got %q", resp)
	}

	pass("multi-digit value length OK")
}

func Stage14_SetConcurrentClients(t *testing.T) {
	stage("STAGE 14: CONCURRENT SET")

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			conn := dial(t)
			defer conn.Close()

			key := fmt.Sprintf("key-%d", id)
			value := fmt.Sprintf("value-%d", id)

			req := fmt.Sprintf(
				"*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
				len(key),
				key,
				len(value),
				value,
			)

			resp := send(conn, req)

			if resp != "+OK\r\n" {
				failf(t, "client %d expected +OK got %q", id, resp)
			}
		}(i)
	}

	wg.Wait()

	pass("100 concurrent SET clients OK")
}