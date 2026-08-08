package commands

import (
	"CacheDB/app/RESP"
	"CacheDB/app/ZSET"
	"CacheDB/app/config"
	"CacheDB/app/storage"
	"strconv"
)

func ZRem(args [][]byte,replconfig *config.SERVER) RESP.Response {

	if len(args) < 2 {
		return RESP.WrongNumberOfArguments("ZREM")
	}

	replconfig.DatabaseMutex.RLock()
	data, exists := replconfig.Database[string(args[0])]
	replconfig.DatabaseMutex.RUnlock()
	var count int64 = 0

	if exists {
		if data.Type != storage.ZSET {
			return RESP.WrongType()
		}

		zs := data.Value.(*zset.ZSet)
		zs.ZSMutex.Lock()
		defer zs.ZSMutex.Unlock()
		for _, member := range args[1:] {

			deleted := zs.ZRem(string(member))

			if deleted {
				count++
			}

		}
	}

	return RESP.Response{
		Body: []byte(strconv.FormatInt(count, 10)),
		Type: RESP.INTEGER,
	}

}
