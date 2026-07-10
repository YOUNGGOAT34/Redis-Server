package server

import (
	"CacheDB/app/RESP"
	"container/list"
	"strconv"
	"time"
)

func blockClient(arguments [][]byte) RESP.Response {

	ch := make(chan []byte, 1)
	blockedClientsMutex.Lock()
	q, ok := blockedClients[string(arguments[0])]

	if !ok {
		q = list.New()
		blockedClients[string(arguments[0])] = q
	}

	element := q.PushBack(ch)

	blockedClientsMutex.Unlock()

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
			Body: encodeArray([][]byte{arguments[0], value}),
			Type: RESP.ARRAY,
		}

	}

	select {
	case value := <-ch:
		return RESP.Response{
			Body: encodeArray([][]byte{arguments[0], value}),
			Type: RESP.ARRAY,
		}

	case <-time.After(time.Duration(timeout) * time.Second):

		blockedClientsMutex.Lock()
		q.Remove(element)
		blockedClientsMutex.Unlock()
		return RESP.Response{
			Body: nil,
			Type: RESP.NIL,
		}

	}

}

func bLPopCommand(arguments [][]byte, client *Client) RESP.Response {
	if len(arguments) != 2 {
		return wrongNumberOfArguments("BLOP")
	}

	databaseMutex.RLock()
	data, exists := database[string(arguments[0])]
	databaseMutex.RUnlock()

	if exists {

		if data.Type != LIST {
			return wrongType()
		}

		listData := data.Value.(*List)

		if listData == nil || listData.len == 0 {
			return blockClient(arguments)

		}

		listData.listMutex.Lock()
		value := listData.LPop()
		listData.listMutex.Unlock()

		markDirty(string(arguments[0]), client)

		return RESP.Response{
			Body: encodeArray([][]byte{arguments[0], value}),
			Type: RESP.ARRAY,
		}
	}

	return blockClient(arguments)
}
