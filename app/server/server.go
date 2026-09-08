package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	aof "CacheDB/app/AOF"
	rdb "CacheDB/app/RDB"
	"CacheDB/app/RESP"
	"CacheDB/app/commands"
	"CacheDB/app/config"
	"CacheDB/app/replication"
	"CacheDB/app/storage"
)

// Replica-to-master reconnect backoff (see connectToMaster). Exposed as vars
// (not consts) so tests can shrink them for fast, deterministic retry tests.
var (
	initialReplicaRetryDelay = 500 * time.Millisecond
	maxReplicaRetryDelay     = 10 * time.Second
)

// How long shutdown waits for in-flight client commands to finish before
// proceeding to close the listener's remaining resources and exit anyway.
const clientDrainTimeout = 5 * time.Second

// for the handshake between the master and the replica
type ExpectedResponse int

const (
	ExpectPong ExpectedResponse = iota
	ExpectOK
	ExpectFullResync
)

//identify write commands

func isWrite(command []byte) bool {

	cmd := strings.ToUpper(string(command))

	switch cmd {
	case "SET", "INCR", "LPUSH", "LPOP", "RPUSH", "XADD":
		return true
	}

	return false
}

func createDefaultUser(client *storage.Client) {
	storage.UsersMutex.RLock()
	defaultUser := storage.Users["default"]
	storage.UsersMutex.RUnlock()

	// UsersMutex only protects the Users MAP itself (which entries exist).
	// The User struct's own mutable fields (Flags, Passwords, ...) are
	// protected by that user's own UserMutex - the same lock every other
	// reader/writer of these fields already uses (see commands/auth.go's
	// getUser/setUser/whoami). Reading Flags.NoPass here without it raced
	// with ACL SETUSER mutating the same fields concurrently.
	defaultUser.UserMutex.RLock()
	defer defaultUser.UserMutex.RUnlock()

	if defaultUser.Flags.NoPass {
		client.User = defaultUser
	}

}

func handleClient(conn net.Conn, replConfig *config.SERVER, rdbConfig *rdb.RDB, aofConfig *aof.AOF, clientWG *sync.WaitGroup) {
	defer clientWG.Done()

	var request []byte
	var temp = make([]byte, 1024)

	defer conn.Close()

	client := &storage.Client{
		Conn:               conn,
		KeysWatched:        make(map[string]struct{}),
		SubscribedChannels: storage.NewSet[string](),
	}

	createDefaultUser(client)

	for {

		bytesRead, err := conn.Read(temp)

		if err == io.EOF || (err != nil && strings.Contains(err.Error(), "connection reset")) {

			return

		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading client request: %s\n", err.Error())
			return
		}

		request = append(request, temp[:bytesRead]...)

		for {
			parsedRequest, bytesConsumed, err := RESP.ParseRequest(request)

			var response RESP.Response

			if err != nil {

				if errors.Is(err, RESP.ErrIncomplete) {
					break
				}

				response = RESP.Response{
					Body: []byte(err.Error()),
					Type: RESP.ERROR,
				}

			} else {

				response = dispatchCommands(client, parsedRequest, replConfig, rdbConfig, aofConfig)
			}

			commandBytes := request[:bytesConsumed]

			request = request[bytesConsumed:]

			_, err = conn.Write(RESP.EncodeResponse(response))

			if len(parsedRequest) > 0 && RESP.CompareBytes(parsedRequest[0], []byte("PSYNC")) {

				if err != nil {
					return
				}

				// CRITICAL ordering: snapshot creation and replica
				// registration must happen atomically with respect to
				// live writes, or a write landing in between would be
				// lost to this replica entirely - present in neither the
				// RDB snapshot (already taken) nor the propagated command
				// stream (registration hadn't happened yet, so
				// PropagateCommands wouldn't have sent it here either).
				// Holding DatabaseMutex+ExpiryMutex (write locks, blocking
				// ALL writers, not just readers) across both the snapshot
				// and the REPLICAS append closes that window: any write
				// that was already in flight either finished before we
				// took the locks (so it's in the snapshot) or blocks until
				// after we've registered this replica (so it's in the
				// propagated stream instead). Network I/O (reading the
				// file back and sending it) deliberately happens AFTER
				// releasing the locks, so a slow replica connection can't
				// stall the whole database.
				snapshotErr := snapshotAndRegisterReplica(rdbConfig, replConfig, conn)

				if snapshotErr != nil {
					fmt.Fprintf(os.Stderr, "Error snapshotting RDB for PSYNC: %s\r\n", snapshotErr.Error())
					return
				}

				data, err := os.ReadFile(rdbConfig.Dir + "/" + rdbConfig.DbFileName)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error reading the rdb file in the master: %s\r\n", err.Error())
					return
				}

				_, err = conn.Write(RESP.EncodeResponse(RESP.Response{
					Body: data,
					Type: RESP.RDBFILE,
				}))

				if err != nil {
					fmt.Fprintf(os.Stderr, "Error sending an rdb file to the replica: %s\r\n", err.Error())
					return
				}

				continue
			}

			if err != nil {
				return
			}

			if replConfig.Role == "master" {
				//only propagate successful write commands
				if len(parsedRequest) > 0 && isWrite(parsedRequest[0]) && response.Type != RESP.ERROR {
					if err := aofConfig.AppendCommand(commandBytes); err != nil {
						fmt.Fprintf(os.Stderr, "Error writing to AOF: %s\r\n", err.Error())
					}
					replication.PropagateCommands(commandBytes, replConfig)
					replConfig.MASTERREPLOFFSET.Add(int32(bytesConsumed))
				}
			}
		}

	}

}

