package server

import (
	"CacheDB/app/RESP"
	"CacheDB/app/storage"
	"strconv"
)

func wakeUpWaitingClients(key string, values *[][]byte) {
	storage.BlockedClientsMutex.Lock()

	for len(*values) > 0 {

		q, ok := storage.BlockedClients[key]

		if !ok {

			break
		}

		if front := q.Front(); front != nil {
			ch := front.Value.(chan []byte)
			q.Remove(front)
			if q.Len() == 0 {
				delete(storage.BlockedClients, key)
			}
			storage.BlockedClientsMutex.Unlock()
			res := (*values)[0]
			*values = (*values)[1:]
			ch <- res

		} else {
			delete(storage.BlockedClients, key)
			break
		}

		storage.BlockedClientsMutex.Lock()

	}

	storage.BlockedClientsMutex.Unlock()
}

func rPushCommand(arguments [][]byte, client *storage.Client) RESP.Response {
	if len(arguments) == 0 {
		return RESP.Response{
			Body: []byte("Wrong number of arguments for 'RPUSH' command"),
			Type: RESP.ERROR,
		}
	}

	/*

	  arguments[0]-->key
	  arguments[1:]--->value(s)

	*/

	key := arguments[0]

	//in a case where there was a key but had no values
	if len(arguments) < 2 {
		return RESP.WrongNumberOfArguments("RPUSH")
	}

	values := arguments[1:]

	storage.DatabaseMutex.Lock()
	data, exists := storage.Database[string(key)]
	storage.DatabaseMutex.Unlock()

	if exists {

		if data.Type != storage.LIST {

			return RESP.WrongType()
		}

		list := data.Value.(*storage.List)
		list.ListMutex.Lock()
		defer list.ListMutex.Unlock()

		wakeUpWaitingClients(string(arguments[0]), &values)

		for _, value := range values {
			list.PushBack(value)
		}

		if len(values) >= 1 {

			markDirty(string(arguments[0]), client)
		}

		var buf [32]byte
		return RESP.Response{

			Body: strconv.AppendInt(buf[:0], int64(list.Len), 10),
			Type: RESP.INTEGER,
		}

	}

	wakeUpWaitingClients(string(arguments[0]), &values)

	if len(values) == 0 {
		var buf [32]byte

		return RESP.Response{

			Body: strconv.AppendInt(buf[:0], 0, 10),
			Type: RESP.INTEGER,
		}
	}

	node := &storage.Node{
		Data: values[0],
	}

	list := &storage.List{
		Head: node,
		Tail: node,
		Len:  1,
	}

	for _, value := range values[1:] {
		list.PushBack(value)
	}

	storage.Database[string(key)] = storage.Data{
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
