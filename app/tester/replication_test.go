package tester

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	rdb "CacheDB/app/RDB"
	"CacheDB/app/RESP"
	"CacheDB/app/server"
)

// Runner & Stages

func replication_test(t *testing.T) {
	stageXX_ReplicaHandshake(t)
	stageXX_WritePropagation(t)
	stageXX_WaitCommandLogic(t)
	stageXX_WaitNoReplicasReturnsImmediately(t)
	stageXX_WaitTimeoutRespected(t)
	stageXX_MultipleReplicas(t)
	stageXX_OrderedPropagation(t)
	stageXX_ReplicaOffsetTracking(t)
	stageXX_PipelinedCommandsKnownBug(t)
	stageXX_PipelineReplication(t)
	stageXX_ReadCommandsAreNotPropagated(t) 
	stageXX_PartialPacketParsing(t)
	stageXX_WaitAfterPipeline(t) 
	stageXX_ReplicaDisconnectDuringPropagation(t) 
}

func stageXX_ReplicaHandshake(t *testing.T) {
	stage("REPLICATION: REPLICA HANDSHAKE")

	masterConfig, masterPort := startMaster(t)
	_, _ = startReplica(t, masterPort)

	replicaRegistered := waitForReplicaCount(masterConfig, 1, 1*time.Second)

	infoResp := sendRawCommand(t, masterPort, encodeCommand("INFO", "replication"))
	if !strings.Contains(infoResp, "role: master") {
		failf(t, "expected 'role: master' in INFO block, got: %q", infoResp)
	}
   
	if !replicaRegistered && !strings.Contains(infoResp, "connected_slaves:1") {
		failf(t, "replica never registered in master's REPLICAS slice, and INFO doesn't report connected_slaves:1")
	}

	pass("replica handshake completed successfully")
}

func stageXX_WritePropagation(t *testing.T) {
	stage("REPLICATION: WRITE PROPAGATION")

	_, masterPort := startMaster(t)
	_, replicaPort := startReplica(t, masterPort)

	setResp := sendRawCommand(t, masterPort, encodeCommand("SET", "key", "value"))
	if !strings.Contains(setResp, "OK") {
		failf(t, "SET failed on master: %q", setResp)
	}

	if !waitForValue(t, replicaPort, "key", "value", 2*time.Second) {
		failf(t, "replica failed to receive and store the propagated command from master")
	}

	pass("writes propagated to replica correctly")
}

func stageXX_WaitCommandLogic(t *testing.T) {
	stage("REPLICATION: WAIT COMMAND LOGIC")

	_, masterPort := startMaster(t)
	_, _ = startReplica(t, masterPort)

	time.Sleep(250 * time.Millisecond) // let the handshake settle

	sendRawCommand(t, masterPort, encodeCommand("INCR", "count"))
	time.Sleep(200 * time.Millisecond)

	waitResp := sendRawCommand(t, masterPort, encodeCommand("WAIT", "1", "2000"))

	if waitResp != ":1\r\n" && waitResp != ":0\r\n" {
		failf(t, "expected WAIT to return exactly ':0\\r\\n' or ':1\\r\\n', got: %q", waitResp)
	}

	pass("WAIT command logic OK")
}

func stageXX_WaitNoReplicasReturnsImmediately(t *testing.T) {
	stage("REPLICATION: WAIT WITH ZERO REPLICAS")

	_, masterPort := startMaster(t)

	start := time.Now()
	resp := sendRawCommand(t, masterPort, encodeCommand("WAIT", "1", "2000"))
	elapsed := time.Since(start)

	if resp != ":0\r\n" {
		failf(t, "expected ':0\\r\\n' with no replicas connected, got: %q", resp)
	}
	if elapsed > 200*time.Millisecond {
		failf(t, "WAIT with zero replicas should return immediately, took %v (timeout was 2000ms)", elapsed)
	}

	pass("WAIT returns immediately when no replicas are connected")
}