// for replicas
func handleMaster(conn net.Conn, replConfig *config.SERVER, aofConfig *aof.AOF) {

	var request []byte
	temp := make([]byte, 1024)

	client := &storage.Client{}
	createDefaultUser(client)

	defer conn.Close()

	for {

		bytesRead, err := conn.Read(temp)

		if err == io.EOF || (err != nil && strings.Contains(err.Error(), "connection reset")) {
			return
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading client request: %s\n", err.Error())
			return
		}

		request = append(request, temp[:bytesRead]...)

		for {

			parsedRequest, bytesConsumed, err := RESP.ParseRequest(request)

			if err != nil {

				if errors.Is(err, RESP.ErrIncomplete) {
					break
				}

				if RESP.CompareBytes([]byte("+OK\r\n"), request) {
					request = request[5:]
					continue
				}

				fmt.Fprintf(os.Stderr, "Parse error %v\n", err)
				return

			}

			request = request[bytesConsumed:]

			response := dispatchCommands(client, parsedRequest, replConfig, &rdb.RDB{}, aofConfig)

			if len(parsedRequest) > 0 && RESP.CompareBytes(parsedRequest[0], []byte("REPLCONF")) {

				_, err = conn.Write(RESP.EncodeResponse(response))

				if err != nil {
					return
				}
			}

			replConfig.MASTERREPLOFFSET.Add(int32(bytesConsumed))
		}

	}

}

// snapshotAndRegisterReplica atomically snapshots current in-memory state to
// the RDB file and registers conn as a replica, so that no write can land in
// the gap between "snapshot taken" and "replica registered for propagation"
// - see the PSYNC handler above for the full explanation of why that gap
// would otherwise silently lose writes for a newly-connecting replica.
func snapshotAndRegisterReplica(rdbConfig *rdb.RDB, replConfig *config.SERVER, conn net.Conn) error {
	replConfig.DatabaseMutex.Lock()
	defer replConfig.DatabaseMutex.Unlock()
	replConfig.ExpiryMutex.Lock()
	defer replConfig.ExpiryMutex.Unlock()
	replConfig.ReplicasMutex.Lock()
	defer replConfig.ReplicasMutex.Unlock()

	if err := rdb.SaveRDBLocked(rdbConfig.Dir+"/"+rdbConfig.DbFileName, replConfig); err != nil {
		return err
	}

	replica := &config.REPLICA{Conn: conn}
	replica.Offset.Store(-1)
	replConfig.REPLICAS = append(replConfig.REPLICAS, replica)

	return nil
}

