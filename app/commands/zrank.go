package commands

import (
	"CacheDB/app/RESP"
	zset "CacheDB/app/ZSET"
	"CacheDB/app/storage"
	"strconv"
)

func Zrank(args [][]byte) RESP.Response {
	if len(args) != 2 {
		return RESP.WrongNumberOfArguments("ZRANK")
	}

	data, exists := storage.Database[string(args[0])]

	if exists {
		if data.Type != storage.ZSET {
			return RESP.WrongType()
		}

		zs := data.Value.(*zset.ZSet)
       
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
