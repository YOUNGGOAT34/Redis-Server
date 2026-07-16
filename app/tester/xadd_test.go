package tester

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func xadd_test(t *testing.T) {
	stage68_XAddAutoID(t)
	stage69_XAddPartialID(t)
	stage70_XAddExplicitID(t)
	stage71_XAddMultipleEntries(t)
	stage72_XAddDuplicateID(t)
	stage73_XAddSmallerID(t)
	stage74_XAddWrongArguments(t)
	stage75_XAddWrongType(t)
	stage76_XAddConcurrent(t)
}

// ------------------------------------------------------------
// RESP parsing RESP
// ------------------------------------------------------------

/*
parseBulkString parses a RESP bulk string reply ("$<len>\r\n<data>\r\n")
and returns its payload.
*/
func parseBulkString(resp string) (string, bool) {
	if len(resp) == 0 || resp[0] != '$' {
		return "", false
	}
	idx := strings.Index(resp, "\r\n")
	if idx == -1 {
		return "", false
	}
	length, err := strconv.Atoi(resp[1:idx])
	if err != nil || length < 0 {
		return "", false
	}
	start := idx + 2
	if start+length > len(resp) {
		return "", false
	}
	return resp[start : start+length], true
}

// parseError parses a RESP error reply ("-<message>\r\n") and returns the message.
func parseError(resp string) (string, bool) {
	if len(resp) == 0 || resp[0] != '-' {
		return "", false
	}
	end := strings.Index(resp, "\r\n")
	if end == -1 {
		end = len(resp)
	}
	return resp[1:end], true
}

// streamID represents a parsed <ms>-<seq> stream entry ID.
type streamID struct {
	ms  uint64
	seq uint64
}

func parseStreamID(id string) (streamID, error) {
	parts := strings.SplitN(id, "-", 2)
	if len(parts) != 2 {
		return streamID{}, fmt.Errorf("malformed stream ID %q: expected <ms>-<seq>", id)
	}
	ms, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return streamID{}, fmt.Errorf("malformed stream ID %q: bad ms part: %w", id, err)
	}
	seq, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return streamID{}, fmt.Errorf("malformed stream ID %q: bad seq part: %w", id, err)
	}
	return streamID{ms: ms, seq: seq}, nil
}

func (a streamID) lessThan(b streamID) bool {
	if a.ms != b.ms {
		return a.ms < b.ms
	}
	return a.seq < b.seq
}

// mustBulkID parses a bulk-string reply as a stream ID or fails the test.
func mustBulkID(t *testing.T, resp string, context string) streamID {
	t.Helper()

	payload, ok := parseBulkString(resp)
	if !ok {
		failf(t, "%s: expected bulk string ID, got %q", context, resp)
	}

	id, err := parseStreamID(payload)
	if err != nil {
		failf(t, "%s: %v", context, err)
	}

	return id
}

/*
mustError parses an error reply and returns its message, failing the test
if the reply isn't an error, and optionally checking it contains a substring.
*/
func mustError(t *testing.T, resp string, context string, wantSubstr string) string {
	t.Helper()

	msg, ok := parseError(resp)
	if !ok {
		failf(t, "%s: expected RESP error, got %q", context, resp)
	}

	if wantSubstr != "" && !strings.Contains(strings.ToLower(msg), strings.ToLower(wantSubstr)) {
		failf(t, "%s: expected error containing %q, got %q", context, wantSubstr, msg)
	}

	return msg
}

// ------------------------------------------------------------

func stage68_XAddAutoID(t *testing.T) {
	stage("STAGE 68: XADD AUTO ID")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn,
		"*5\r\n"+
			"$4\r\nXADD\r\n"+
			"$9\r\nstream-68\r\n"+
			"$1\r\n*\r\n"+
			"$4\r\nname\r\n"+
			"$4\r\ngoat\r\n")

	id := mustBulkID(t, resp, "auto-generated ID")

	if id.ms == 0 && id.seq == 0 {
		failf(t, "auto-generated ID: expected nonzero ID, got 0-0")
	}

	pass("auto-generated ID OK")
}

