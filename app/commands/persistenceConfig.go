package commands

import (
	aof "CacheDB/app/AOF"
	rdb "CacheDB/app/RDB"
	"CacheDB/app/RESP"
)

func GetConfig(args [][]byte, rdbConfig *rdb.RDB, aofConfig *aof.AOF) RESP.Response {
	if len(args) < 2 {
		return RESP.WrongNumberOfArguments("CONFIG")
	}

	if RESP.CompareBytes([]byte("GET"), args[0]) {
		switch string(args[1]) {
		case "dir":
			return RESP.Response{
				Array: []RESP.Response{
					{Body: args[1], Type: RESP.BULK_STRING},
					{Body: []byte(rdbConfig.Dir), Type: RESP.BULK_STRING},
				},
				Type: RESP.ARRAY,
			}
		case "dbfilename":
			return RESP.Response{
				Array: []RESP.Response{
					{Body: args[1], Type: RESP.BULK_STRING},
					{Body: []byte(rdbConfig.DbFileName), Type: RESP.BULK_STRING},
				},
				Type: RESP.ARRAY,
			}
		case "appendonly":

			return RESP.Response{
				Array: []RESP.Response{
					{Body: args[1], Type: RESP.BULK_STRING},
					{Body: []byte(aofConfig.AppendOnly), Type: RESP.BULK_STRING},
				},
				Type: RESP.ARRAY,
			}

		case "appenddirname":
			return RESP.Response{
				Array: []RESP.Response{
					{Body: args[1], Type: RESP.BULK_STRING},
					{Body: []byte(aofConfig.AppendDirName), Type: RESP.BULK_STRING},
				},
				Type: RESP.ARRAY,
			}

		case "appendfilename":
			return RESP.Response{
				Array: []RESP.Response{
					{Body: args[1], Type: RESP.BULK_STRING},
					{Body: []byte(aofConfig.AppendFilename), Type: RESP.BULK_STRING},
				},
				Type: RESP.ARRAY,
			}

		case "appendfsync":
			return RESP.Response{
				Array: []RESP.Response{
					{Body: args[1], Type: RESP.BULK_STRING},
					{Body: []byte(aofConfig.AppendFsync), Type: RESP.BULK_STRING},
				},
				Type: RESP.ARRAY,
			}

		default:
			return RESP.Response{
				Body: []byte("Unknown configuration"),
				Type: RESP.ERROR,
			}

		}

	}

	return RESP.Response{}
}
