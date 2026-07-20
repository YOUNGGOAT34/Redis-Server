package server

import (
	"strings"

	rdb "CacheDB/app/RDB"
	"CacheDB/app/RESP"
	"CacheDB/app/replication"
)

func dispatchCommands(client *Client, args [][]byte, replConfig *RESP.SERVER,rdbConfig *rdb.RDB) RESP.Response {

	if len(args) < 1 {
		return RESP.Response{
			Body: nil,
			Type: RESP.NIL,
		}
	}

	command := args[0]

	
	//convert to a string and make it case insensitive so that it can be used in a switch case
	cmd := strings.ToUpper(string(command))

	switch cmd {

	case "MULTI":
		return multiCommand(args[1:], client)
	case "EXEC":
		return execCommand(args[1:], client, replConfig)

	case "DISCARD":
		return discardCommand(args[1:], client)
	case "WATCH":
		return watchCommand(args[1:], client)

	}

	if client.InTransaction {
		client.Queue = append(client.Queue,
			Command{
				Args: args,
			})

		return RESP.Response{
			Body: []byte("QUEUED"),
			Type: RESP.SIMPLE_STRING,
		}
	}

	switch cmd {

	case "ECHO":
		if len(args) < 2 {
			return RESP.Response{
				Body: nil,
				Type: RESP.NIL,
			}
		}

		return RESP.Response{
			Body: args[1],
			Type: RESP.BULK_STRING,
		}

	case "PING":

		return RESP.Response{
			Body: []byte("PONG"),
			Type: RESP.SIMPLE_STRING,
		}

	case "SET":

		if len(args) < 2 {
			return RESP.Response{
				Body: nil,
				Type: RESP.NIL,
			}
		}
		return setCommand(args[1:], client)

	case "GET":
		return getCommand(args[1:])
	case "RPUSH":
		return rPushCommand(args[1:], client)

	case "LRANGE":
		return lRangeCommand(args[1:])
	case "LPUSH":
		return lPushCommand(args[1:], client)

	case "LLEN":
		return llenCommand(args[1:])

	case "LPOP":
		return lPopCommand(args[1:], client)

	case "BLPOP":
		return bLPopCommand(args[1:], client)
	case "TYPE":
		return typeCommand(args[1:])
	case "XADD":
		return xAddCommand(args[1:], client)
	case "XRANGE":
		return xRangeCommand(args[1:])
	case "XREAD":
		return decideTypeOfRead(args[1:])
	case "INCR":
		return incrCommand(args[1:], client)

	case "UNWATCH":
		return unwatchCommand(args[1:], client)
	case "INFO":
		return replication.InfoCommand(args[1:], replConfig)
	case "REPLCONF":
		return replication.ReplConfig(args[1:],replConfig,client.Conn)
	case "PSYNC":
		return replication.Psync(args[:],replConfig)
	case "WAIT":
		return replication.WaitCommand(args[1:],replConfig)

	case "CONFIG":
		  return rdbConfig.RdbConfig(args[1:])
	case "KEYS":
		return RESP.Response{}

	default:
		return RESP.Response{
			Body: []byte("Error: Unknown command"),
			Type: RESP.ERROR,
		}

	}
}