func accept(listener net.Listener) net.Conn {
	conn, err := listener.Accept()

	if err != nil {
		// A closed listener (intentional shutdown) is expected and not an
		// operational error worth logging - the caller checks shuttingDown
		// to distinguish this from a transient accept error.
		if !errors.Is(err, net.ErrClosed) {
			fmt.Fprintf(os.Stderr, "Error accepting connection: %s\r\n", err.Error())
		}
		return nil
	}

	return conn

}

func StartServer(replConfig *config.SERVER, rdbConfig *rdb.RDB, aofFileConfig *aof.AOF) {

	address := fmt.Sprintf("0.0.0.0:%d", replConfig.PORT)
	l, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Printf("Failed to bind to port %d\n", replConfig.PORT)
		os.Exit(1)
	}

	//load rdb file from memory
	err = rdb.LoadFileToMemory(rdbConfig, replConfig)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading an rdb file :%s\r\n", err.Error())
		return
	}

	//create append only file directory

	err = aofFileConfig.CreateAOFDir()

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error:%s\r\n", err.Error())
		return
	}

	// Listen for SIGTERM/SIGINT so `docker stop`/Ctrl-C trigger a graceful
	// shutdown instead of an immediate kill: stop accepting new clients, let
	// in-flight commands finish (bounded), close replica/master connections
	// and the AOF file, then exit. This does NOT perform an automatic RDB
	// SAVE - CacheDB's persistence model only snapshots on an explicit SAVE
	// (there is no autosave/save-points design), so shutdown preserves that
	// contract rather than silently changing when snapshots happen.
	ctx, stopSignals := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stopSignals()

	aofStopCh := make(chan struct{})
	var aofStopOnce sync.Once
	closeAOFStopCh := func() { aofStopOnce.Do(func() { close(aofStopCh) }) }

	// Background fsync for --appendfsync everysec; no-op for "always"/"no".
	// fsyncDone closes once the loop has actually stopped touching the AOF
	// file - closing aofStopCh only wakes the goroutine, it doesn't wait for
	// it, so shutdown waits on fsyncDone before calling File.Close() to rule
	// out any chance of the fsync goroutine touching the file after it's
	// closed.
	fsyncDone := aofFileConfig.StartFsyncLoop(aofStopCh)

	var clientWG sync.WaitGroup
	var shuttingDown atomic.Bool

	go func() {
		<-ctx.Done()
		fmt.Fprintln(os.Stderr, "Received shutdown signal, shutting down gracefully...")
		shuttingDown.Store(true)

		l.Close() // unblocks accept() below

		closeAOFStopCh()

		replConfig.ReplicasMutex.Lock()
		for _, replica := range replConfig.REPLICAS {
			replica.Conn.Close()
		}
		replConfig.ReplicasMutex.Unlock()

		if replConfig.MASTERCONN != nil {
			replConfig.MASTERCONN.Close()
		}
	}()

	// The default ACL user must exist before AOF replay: replay dispatches
	// each command through the same ACL-enforced path a live client uses,
	// and that path requires an authenticated client.User. Replayed commands
	// are already-committed writes reconstructing prior state (not fresh
	// input needing authorization), so they run as the trusted default user
	// rather than being rejected as NOAUTH.
	defaultUser := storage.User{
		Name:      "default",
		Passwords: make([][32]byte, 0),
		Flags: storage.UserFlags{
			NoPass:  true,
			Enabled: true,
		},
	}

	defaultUser.CommandPermissions = storage.AllCommands
	commands.GrantOrRevokePermission(&defaultUser, []byte("@ADMIN"), true)
	storage.UsersMutex.Lock()
	storage.Users["default"] = &defaultUser
	storage.UsersMutex.Unlock()

	//replay the append only file if enabled
	if aofFileConfig.AppendOnly == "yes" {

		err = replayAOF(replConfig, rdbConfig, aofFileConfig, &defaultUser)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error replaying AOF:%s\r\n", err.Error())
			return
		}
	}

	//sync with the master if this server is a replica
	if replConfig.Role == "slave" {
		conn, err := connectToMaster(replConfig, rdbConfig, ctx.Done())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Replication: %s\r\n", err.Error())
			return
		}

		go handleMaster(conn, replConfig, aofFileConfig)

	}

	for {

		conn := accept(l)
		if conn == nil {
			if shuttingDown.Load() {
				break
			}
			continue
		}
		clientWG.Add(1)
		go handleClient(conn, replConfig, rdbConfig, aofFileConfig, &clientWG)
	}

	// Let in-flight client commands finish, but don't block shutdown forever.
	drained := make(chan struct{})
	go func() {
		clientWG.Wait()
		close(drained)
	}()

	select {
	case <-drained:
	case <-time.After(clientDrainTimeout):
		fmt.Fprintln(os.Stderr, "Shutdown: timed out waiting for in-flight clients, exiting anyway")
	}

	closeAOFStopCh()

	select {
	case <-fsyncDone:
	case <-time.After(2 * time.Second):
		fmt.Fprintln(os.Stderr, "Shutdown: timed out waiting for fsync loop to stop, closing AOF file anyway")
	}

	if aofFileConfig.File != nil {
		if err := aofFileConfig.File.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing AOF file during shutdown: %s\r\n", err.Error())
		}
	}

	fmt.Fprintln(os.Stderr, "Shutdown complete.")
}

