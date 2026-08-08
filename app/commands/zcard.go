package commands

import (
	"CacheDB/app/RESP"
	zset "CacheDB/app/ZSET"
	"CacheDB/app/config"
	"CacheDB/app/storage"
	"strconv"
)

func Zcard(args [][]byte,replconfig *config.SERVER) RESP.Response {
	if len(args) != 1 {
		return RESP.WrongNumberOfArguments("ZCARD")
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
		return RESP.Response{
			Body: []byte(strconv.FormatInt(int64(zs.List.Length), 10)),
			Type: RESP.INTEGER,
		}
	}

	return RESP.Response{
		Body: []byte(strconv.FormatInt(int64(0), 10)),
		Type: RESP.INTEGER,
	}

}
