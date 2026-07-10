package server

import (
	"CacheDB/app/RESP"
	"strconv"
)

func lPushCommand(arguments [][]byte, client *Client) RESP.Response {
	if len(arguments) < 2 {
		return wrongNumberOfArguments("LPUSH")
	}

	key := string(arguments[0])

	databaseMutex.Lock()
	data, exists := database[key]
	databaseMutex.Unlock()

	if exists {

		if data.Type != LIST {

			return wrongType()
		}

		list := data.Value.(*List)
		list.listMutex.Lock()
		defer list.listMutex.Unlock()
		for _, argument := range arguments[1:] {
			list.PushFront(argument)
		}

		markDirty(string(arguments[0]), client)
		var buf [32]byte

		return RESP.Response{
			Body: strconv.AppendInt(buf[:0], int64(list.len), 10),
			Type: RESP.INTEGER,
		}

	}

	node := &Node{
		data: arguments[1],
	}

	list := &List{
		Tail: node,
		Head: node,
		len:  1,
	}

	for _, argument := range arguments[2:] {
		list.PushFront(argument)
	}

	database[key] = Data{
		Type:  LIST,
		Value: list,
	}

	markDirty(string(arguments[0]), client)

	var buf [32]byte
	return RESP.Response{

		Body: strconv.AppendInt(buf[:0], int64(list.len), 10),
		Type: RESP.INTEGER,
	}

}