// connectToMaster performs the replica handshake against the master,
// retrying with capped exponential backoff instead of panicking if the
// master is temporarily unreachable or the handshake fails. This lets a
// replica started before (or briefly after) its master recover automatically
// once the master becomes reachable, without busy-looping or hammering it
// with connection attempts. stopSignal (e.g. a shutdown context's Done()
// channel) cancels the retry loop if the process is asked to exit before the
// master ever becomes available.
func connectToMaster(replConfig *config.SERVER, rdbConfig *rdb.RDB, stopSignal <-chan struct{}) (net.Conn, error) {
	delay := initialReplicaRetryDelay

	for {
		conn, err := syncWithMaster(replConfig, rdbConfig)
		if err == nil {
			return conn, nil
		}

		fmt.Fprintf(os.Stderr,
			"Replication: could not sync with master %s:%d (%s), retrying in %s\r\n",
			replConfig.MasterHost, replConfig.MasterPort, err.Error(), delay)

		select {
		case <-time.After(delay):
		case <-stopSignal:
			return nil, errors.New("replica shutdown requested before master became available")
		}

		delay *= 2
		if delay > maxReplicaRetryDelay {
			delay = maxReplicaRetryDelay
		}
	}
}

func handShake(message string, conn net.Conn, RES ExpectedResponse) error {
	response := make([]byte, 128)

	_, err := conn.Write([]byte(message))

	if err != nil {
		return err
	}

	n, err := conn.Read(response)

	if err != nil {
		return err
	}

	switch RES {
	case ExpectPong:

		if string(response[:n]) != "+PONG\r\n" {
			return errors.New("Unexpected Response from the master\n")
		}

	case ExpectOK:
		if string(response[:n]) != "+OK\r\n" {
			return errors.New("Unexpected Response from the master\n")
		}
	}

	return nil
}

