package tester

import (
	"fmt"
	"sync"
	"testing"
)


func type_test(t *testing.T) {
	stage61_TypeString(t)
	stage62_TypeList(t)
	stage63_TypeMissingKey(t)
	// stage64_TypeAfterDelete(t)
	stage65_TypeWrongArguments(t)
	stage67_TypeConcurrentAccess(t)
}

// ---------------- TYPE TESTS ----------------

func stage61_TypeString(t *testing.T) {
	stage("STAGE 61: TYPE STRING")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n")

	resp := send(conn, "*2\r\n$4\r\nTYPE\r\n$3\r\nfoo\r\n")

	expected := "+string\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("TYPE string OK")
}

func stage62_TypeList(t *testing.T) {
	stage("STAGE 62: TYPE LIST")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*3\r\n$5\r\nRPUSH\r\n$4\r\nlist\r\n$1\r\na\r\n")

	resp := send(conn, "*2\r\n$4\r\nTYPE\r\n$4\r\nlist\r\n")

	expected := "+list\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("TYPE list OK")
}

func stage63_TypeMissingKey(t *testing.T) {
	stage("STAGE 63: TYPE MISSING KEY")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn, "*2\r\n$4\r\nTYPE\r\n$7\r\nmissing\r\n")

	expected := "+none\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("TYPE missing key OK")
}

// func stage64_TypeAfterDelete(t *testing.T) {
// 	stage("STAGE 64: TYPE AFTER DELETE")

// 	conn := dial(t)
// 	defer conn.Close()

// 	send(conn, "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n")

// 	// assuming DEL implemented later; if not, adapt/remove
// 	send(conn, "*2\r\n$3\r\nDEL\r\n$3\r\nfoo\r\n")

// 	resp := send(conn, "*2\r\n$4\r\nTYPE\r\n$3\r\nfoo\r\n")

// 	expected := "+none\r\n"

// 	if resp != expected {
// 		failf(t, "expected %q got %q", expected, resp)
// 	}

// 	pass("TYPE after delete OK")
// }

func stage65_TypeWrongArguments(t *testing.T) {
	stage("STAGE 65: TYPE WRONG ARGUMENTS")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn, "*1\r\n$4\r\nTYPE\r\n")

	if len(resp) == 0 || resp[0] != '-' {
		failf(t, "expected error got %q", resp)
	}

	pass("TYPE wrong args handled")
}


func stage67_TypeConcurrentAccess(t *testing.T) {
	stage("STAGE 67: TYPE CONCURRENT")

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			conn := dial(t)
			defer conn.Close()

			key := fmt.Sprintf("k%d", i)

			send(conn, fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$3\r\nval\r\n", len(key), key))

			resp := send(conn, fmt.Sprintf("*2\r\n$4\r\nTYPE\r\n$%d\r\n%s\r\n", len(key), key))

			if len(resp) == 0 || resp[0] != '+' {
				failf(t, "client %d invalid TYPE response %q", i, resp)
			}
		}(i)
	}

	wg.Wait()

	pass("TYPE concurrent OK")
}