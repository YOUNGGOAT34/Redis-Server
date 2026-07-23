package server

import (
	"CacheDB/app/RESP"
	"CacheDB/app/storage"
	"container/list"
	"strconv"
	"time"
)

func blockClient(arguments [][]byte) RESP.Response {

	ch := make(chan []byte, 1)
	storage.BlockedClientsMutex.Lock()
	q, ok := storage.BlockedClients[string(arguments[0])]

	if !ok {
		q = list.New()
		storage.BlockedClients[string(arguments[0])] = q
	}

	element := q.PushBack(ch)

	storage.BlockedClientsMutex.Unlock()

	timeout, err := strconv.Atoi(string(arguments[1]))

	if err != nil {
		return RESP.Response{

			Body: []byte("Error timeout should be an integer"),
			Type: RESP.ERROR,
		}
	}

	if timeout == 0 {
		value := <-ch

		return RESP.Response{
			Body: RESP.EncodeArray([][]byte{arguments[0], value}),
			Type: RESP.ARRAY,
		}

	}

	select {
	case value := <-ch:
		return RESP.Response{
			Body: RESP.EncodeArray([][]byte{arguments[0], value}),
			Type: RESP.ARRAY,
		}

	case <-time.After(time.Duration(timeout) * time.Second):

		storage.BlockedClientsMutex.Lock()
		q.Remove(element)
		storage.BlockedClientsMutex.Unlock()
		return RESP.Response{
			Body: nil,
			Type: RESP.NIL,
		}

	}

}

func bLPopCommand(arguments [][]byte, client *storage.Client) RESP.Response {
	if len(arguments) != 2 {
		return RESP.WrongNumberOfArguments("BLOP")
	}

	storage.DatabaseMutex.RLock()
	data, exists := storage.Database[string(arguments[0])]
	storage.DatabaseMutex.RUnlock()

	if exists {

		if data.Type != storage.LIST {
			return RESP.WrongType()
		}

		listData := data.Value.(*storage.List)

		if listData == nil || listData.Len == 0 {
			return blockClient(arguments)

		}

		listData.ListMutex.Lock()
		value := listData.LPop()
		listData.ListMutex.Unlock()

		markDirty(string(arguments[0]), client)

		return RESP.Response{
			Body: RESP.EncodeArray([][]byte{arguments[0], value}),
			Type: RESP.ARRAY,
		}
	}

	return blockClient(arguments)
}