func stageXX_WaitTimeoutRespected(t *testing.T) {
	stage("REPLICATION: WAIT TIMEOUT RESPECTED")

	masterConfig, masterPort := startMaster(t)

	stuck := fakeReplica(t, masterPort)
	defer stuck.close()

	if !waitForReplicaCount(masterConfig, 1, 1*time.Second) {
		failf(t, "fake replica never registered in master's REPLICAS slice")
	}

	// Advance master offset so the silent replica lags behind (forcing a wait/timeout)
	sendRawCommand(t, masterPort, encodeCommand("SET", "lagkey", "lagval"))
	time.Sleep(50 * time.Millisecond)

	const timeoutMs = 300
	start := time.Now()
	resp := sendRawCommand(t, masterPort, encodeCommand("WAIT", "1", fmt.Sprintf("%d", timeoutMs)))
	elapsed := time.Since(start)

	if resp != ":0\r\n" {
		failf(t, "expected ':0\\r\\n' since the fake replica never ACKs, got: %q", resp)
	}
	if elapsed < timeoutMs*time.Millisecond {
		failf(t, "WAIT returned early (%v) despite requiring 1 replica ack that never came", elapsed)
	}
	if elapsed > (timeoutMs+250)*time.Millisecond {
		failf(t, "WAIT overran its timeout by too much: waited %v for a %dms timeout", elapsed, timeoutMs)
	}

	pass("WAIT timeout respected when replicas do not acknowledge")
}

func stageXX_MultipleReplicas(t *testing.T) {
	stage("REPLICATION: MULTIPLE REPLICAS")

	masterConfig, masterPort := startMaster(t)
	_, replicaPort1 := startReplica(t, masterPort)
	_, replicaPort2 := startReplica(t, masterPort)

	if !waitForReplicaCount(masterConfig, 2, 2*time.Second) {
		failf(t, "both replicas never registered with the master")
	}

	setResp := sendRawCommand(t, masterPort, encodeCommand("SET", "shared", "fanout"))
	if !strings.Contains(setResp, "OK") {
		failf(t, "SET failed on master: %q", setResp)
	}

	if !waitForValue(t, replicaPort1, "shared", "fanout", 2*time.Second) {
		failf(t, "replica 1 never received the propagated write")
	}
	if !waitForValue(t, replicaPort2, "shared", "fanout", 2*time.Second) {
		failf(t, "replica 2 never received the propagated write")
	}

	pass("writes propagated to multiple replicas concurrently")
}

func stageXX_OrderedPropagation(t *testing.T) {
	stage("REPLICATION: ORDERED PROPAGATION")

	_, masterPort := startMaster(t)
	_, replicaPort := startReplica(t, masterPort)

	master := dialWithReader(t, masterPort)
	defer master.close()

	for _, v := range []string{"v1", "v2", "v3"} {
		resp := master.send(t, "SET", "seq", v)
		if !strings.Contains(resp, "OK") {
			failf(t, "SET seq=%s failed: %q", v, resp)
		}
	}

	if !waitForValue(t, replicaPort, "seq", "v3", 2*time.Second) {
		failf(t, "replica never converged to the final value 'v3'")
	}

	time.Sleep(100 * time.Millisecond)
	finalResp := sendRawCommand(t, replicaPort, encodeCommand("GET", "seq"))
	if !strings.Contains(finalResp, "v3") {
		failf(t, "replica value for 'seq' regressed after settling: %q", finalResp)
	}

	pass("sequential writes propagated in strict order")
}

func stageXX_ReplicaOffsetTracking(t *testing.T) {
	stage("REPLICATION: REPLICA OFFSET TRACKING")

	masterConfig, masterPort := startMaster(t)
	_, _ = startReplica(t, masterPort)

	if !waitForReplicaCount(masterConfig, 1, 1*time.Second) {
		failf(t, "replica never registered with the master")
	}

	sendRawCommand(t, masterPort, encodeCommand("SET", "k", "v"))
	sendRawCommand(t, masterPort, encodeCommand("WAIT", "1", "2000"))

	masterConfig.ReplicasMutex.RLock()
	defer masterConfig.ReplicasMutex.RUnlock()
	if len(masterConfig.REPLICAS) == 0 {
		failf(t, "replica disappeared from REPLICAS slice")
	}
	offset := masterConfig.REPLICAS[0].Offset.Load()
	if offset <= 0 {
		failf(t, "expected replica offset to advance above 0 after SET+WAIT, got %d", offset)
	}

	pass("replica offset advanced and tracked successfully")
}

