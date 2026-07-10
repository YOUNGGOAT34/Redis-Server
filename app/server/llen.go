package server

import (
	"CacheDB/app/RESP"
	"strconv"
)

func llenCommand(arguments [][]byte) RESP.Response {
	if len(arguments) != 1 {
		return wrongNumberOfArguments("LLEN")
	}

	databaseMutex.RLock()
	data, exists := database[string(arguments[0])]
	databaseMutex.RUnlock()

	if exists {
		if data.Type != LIST {
			return wrongType()
		}

		list := data.Value.(*List)

		list.listMutex.RLock()
		defer list.listMutex.RUnlock()

		if list == nil {
			return RESP.Response{
				Body: []byte("0"),
				Type: RESP.INTEGER,
			}
		}

		var buf [32]byte

		return RESP.Response{
			Body: strconv.AppendInt(buf[:0], int64(list.len), 10),
			Type: RESP.INTEGER,
		}
	}

	return RESP.Response{
		Body: []byte("0"),
		Type: RESP.INTEGER,
	}

}
