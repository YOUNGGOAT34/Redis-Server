package replication

import (
	"CacheDB/app/RESP"
	"CacheDB/app/config"
	"strconv"
	"time"
)

func WaitCommand(args [][]byte, serverConfig *config.SERVER) RESP.Response {

	if len(args) < 2 {

		return RESP.Response{
			Body: []byte("Wrong number of arguments for 'WAIT' command"),
			Type: RESP.ERROR,
		}
	}

	targetOffset := serverConfig.MASTERREPLOFFSET.Load()
	ack := RESP.EncodeResponse(RESP.Response{
		Array: []RESP.Response{
			{Type:RESP.BULK_STRING,Body:[]byte("REPLCONF")},
			{Type:RESP.BULK_STRING,Body:[]byte("GETACK")},
			{Type:RESP.BULK_STRING,Body:[]byte("*")},
		},
		Type: RESP.ARRAY,
	})

	serverConfig.ReplicasMutex.RLock()

	replicas := append([]*config.REPLICA(nil), serverConfig.REPLICAS...)

	serverConfig.ReplicasMutex.RUnlock()

	for _, replica := range replicas {
		_, err := replica.Conn.Write(ack)

		if err != nil {
			//if the write fails remove the replica

			serverConfig.ReplicasMutex.Lock()
			for j, r := range serverConfig.REPLICAS {
				if r == replica {
					serverConfig.REPLICAS[j].Conn.Close()
					serverConfig.REPLICAS = append(serverConfig.REPLICAS[:j], serverConfig.REPLICAS[j+1:]...)
					break
				}

			}
			serverConfig.ReplicasMutex.Unlock()

		}

	}

	requiredReplicas, err := strconv.Atoi(string(args[0]))

	if err != nil {
		return RESP.Response{
			Body: []byte(err.Error()),
			Type: RESP.ERROR,
		}
	}

	serverConfig.ReplicasMutex.RLock()
	if requiredReplicas == 0 || len(serverConfig.REPLICAS) == 0 {
		serverConfig.ReplicasMutex.RUnlock()
		return RESP.Response{
			Body: []byte("0"),
			Type: RESP.INTEGER,
		}
	}

	if requiredReplicas > len(serverConfig.REPLICAS) {
		requiredReplicas = len(serverConfig.REPLICAS)
	}

	serverConfig.ReplicasMutex.RUnlock()

	timeout, err := strconv.Atoi(string(args[1]))

	if err != nil {
		return RESP.Response{
			Body: []byte(err.Error()),
			Type: RESP.ERROR,
		}
	}

	deadline := time.Now().Add(time.Duration(timeout) * time.Millisecond)

	var totalCount int64

	for {

		serverConfig.ReplicasMutex.RLock()

		replicas := append([]*config.REPLICA(nil), serverConfig.REPLICAS...)

		serverConfig.ReplicasMutex.RUnlock()

		count := 0

		for _, replica := range replicas {
			if int32(replica.Offset.Load()) >= targetOffset {
				count++
			}
		}

		if count >= requiredReplicas {
			totalCount = int64(count)
			break
		}

		if time.Now().After(deadline) {
			totalCount = int64(count)
			break
		}

		//sleep for a millisecond to avoid busy spinning
		time.Sleep(time.Millisecond)
	}

	return RESP.Response{
		Body: []byte(strconv.FormatInt(totalCount, 10)),
		Type: RESP.INTEGER,
	}
}
