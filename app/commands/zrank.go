package commands

import (
	"CacheDB/app/RESP"
	"CacheDB/app/ZSET"
	"CacheDB/app/config"
	"CacheDB/app/storage"
	"strconv"
)

func Zrank(args [][]byte,replconfig *config.SERVER) RESP.Response {
	if len(args) != 2 {
		return RESP.WrongNumberOfArguments("ZRANK")
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
		node, exists := zs.Dict[string(args[1])]
		if exists {
			target, rank := zs.List.Search(node)
			if target != nil {

				return RESP.Response{
					Body: []byte(strconv.FormatInt(int64(rank), 10)),
					Type: RESP.INTEGER,
				}
			} else {
				return RESP.Response{
					Body: []byte("internal sorted set corruption"),
					Type: RESP.ERROR,
				}
			}
		}
	}

	return RESP.Response{
		Body: []byte{},
		Type: RESP.NIL,
	}
}
