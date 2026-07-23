package server

import (
	"CacheDB/app/RESP"
	"CacheDB/app/storage"
	"strconv"
)

func lPushCommand(arguments [][]byte, client *storage.Client) RESP.Response {
	if len(arguments) < 2 {
		return RESP.WrongNumberOfArguments("LPUSH")
	}

	key := string(arguments[0])

	storage.DatabaseMutex.Lock()
	data, exists := storage.Database[key]
	storage.DatabaseMutex.Unlock()

	if exists {

		if data.Type != storage.LIST {

			return RESP.WrongType()
		}

		list := data.Value.(*storage.List)
		list.ListMutex.Lock()
		defer list.ListMutex.Unlock()
		for _, argument := range arguments[1:] {
			list.PushFront(argument)
		}

		markDirty(string(arguments[0]), client)
		var buf [32]byte

		return RESP.Response{
			Body: strconv.AppendInt(buf[:0], int64(list.Len), 10),
			Type: RESP.INTEGER,
		}

	}

	node := &storage.Node{
		Data: arguments[1],
	}

	list := &storage.List{
		Tail: node,
		Head: node,
		Len:  1,
	}

	for _, argument := range arguments[2:] {
		list.PushFront(argument)
	}

	storage.Database[key] = storage.Data{
		Type:  storage.LIST,
		Value: list,
	}

	markDirty(string(arguments[0]), client)

	var buf [32]byte
	return RESP.Response{

		Body: strconv.AppendInt(buf[:0], int64(list.Len), 10),
		Type: RESP.INTEGER,
	}

}