func stage69_XAddPartialID(t *testing.T) {
	stage("STAGE 69: XADD PARTIAL ID")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn,
		"*5\r\n"+
			"$4\r\nXADD\r\n"+
			"$9\r\nstream-69\r\n"+
			"$6\r\n1000-*\r\n"+
			"$4\r\nname\r\n"+
			"$5\r\nfirst\r\n")

	payload, ok := parseBulkString(resp)
	if !ok {
		failf(t, "partial ID generation: expected bulk string ID, got %q", resp)
	}

	if !strings.HasPrefix(payload, "1000-") {
		failf(t, "partial ID generation: expected ID with ms prefix 1000-, got %q", payload)
	}

	id, err := parseStreamID(payload)
	if err != nil {
		failf(t, "partial ID generation: %v", err)
	}
	if id.ms != 1000 {
		failf(t, "partial ID generation: expected ms=1000, got ms=%d", id.ms)
	}

	// Second insert with the same explicit ms should get a higher seq than the first.
	resp2 := send(conn,
		"*5\r\n"+
			"$4\r\nXADD\r\n"+
			"$9\r\nstream-69\r\n"+
			"$6\r\n1000-*\r\n"+
			"$4\r\nname\r\n"+
			"$6\r\nsecond\r\n")

	id2 := mustBulkID(t, resp2, "partial ID generation (second insert)")

	if !id.lessThan(id2) {
		failf(t, "partial ID generation: expected second ID %d-%d to be greater than first ID %d-%d",
			id2.ms, id2.seq, id.ms, id.seq)
	}

	pass("partial ID generation OK")
}

func stage70_XAddExplicitID(t *testing.T) {
	stage("STAGE 70: XADD EXPLICIT ID")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn,
		"*5\r\n"+
			"$4\r\nXADD\r\n"+
			"$9\r\nstream-70\r\n"+
			"$3\r\n1-0\r\n"+
			"$4\r\nname\r\n"+
			"$4\r\ngoat\r\n")

	expected := "$3\r\n1-0\r\n"

	if resp != expected {
		failf(t, "explicit ID: expected %q got %q", expected, resp)
	}

	pass("explicit ID OK")
}

func stage71_XAddMultipleEntries(t *testing.T) {
	stage("STAGE 71: XADD MULTIPLE ENTRIES")

	conn := dial(t)
	defer conn.Close()

	var prev streamID
	for i := 0; i < 5; i++ {
		resp := send(conn,
			fmt.Sprintf(
				"*5\r\n$4\r\nXADD\r\n$9\r\nstream-71\r\n$1\r\n*\r\n$5\r\nvalue\r\n$1\r\n%d\r\n",
				i,
			),
		)

		id := mustBulkID(t, resp, fmt.Sprintf("multiple entries (insert %d)", i))

		if i > 0 && !prev.lessThan(id) {
			failf(t, "multiple entries: expected strictly increasing IDs, got %d-%d after %d-%d",
				id.ms, id.seq, prev.ms, prev.seq)
		}
		prev = id
	}

	pass("multiple entries inserted with strictly increasing IDs")
}

func stage72_XAddDuplicateID(t *testing.T) {
	stage("STAGE 72: XADD DUPLICATE ID")

	conn := dial(t)
	defer conn.Close()

	first := send(conn,
		"*5\r\n"+
			"$4\r\nXADD\r\n"+
			"$9\r\nstream-72\r\n"+
			"$3\r\n5-0\r\n"+
			"$1\r\na\r\n"+
			"$1\r\nb\r\n")

	if first != "$3\r\n5-0\r\n" {
		failf(t, "duplicate ID setup: expected $3\\r\\n5-0\\r\\n, got %q", first)
	}

	resp := send(conn,
		"*5\r\n"+
			"$4\r\nXADD\r\n"+
			"$9\r\nstream-72\r\n"+
			"$3\r\n5-0\r\n"+
			"$1\r\nx\r\n"+
			"$1\r\ny\r\n")

	mustError(t, resp, "duplicate ID", "equal or smaller")

	pass("duplicate ID rejected")
}