func stageXX_PipelinedCommandsKnownBug(t *testing.T) {
	stage("REPLICATION: PIPELINED COMMANDS KNOWN BUG")

	_, masterPort := startMaster(t)

	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", masterPort))
	if err != nil {
		failf(t, "failed to connect: %v", err)
	}
	defer conn.Close()

	pipelined := encodeCommand("SET", "pkey", "pval") + encodeCommand("GET", "pkey")
	if _, err := conn.Write([]byte(pipelined)); err != nil {
		failf(t, "failed to write pipelined commands: %v", err)
	}

	reader := bufio.NewReader(conn)

	firstResp, err := readRESP(reader)
	if err != nil {
		failf(t, "failed to read response to pipelined SET: %v", err)
	}
	if !strings.Contains(firstResp, "OK") {
		failf(t, "expected OK for pipelined SET, got: %q", firstResp)
	}

	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	secondResp, err := readRESP(reader)
	if err != nil {
		failf(t, "KNOWN BUG: server never responded to the second pipelined command (GET pkey) — "+
			"handleClient drops unconsumed bytes from the read buffer between loop iterations. "+
			"Underlying read error: %v", err)
	}
	if !strings.Contains(secondResp, "pval") {
		failf(t, "KNOWN BUG: expected pipelined GET to return 'pval', got: %q — "+
			"the second command in the pipeline was likely misparsed or ignored", secondResp)
	}

	pass("pipelined commands handled over same connection successfully")
}

func stageXX_PipelineReplication(t *testing.T) {
	stage("REPLICATION: PIPELINE REPLICATION")

	master, masterPort := startMaster(t)
	_ = master

	_, replicaPort := startReplica(t, masterPort)

	waitForReplicaCount(master, 1, time.Second)

	conn := dialPort(t, masterPort)
	defer conn.Close()

	send(conn,
		encodeCommand("SET", "a", "1")+
			encodeCommand("SET", "b", "2")+
			encodeCommand("SET", "c", "3"))

	replicaConn := dialPort(t, replicaPort)
	defer replicaConn.Close()

	if resp := send(replicaConn, encodeCommand("GET", "a")); resp != "$1\r\n1\r\n" {
		failf(t, "expected a=1 got %q", resp)
	}

	if resp := send(replicaConn, encodeCommand("GET", "b")); resp != "$1\r\n2\r\n" {
		failf(t, "expected b=2 got %q", resp)
	}

	if resp := send(replicaConn, encodeCommand("GET", "c")); resp != "$1\r\n3\r\n" {
		failf(t, "expected c=3 got %q", resp)
	}

	pass("multiple pipelined writes replicated")
}


func stageXX_ReadCommandsAreNotPropagated(t *testing.T) {
	stage("REPLICATION: READ COMMANDS ARE NOT PROPAGATED")

	master, masterPort := startMaster(t)
	_, replicaPort := startReplica(t, masterPort)

	if !waitForReplicaCount(master, 1, 1*time.Second) {
		failf(t, "replica never registered with master")
	}

	// SET a 1
	resp := sendRawCommand(t, masterPort, encodeCommand("SET", "a", "1"))
	if resp != "+OK\r\n" {
		failf(t, "SET a failed: got %q", resp)
	}

	time.Sleep(100 * time.Millisecond)

	offsetAfterFirstWrite := master.MASTERREPLOFFSET.Load()

	// GET a — must not be propagated.
	resp = sendRawCommand(t, masterPort, encodeCommand("GET", "a"))
	if resp != "$1\r\n1\r\n" {
		failf(t, "GET a failed: got %q", resp)
	}

	time.Sleep(100 * time.Millisecond)

	offsetAfterRead := master.MASTERREPLOFFSET.Load()

	if offsetAfterRead != offsetAfterFirstWrite {
		failf(
			t,
			"GET incorrectly changed replication offset: before=%d after=%d",
			offsetAfterFirstWrite,
			offsetAfterRead,
		)
	}

	// SET b 2 — should be propagated.
	resp = sendRawCommand(t, masterPort, encodeCommand("SET", "b", "2"))
	if resp != "+OK\r\n" {
		failf(t, "SET b failed: got %q", resp)
	}

	if !waitForValue(t, replicaPort, "a", "1", 2*time.Second) {
		failf(t, "replica never received SET a 1")
	}

	if !waitForValue(t, replicaPort, "b", "2", 2*time.Second) {
		failf(t, "replica never received SET b 2")
	}

	pass("read commands do not advance replication offset or get propagated")
}


