package commands

import (
	"CacheDB/app/RESP"
	zset "CacheDB/app/ZSET"
	"CacheDB/app/storage"
	"strconv"
)

func ZRem(args [][]byte) RESP.Response {

	if len(args) < 2 {
		return RESP.WrongNumberOfArguments("ZREM")
	}

	storage.DatabaseMutex.RLock()
	data, exists := storage.Database[string(args[0])]
	storage.DatabaseMutex.RUnlock()
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
