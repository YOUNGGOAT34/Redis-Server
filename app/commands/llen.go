package commands

import (
	"CacheDB/app/RESP"
	"CacheDB/app/config"
	"CacheDB/app/storage"
	"strconv"
)

func LlenCommand(arguments [][]byte,replconfig *config.SERVER) RESP.Response {
	if len(arguments) != 1 {
		return RESP.WrongNumberOfArguments("LLEN")
	}

	replconfig.DatabaseMutex.RLock()
	data, exists := replconfig.Database[string(arguments[0])]
	replconfig.DatabaseMutex.RUnlock()

	if exists {
		if data.Type != storage.LIST {
			return RESP.WrongType()
		}

		list := data.Value.(*storage.List)

		list.ListMutex.RLock()
		defer list.ListMutex.RUnlock()

		if list == nil {
			return RESP.Response{
				Body: []byte("0"),
				Type: RESP.INTEGER,
			}
		}

		var buf [32]byte

		return RESP.Response{
			Body: strconv.AppendInt(buf[:0], int64(list.Len), 10),
			Type: RESP.INTEGER,
		}
	}

	return RESP.Response{
		Body: []byte("0"),
		Type: RESP.INTEGER,
	}

}