func stageXX_PartialPacketParsing(t *testing.T) {
	stage("REPLICATION: PARTIAL PACKET PARSING")

	_, masterPort := startMaster(t)

	// Open a raw TCP connection directly to control exact write timing
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", masterPort))
	if err != nil {
		failf(t, "failed to connect to master: %v", err)
	}
	defer conn.Close()

	// Stage 1: Write the first half of the command
	// We want to send: *3\r\n$3\r\nSET\r\n$1\r\na\r\n$1\r\n1\r\n
	// We split "SET" in half -> send "*3\r\n$3\r\nS" first
	firstHalf := "*3\r\n$3\r\nS"
	if _, err := conn.Write([]byte(firstHalf)); err != nil {
		failf(t, "failed to write first half of the packet: %v", err)
	}

	// Set a small read deadline to verify the server does not respond prematurely
	err = conn.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
	if err != nil {
		failf(t, "failed to set read deadline: %v", err)
	}

	// Try reading. The server should NOT send anything yet because the command is incomplete.
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err == nil {
		failf(t, "server responded prematurely to a partial packet with %q", string(buf[:n]))
	}

	// Verify that the error was indeed a timeout (which is correct behavior)
	if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
		failf(t, "expected timeout waiting for incomplete packet response, got error: %v", err)
	}

	// Reset read deadline back to normal
	_ = conn.SetReadDeadline(time.Time{})

	// Stage 2: Write the remaining half of the command after the delay
	secondHalf := "ET\r\n$1\r\na\r\n$1\r\n1\r\n"
	if _, err := conn.Write([]byte(secondHalf)); err != nil {
		failf(t, "failed to write second half of the packet: %v", err)
	}

	// Now wait for the complete command's response
	reader := bufio.NewReader(conn)
	resp, err := readRESP(reader)
	if err != nil {
		failf(t, "failed to read response after sending complete packet: %v", err)
	}

	if resp != "+OK\r\n" {
		failf(t, "expected +OK\r\n after completing command, got %q", resp)
	}

	pass("parser correctly buffered and reconstructed the partial packet")
}



