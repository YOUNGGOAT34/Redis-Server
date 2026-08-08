package commands

import (
	"CacheDB/app/RESP"
	zset "CacheDB/app/ZSET"
	"CacheDB/app/config"
	"CacheDB/app/storage"
	"strconv"
)

func ZScore(args [][]byte,replconfig *config.SERVER) RESP.Response {
	if len(args) != 2 {
		return RESP.WrongNumberOfArguments("ZSCORE")
	}

	replconfig.DatabaseMutex.RLock()
	data, exists := replconfig.Database[string(args[0])]
	replconfig.DatabaseMutex.RUnlock()

	if exists {
		if data.Type != storage.ZSET {
			return RESP.WrongType()
		}

		zs := data.Value.(*zset.ZSet)
		zs.ZSMutex.RLock()
		defer zs.ZSMutex.RUnlock()

		node := zs.ZScore(string(args[1]))
		if node != nil {
			return RESP.Response{
				Body: []byte(strconv.FormatFloat(node.Score, 'g', -1, 64)),
				Type: RESP.BULK_STRING,
			}
		}
	}

	return RESP.Response{
		Body: []byte{},
		Type: RESP.NIL,
	}
}