func receiveFullResync(message string, conn net.Conn) ([]byte, error) {
	temp := make([]byte, 1024)
	var buffer []byte

	_, err := conn.Write([]byte(message))

	if err != nil {
		return nil, err
	}

	firstNewLinePos := 0

	//read the fullresync line

	for {

		n, err := conn.Read(temp)

		if err != nil {
			return nil, err
		}

		buffer = append(buffer, temp[:n]...)

		firstNewLinePos = bytes.Index(buffer[:n], []byte("\r\n"))

		if firstNewLinePos == -1 {

			continue
		}

		fullyResync := buffer[:firstNewLinePos]

		if !strings.HasPrefix(string(fullyResync), "+FULLRESYNC") {
			return nil, errors.New("Unexpected Response from the master\n")
		}

		buffer = buffer[firstNewLinePos+2:]

		break
	}

	//read the rdb length
	secondNewLine := 0
	rdbLength := 0
	for {
		secondNewLine = bytes.Index(buffer, []byte("\r\n"))

		if secondNewLine != -1 {
			lengthLine := buffer[:secondNewLine]
			//skip the header(+2 for the crlf)
			buffer = buffer[secondNewLine+2:]

			if len(lengthLine) == 0 || lengthLine[0] != '$' {
				return nil, errors.New("Expected RDB length")
			}

			rdbLength, err = strconv.Atoi(string(lengthLine[1:]))

			if err != nil {
				return nil, err
			}

			break

		}

		n, err := conn.Read(temp)

		if err != nil {
			return nil, err
		}

		buffer = append(buffer, temp[:n]...)
	}

	rdbData := make([]byte, rdbLength)
	/*
	   There might be some rdb files in the buffer ,so copy them and get the number of bytes we copied
	*/
	copiedBytes := copy(rdbData, buffer)
	buffer = buffer[copiedBytes:]

	for copiedBytes < rdbLength {
		n, err := conn.Read(temp)

		if err != nil {
			return nil, err
		}

		bytesNeeded := rdbLength - copiedBytes

		if n <= bytesNeeded {
			copy(rdbData[copiedBytes:], temp[:n])
			copiedBytes += n
		} else {

			copy(rdbData[copiedBytes:], temp[:bytesNeeded])
			buffer = append(buffer, temp[bytesNeeded:n]...)
			copiedBytes = bytesNeeded

		}
	}

	return rdbData, nil

}

func syncWithMaster(replConfig *config.SERVER, rdbConfig *rdb.RDB) (net.Conn, error) {
	address := net.JoinHostPort(replConfig.MasterHost, fmt.Sprintf("%d", replConfig.MasterPort))
	conn, err := net.Dial("tcp", address)

	if err != nil {
		return nil, err
	}

	replConfig.MASTERCONN = conn

	message := "*1\r\n$4\r\nPING\r\n"
	err = handShake(message, conn, ExpectPong)

	if err != nil {
		return nil, err
	}

	port := fmt.Sprintf("%d", replConfig.PORT)
	message = fmt.Sprintf("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$%d\r\n%s\r\n", len(port), port)
	err = handShake(message, conn, ExpectOK)

	if err != nil {
		return nil, err
	}

	message = "*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n"

	err = handShake(message, conn, ExpectOK)

	if err != nil {
		return nil, err
	}

	message = "*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n"
	rdbBytes, err := receiveFullResync(message, conn)

	if err != nil {
		return nil, err
	}

	path := rdbConfig.Dir + "/" + rdbConfig.DbFileName

	_, err = os.Stat(path)

	if os.IsNotExist(err) {
		err = os.WriteFile(path, rdbBytes, 0644)

		if err != nil {

			return nil, err
		}

	} else {

		err = os.WriteFile(path, rdbBytes, 0644)

	}

	if err != nil {
		return nil, err
	}

	// Writing the transferred bytes to disk is not enough on its own: the
	// replica's in-memory Database/Expiry maps must actually be populated
	// from this snapshot, or a fresh replica will never see any data that
	// existed on the master before this PSYNC (it would only start
	// reflecting writes streamed AFTER this handshake). This mirrors the
	// same load step StartServer already performs at boot from a local RDB
	// file.
	if err := rdb.LoadFileToMemory(rdbConfig, replConfig); err != nil {
		return nil, err
	}

	return conn, nil
}