func stageXX_PartialAndPipelineParsing(t *testing.T) {
	stage("STAGE 46: PARTIAL AND PIPELINE PARSING")

	_, masterPort := startMaster(t)

	// Open raw TCP connection
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", masterPort))
	if err != nil {
		failf(t, "failed to connect to master: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	/*
	    PACKET 1: Complete "SET a 1" + Incomplete "SET b"
		The master will send to send:
		*3\r\n$3\r\nSET\r\n$1\r\na\r\n$1\r\n1\r\n  <- Complete
		*3\r\n$3\r\nSET\r\n$1\r\nb\r\n           <- Incomplete
	*/
	packet1 := "*3\r\n$3\r\nSET\r\n$1\r\na\r\n$1\r\n1\r\n*3\r\n$3\r\nSET\r\n$1\r\nb\r\n"
	if _, err := conn.Write([]byte(packet1)); err != nil {
		failf(t, "failed to write first packet: %v", err)
	}

	// The server should respond to the first complete SET a 1 command immediately.
	resp1, err := readRESP(reader)
	if err != nil {
		failf(t, "failed to read response for first complete command: %v", err)
	}
	if resp1 != "+OK\r\n" {
		failf(t, "expected +OK\r\n for SET a 1, got %q", resp1)
	}

	/*
	 set a short read deadline. 
	 The server must NOT respond to the second command because it's incomplete.
	 */
	err = conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	if err != nil {
		failf(t, "failed to set read deadline: %v", err)
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err == nil {
		failf(t, "server incorrectly responded prematurely to a partial pipeline with %q", string(buf[:n]))
	}

	if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
		failf(t, "expected timeout waiting for incomplete pipeline response, got: %v", err)
	}

	// Reset read deadline back to normal
	_ = conn.SetReadDeadline(time.Time{})

	
	/*
	    PACKET 2: Remainder of "SET b 2" + Complete "SET c 3"
	   Completes SET b with "$1\r\n2\r\n", then pipelines:
	   *3\r\n$3\r\nSET\r\n$1\r\nc\r\n$1\r\n3\r\n  <- Complete
	*/
	packet2 := "$1\r\n2\r\n*3\r\n$3\r\nSET\r\n$1\r\nc\r\n$1\r\n3\r\n"
	if _, err := conn.Write([]byte(packet2)); err != nil {
		failf(t, "failed to write second packet: %v", err)
	}

	/*
		expecting two separate "+OK\r\n" responses in order.
		One for the completed "SET b 2"
	*/
	resp2, err := readRESP(reader)
	if err != nil {
		failf(t, "failed to read response for completed SET b 2: %v", err)
	}
	if resp2 != "+OK\r\n" {
		failf(t, "expected +OK\r\n for SET b 2, got %q", resp2)
	}

	// One for the pipelined "SET c 3"
	resp3, err := readRESP(reader)
	if err != nil {
		failf(t, "failed to read response for pipelined SET c 3: %v", err)
	}
	if resp3 != "+OK\r\n" {
		failf(t, "expected +OK\r\n for SET c 3, got %q", resp3)
	}

	// Finally, verify all values were actually saved to DB
	if !waitForValue(t, masterPort, "a", "1", 1*time.Second) {
		failf(t, "master DB missing 'a' value")
	}
	if !waitForValue(t, masterPort, "b", "2", 1*time.Second) {
		failf(t, "master DB missing 'b' value")
	}
	if !waitForValue(t, masterPort, "c", "3", 1*time.Second) {
		failf(t, "master DB missing 'c' value")
	}

	pass("parser beautifully handled merged partial stream segments with pipelined commands")
}


func stageXX_WaitAfterPipeline(t *testing.T) {
	stage("REPLICATION: WAIT AFTER PIPELINE")

	master, masterPort := startMaster(t)
	_, replicaPort := startReplica(t, masterPort)

	// Wait for the replica to perform its handshake and register
	if !waitForReplicaCount(master, 1, 2*time.Second) {
		failf(t, "replica never registered with master")
	}

	// Open a single TCP connection to the master
	conn := dialPort(t, masterPort)
	defer conn.Close()

	// Pipeline three SET commands followed immediately by a WAIT 1 500
	pipeline := encodeCommand("SET", "a", "1") +
		encodeCommand("SET", "b", "2") +
		encodeCommand("SET", "c", "3") +
		encodeCommand("WAIT", "1", "500")

	if _, err := conn.Write([]byte(pipeline)); err != nil {
		failf(t, "failed to write pipelined writes and WAIT: %v", err)
	}

	reader := bufio.NewReader(conn)

	// expecting three "+OK\r\n" responses first, in order
	for _, key := range []string{"a", "b", "c"} {
		resp, err := readRESP(reader)
		if err != nil {
			failf(t, "failed to read response for SET %s: %v", key, err)
		}
		if resp != "+OK\r\n" {
			failf(t, "expected +OK\r\n for SET %s, got %q", key, resp)
		}
	}

	// Now we read the response for the pipelined WAIT 1 500
	waitResp, err := readRESP(reader)
	if err != nil {
		failf(t, "failed to read response for pipelined WAIT: %v", err)
	}

	/*
			It should return ":1\r\n" because our replica should successfully catch up
			to all three writes well within the 500ms window.
	*/
	if waitResp != ":1\r\n" {
		failf(t, "expected ':1\\r\\n' from WAIT after pipeline, got %q", waitResp)
	}

	// Sanity check: Ensure replica actually got and saved the final pipelined value
	if !waitForValue(t, replicaPort, "c", "3", 2*time.Second) {
		failf(t, "replica never synchronized the last pipelined command (SET c 3)")
	}

	pass("pipelined writes followed by WAIT executed and acknowledged sequentially")
}
func stageXX_ReplicaDisconnectDuringPropagation(t *testing.T) {
	stage("REPLICATION: REPLICA DISCONNECTS DURING PROPAGATION")

	master, masterPort := startMaster(t)

	// 1. Connect a mock replica
	replicaConn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", masterPort))
	if err != nil {
		failf(t, "failed to connect replica: %v", err)
	}

	// Handshake sequence to register as a replica
	reader := bufio.NewReader(replicaConn)

	// Step A: Send PING
	if _, err := replicaConn.Write([]byte(encodeCommand("PING"))); err != nil {
		failf(t, "failed to send PING: %v", err)
	}
	if _, err := readRESP(reader); err != nil {
		failf(t, "failed to read PING response: %v", err)
	}

	// Step B: Send REPLCONF listening-port
	if _, err := replicaConn.Write([]byte(encodeCommand("REPLCONF", "listening-port", "6380"))); err != nil {
		failf(t, "failed to send REPLCONF listening-port: %v", err)
	}
	if _, err := readRESP(reader); err != nil {
		failf(t, "failed to read REPLCONF listening-port response: %v", err)
	}

	// Step C: Send REPLCONF capa psync2
	if _, err := replicaConn.Write([]byte(encodeCommand("REPLCONF", "capa", "psync2"))); err != nil {
		failf(t, "failed to send REPLCONF capa: %v", err)
	}
	if _, err := readRESP(reader); err != nil {
		failf(t, "failed to read REPLCONF capa response: %v", err)
	}

	// Step D: Send PSYNC ? -1
	if _, err := replicaConn.Write([]byte(encodeCommand("PSYNC", "?", "-1"))); err != nil {
		failf(t, "failed to send PSYNC: %v", err)
	}
	// Master will respond with +FULLRESYNC ... and an RDB file
	if _, err := readRESP(reader); err != nil {
		failf(t, "failed to read PSYNC response: %v", err)
	}

	// Wait for the master to recognize and register the replica
	if !waitForReplicaCount(master, 1, 2*time.Second) {
		failf(t, "master never registered the replica")
	}

	// 2. Disconnect the replica abruptly
	replicaConn.Close()

	// Give the master server loop a brief moment to detect the closed connection
	time.Sleep(200 * time.Millisecond)

	// 3. Execute a write command on the master.
	// This forces the master to propagate the write. If the master hasn't cleanly
	// handled write errors to dead replica connections, it will panic here.
	resp := sendRawCommand(t, masterPort, encodeCommand("SET", "a", "1"))
	if resp != "+OK\r\n" {
		failf(t, "SET a failed after replica disconnected: got %q", resp)
	}

	// 4. Verify the replica count on the master has decreased to 0
	if !waitForReplicaCount(master, 0, 2*time.Second) {
		failf(t, "master replica count did not decrease to 0 after replica disconnected")
	}

	pass("master survived replica disconnection and cleanly removed it from replication pool")
}


// Connection & Port Helpers

func dialPort(t *testing.T, port int) net.Conn {
	t.Helper()

	conn, err := net.DialTimeout(
		"tcp",
		fmt.Sprintf("127.0.0.1:%d", port),
		time.Second,
	)
	if err != nil {
		t.Fatalf("dial %d: %v", port, err)
	}
	return conn
}

func getFreePort(t *testing.T) int {
	t.Helper()
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to resolve address: %v", err)
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		t.Fatalf("failed to bind ephemeral port: %v", err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func encodeCommand(args ...string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "*%d\r\n", len(args))
	for _, a := range args {
		fmt.Fprintf(&b, "$%d\r\n%s\r\n", len(a), a)
	}
	return b.String()
}

func readRESP(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read reply header: %w", err)
	}
	if len(line) == 0 {
		return "", fmt.Errorf("empty reply header")
	}

	switch line[0] {
	case '+', '-', ':':
		return line, nil

	case '$':
		var size int
		if _, err := fmt.Sscanf(line, "$%d\r\n", &size); err != nil {
			return "", fmt.Errorf("failed to parse bulk length from %q: %w", line, err)
		}
		if size < 0 {
			return line, nil
		}
		payload := make([]byte, size+2)
		if _, err := io.ReadFull(reader, payload); err != nil {
			return "", fmt.Errorf("failed to read bulk payload (%d bytes): %w", size, err)
		}
		return line + string(payload), nil

	case '*':
		var count int
		if _, err := fmt.Sscanf(line, "*%d\r\n", &count); err != nil {
			return "", fmt.Errorf("failed to parse array length from %q: %w", line, err)
		}
		full := line
		for i := 0; i < count; i++ {
			elem, err := readRESP(reader)
			if err != nil {
				return "", fmt.Errorf("failed to read array element %d: %w", i, err)
			}
			full += elem
		}
		return full, nil

	default:
		return "", fmt.Errorf("unrecognized RESP type byte %q in line %q", line[0], line)
	}
}

type testConn struct {
	conn   net.Conn
	reader *bufio.Reader
}

func dialWithReader(t *testing.T, port int) *testConn {
	t.Helper()
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("failed to connect to server on port %d: %v", port, err)
	}
	return &testConn{conn: conn, reader: bufio.NewReader(conn)}
}

func (c *testConn) send(t *testing.T, args ...string) string {
	t.Helper()
	if _, err := c.conn.Write([]byte(encodeCommand(args...))); err != nil {
		t.Fatalf("failed to write command %v: %v", args, err)
	}
	resp, err := readRESP(c.reader)
	if err != nil {
		t.Fatalf("failed to read response for command %v: %v", args, err)
	}
	return resp
}

func (c *testConn) close() { c.conn.Close() }

func sendRawCommand(t *testing.T, port int, rawPayload string) string {
	t.Helper()
	tc := dialWithReader(t, port)
	defer tc.close()
	if _, err := tc.conn.Write([]byte(rawPayload)); err != nil {
		t.Fatalf("failed to write raw command payload: %v", err)
	}
	resp, err := readRESP(tc.reader)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}
	return resp
}

