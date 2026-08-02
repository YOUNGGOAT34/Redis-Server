package server

import (
	"fmt"
	"strings"

	aof "CacheDB/app/AOF"
	rdb "CacheDB/app/RDB"
	"CacheDB/app/RESP"
	"CacheDB/app/commands"
	"CacheDB/app/config"
	"CacheDB/app/replication"
	"CacheDB/app/storage"
)

func dispatchCommands(client *storage.Client, args [][]byte, replConfig *config.SERVER, rdbConfig *rdb.RDB, aofConfig *aof.AOF) RESP.Response {

	if len(args) < 1 {
		return RESP.Response{
			Body: nil,
			Type: RESP.NIL,
		}
	}

	command := args[0]

	//convert to a string and make it case insensitive so that it can be used in a switch case
	cmd := strings.ToUpper(string(command))

	if cmd == "AUTH" {
		return commands.Auth(client, args[1:])
	}

	if client.User == nil {
		return RESP.Response{
			Type: RESP.ERROR,
			Body: []byte("NOAUTH Authentication required."),
		}

	}

	if !client.User.Flags.Enabled {
		return commands.Invalid()
	}

	if client.InSubscribeMode {
		if !commands.IsLegal(cmd) {
			return RESP.Response{
				Body: fmt.Appendf(nil, "ERR Can't execute '%s': only (P|S)SUBSCRIBE / (P|S)UNSUBSCRIBE / PING / QUIT / RESET are allowed in this context", cmd),
				Type: RESP.ERROR,
			}
		}
	}

	switch cmd {

	case "MULTI":
		return multiCommand(args[1:], client)
	case "EXEC":
		return execCommand(args[1:], client, replConfig, aofConfig)

	case "DISCARD":
		return discardCommand(args[1:], client)
	case "WATCH":
		return watchCommand(args[1:], client)

	}

	if client.InTransaction {
		client.Queue = append(client.Queue,
			storage.Command{
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

		return commands.Ping(client, args[1:])

	case "SET":

		if len(args) < 2 {
			return RESP.Response{
				Body: nil,
				Type: RESP.NIL,
			}
		}
		return commands.SetCommand(args[1:], client)

	case "GET":
		return commands.GetCommand(args[1:])
	case "RPUSH":
		return commands.RPushCommand(args[1:], client)
	case "LRANGE":
		return commands.LRangeCommand(args[1:])
	case "LPUSH":
		return commands.LPushCommand(args[1:], client)
	case "LLEN":
		return commands.LlenCommand(args[1:])
	case "LPOP":
		return commands.LPopCommand(args[1:], client)
	case "BLPOP":
		return commands.BLPopCommand(args[1:], client)
	case "TYPE":
		return commands.TypeCommand(args[1:])
	case "XADD":
		return commands.XAddCommand(args[1:], client)
	case "XRANGE":
		return commands.XRangeCommand(args[1:])
	case "XREAD":
		return commands.DecideTypeOfRead(args[1:])
	case "INCR":
		return commands.IncrCommand(args[1:], client)
	case "UNWATCH":
		return unwatchCommand(args[1:], client)
	case "INFO":
		return replication.InfoCommand(args[1:], replConfig)
	case "REPLCONF":
		return replication.ReplConfig(args[1:], replConfig, client.Conn)
	case "PSYNC":
		return replication.Psync(args[:], replConfig)
	case "WAIT":
		return replication.WaitCommand(args[1:], replConfig)
	case "CONFIG":
		return commands.GetConfig(args[1:], rdbConfig, aofConfig)
	case "KEYS":
		return commands.Keys(args[1:])
	case "SAVE":
		return commands.HandleSave(args, rdbConfig)
	case "SUBSCRIBE":
		return commands.Sub(replConfig, client, args[1:])
	case "UNSUBSCRIBE":
		return commands.UnSub(replConfig, client, args[1:])
	case "PUBLISH":
		return commands.Pub(replConfig, args[1:])
	case "ACL":
		return commands.Acl(client, args[1:])
	case "ZADD":
		return commands.ZaddCommand(args[1:])
	case "ZSCORE":
		return commands.ZScore(args[1:])
	case "ZCARD":
		return commands.Zcard(args[1:])
	case "ZREM":
		return commands.ZRem(args[1:])
	default:
		return RESP.Response{
			Body: []byte("Error: Unknown command"),
			Type: RESP.ERROR,
		}

	}
}
