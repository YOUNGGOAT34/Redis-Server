package tester

import (
	"fmt"
	"sync"
	"testing"
	"time"
)


func blpop_test(t *testing.T) {
	stage54_BLPopImmediate(t)
	stage55_BLPopBlocksUntilPush(t)
	stage56_BLPopTimeout(t)
	// stage57_BLPopMultipleKeys(t)
	stage58_BLPopWrongType(t)
	stage59_BLPopWrongArguments(t)
	stage60_ConcurrentBLPop(t)
}

func stage54_BLPopImmediate(t *testing.T) {
	stage("STAGE 54: BLPOP IMMEDIATE")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*3\r\n$5\r\nRPUSH\r\n$4\r\njobs\r\n$4\r\ntask\r\n")

	resp := send(conn, "*3\r\n$5\r\nBLPOP\r\n$4\r\njobs\r\n$1\r\n0\r\n")

	expected := "*2\r\n$4\r\njobs\r\n$4\r\ntask\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("BLPOP returns immediately when list is non-empty")
}

func stage55_BLPopBlocksUntilPush(t *testing.T) {
	stage("STAGE 55: BLPOP BLOCKS UNTIL RPUSH")

	go func() {
		time.Sleep(300 * time.Millisecond)

		conn := dial(t)
		defer conn.Close()

		send(conn, "*3\r\n$5\r\nRPUSH\r\n$9\r\nqueuelpop\r\n$5\r\nhello\r\n")
	}()

	conn := dial(t)
	defer conn.Close()

	resp := send(conn, "*3\r\n$5\r\nBLPOP\r\n$9\r\nqueuelpop\r\n$1\r\n0\r\n")

	expected := "*2\r\n$9\r\nqueuelpop\r\n$5\r\nhello\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("BLPOP unblocked by RPUSH")
}

func stage56_BLPopTimeout(t *testing.T) {
	stage("STAGE 56: BLPOP TIMEOUT")

	conn := dial(t)
	defer conn.Close()

	start := time.Now()

	resp := send(conn, "*3\r\n$5\r\nBLPOP\r\n$5\r\nghost\r\n$1\r\n1\r\n")

	if time.Since(start) < time.Second {
		failf(t, "BLPOP returned before timeout")
	}

	expected := "$-1\r\n"

	if resp != expected {
		failf(t, "expected %q got %q", expected, resp)
	}

	pass("timeout handled correctly")
}

// func stage57_BLPopMultipleKeys(t *testing.T) {
// 	stage("STAGE 57: BLPOP MULTIPLE KEYS")

// 	conn := dial(t)
// 	defer conn.Close()

// 	send(conn, "*3\r\n$5\r\nRPUSH\r\n$5\r\nlist2\r\n$1\r\nx\r\n")

// 	resp := send(conn,
// 		"*4\r\n" +
// 			"$5\r\nBLPOP\r\n" +
// 			"$5\r\nlist1\r\n" +
// 			"$5\r\nlist2\r\n" +
// 			"$1\r\n0\r\n")

// 	expected := "*2\r\n$5\r\nlist2\r\n$1\r\nx\r\n"

// 	if resp != expected {
// 		failf(t, "expected %q got %q", expected, resp)
// 	}

// 	pass("multiple-key search OK")
// }

func stage58_BLPopWrongType(t *testing.T) {
	stage("STAGE 58: BLPOP WRONG TYPE")

	conn := dial(t)
	defer conn.Close()

	send(conn, "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n")

	resp := send(conn, "*3\r\n$5\r\nBLPOP\r\n$3\r\nfoo\r\n$1\r\n0\r\n")

	if len(resp) == 0 || resp[0] != '-' {
		failf(t, "expected RESP error got %q", resp)
	}

	pass("wrong type handled")
}

func stage59_BLPopWrongArguments(t *testing.T) {
	stage("STAGE 59: BLPOP WRONG ARGUMENTS")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn, "*1\r\n$5\r\nBLPOP\r\n")

	if len(resp) == 0 || resp[0] != '-' {
		failf(t, "expected RESP error got %q", resp)
	}

	pass("wrong argument count handled")
}

func stage60_ConcurrentBLPop(t *testing.T) {
	stage("STAGE 60: CONCURRENT BLPOP")

	const clients = 20

	var wg sync.WaitGroup

	for i := 0; i < clients; i++ {
		wg.Add(1)

		go func(id int) {
					defer wg.Done()

					conn := dial(t)
					defer conn.Close()
					key:=fmt.Sprintf("queuelll%d",id)
					resp := send(conn,
						fmt.Sprintf("*3\r\n$5\r\nBLPOP\r\n$%d\r\n%s\r\n$1\r\n0\r\n",len(key),key))
						
					expected := fmt.Sprintf(
					"*2\r\n$%d\r\n%s\r\n$1\r\nx\r\n",
					len(key),
					key,
				)

				if resp != expected {
					failf(t, "client %d: expected %q got %q", id, expected, resp)
		}
		}(i)
	}

	time.Sleep(300 * time.Millisecond)
   
	for i := 0; i < clients; i++ {
		conn := dial(t)
      key:=fmt.Sprintf("queuelll%d",i)
		send(conn,
			fmt.Sprintf("*3\r\n$5\r\nRPUSH\r\n$%d\r\n%s\r\n$1\r\nx\r\n",len(key),key))

		conn.Close()
	}

	wg.Wait()

	pass("concurrent BLPOP clients OK")
}