func startMaster(t *testing.T) (*RESP.SERVER, int) {
	t.Helper()
	port := getFreePort(t)
	cfg := &RESP.SERVER{
		Role:         "master",
		PORT:         port,
		MASTERREPLID: "8371b4fb115b71c4a0413b1db346e45071511224",
		REPLICAS:     make([]*RESP.REPLICA, 0),
	}

	rdb:=&rdb.RDB{
		  Dir: ".",
		  DbFileName: "dump.rdb",
	}
	go server.StartServer(cfg,rdb)
	time.Sleep(100 * time.Millisecond)
	return cfg, port
}

func startReplica(t *testing.T, masterPort int) (*RESP.SERVER, int) {
	t.Helper()
	port := getFreePort(t)
	cfg := &RESP.SERVER{
		Role:       "slave",
		PORT:       port,
		MasterHost: "127.0.0.1",
		MasterPort: masterPort,
	}
	rdb:=&rdb.RDB{
		  Dir: ".",
		  DbFileName: "dump.rdb",
	}
	go server.StartServer(cfg,rdb)
	return cfg, port
}

func waitForReplicaCount(cfg *RESP.SERVER, n int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		cfg.ReplicasMutex.RLock()
		count := len(cfg.REPLICAS)
		cfg.ReplicasMutex.RUnlock()
		if count >= n {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return false
}

func waitForValue(t *testing.T, port int, key, expected string, timeout time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp := sendRawCommand(t, port, encodeCommand("GET", key))
		if strings.Contains(resp, expected) {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return false
}

func fakeReplica(t *testing.T, masterPort int) *testConn {
	t.Helper()
	tc := dialWithReader(t, masterPort)

	steps := []struct {
		payload      string
		expectPrefix string
	}{
		{"*1\r\n$4\r\nPING\r\n", "+PONG"},
		{"*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$5\r\n55555\r\n", "+OK"},
		{"*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n", "+OK"},
		{"*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n", "+FULLRESYNC"},
	}

	for _, step := range steps {
		if _, err := tc.conn.Write([]byte(step.payload)); err != nil {
			t.Fatalf("fake replica handshake write failed: %v", err)
		}
		resp, err := readRESP(tc.reader)
		if err != nil {
			t.Fatalf("fake replica handshake read failed: %v", err)
		}
		if !strings.HasPrefix(resp, step.expectPrefix) {
			t.Fatalf("fake replica handshake: expected prefix %q, got %q", step.expectPrefix, resp)
		}
	}

	return tc
}