func stage73_XAddSmallerID(t *testing.T) {
	stage("STAGE 73: XADD SMALLER ID")

	conn := dial(t)
	defer conn.Close()

	first := send(conn,
		"*5\r\n"+
			"$4\r\nXADD\r\n"+
			"$9\r\nstream-73\r\n"+
			"$3\r\n9-9\r\n"+
			"$1\r\na\r\n"+
			"$1\r\nb\r\n")

	if first != "$3\r\n9-9\r\n" {
		failf(t, "smaller ID setup: expected $3\\r\\n9-9\\r\\n, got %q", first)
	}

	resp := send(conn,
		"*5\r\n"+
			"$4\r\nXADD\r\n"+
			"$9\r\nstream-73\r\n"+
			"$3\r\n9-8\r\n"+
			"$1\r\nx\r\n"+
			"$1\r\ny\r\n")

	mustError(t, resp, "smaller ID", "equal or smaller")

	pass("smaller ID rejected")
}

func stage74_XAddWrongArguments(t *testing.T) {
	stage("STAGE 74: XADD WRONG ARGUMENTS")

	conn := dial(t)
	defer conn.Close()

	resp := send(conn, "*1\r\n$4\r\nXADD\r\n")

	mustError(t, resp, "wrong argument count", "wrong number of arguments")

	pass("wrong argument count handled")
}

func stage75_XAddWrongType(t *testing.T) {
	stage("STAGE 75: XADD WRONG TYPE")

	conn := dial(t)
	defer conn.Close()

	setResp := send(conn,
		"*3\r\n"+
			"$3\r\nSET\r\n"+
			"$9\r\nstream-75\r\n"+
			"$5\r\nhello\r\n")

	if setResp != "+OK\r\n" {
		failf(t, "wrong type setup: expected +OK, got %q", setResp)
	}

	resp := send(conn,
		"*5\r\n"+
			"$4\r\nXADD\r\n"+
			"$9\r\nstream-75\r\n"+
			"$1\r\n*\r\n"+
			"$1\r\na\r\n"+
			"$1\r\nb\r\n")

	mustError(t, resp, "wrong type", "WRONGTYPE")

	pass("wrong type handled")
}

func stage76_XAddConcurrent(t *testing.T) {
	stage("STAGE 76: CONCURRENT XADD")

	var wg sync.WaitGroup
	var mu sync.Mutex
	ids := make(map[string]int) // ID -> client index that produced it

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			conn := dial(t)
			defer conn.Close()
         
			val:=strconv.Itoa(i)

			resp := send(conn,
				fmt.Sprintf(
					"*5\r\n$4\r\nXADD\r\n$9\r\nstream-76\r\n$1\r\n*\r\n$2\r\nid\r\n$%d\r\n%s\r\n",
					len(val),val,
				),
			)

			payload, ok := parseBulkString(resp)
			if !ok {
				failf(t, "client %d: expected bulk string ID, got %q", i, resp)
				return
			}

			if _, err := parseStreamID(payload); err != nil {
				failf(t, "client %d: %v", i, err)
				return
			}

			mu.Lock()
			if owner, exists := ids[payload]; exists {
				mu.Unlock()
				failf(t, "client %d: duplicate ID %q also produced by client %d", i, payload, owner)
				return
			}
			ids[payload] = i
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	if len(ids) != 100 {
		failf(t, "expected 100 unique IDs from concurrent XADD, got %d", len(ids))
	}

	pass("100 concurrent XADD clients OK, all IDs unique")
}