package server

import (
	"CacheDB/app/helpers"
	"container/list"
	"strconv"
	"time"
)

func blockClient(arguments [][]byte) helpers.Response {

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
		return helpers.Response{

			Body: []byte("Error timeout should be an integer"),
			Type: helpers.ERROR,
		}
	}

	if timeout == 0 {
		value := <-ch

		return helpers.Response{
			Body: encodeArray([][]byte{arguments[0], value}),
			Type: helpers.ARRAY,
		}

	}

	select {
	case value := <-ch:
		return helpers.Response{
			Body: encodeArray([][]byte{arguments[0], value}),
			Type: helpers.ARRAY,
		}

	case <-time.After(time.Duration(timeout) * time.Second):

		blockedClientsMutex.Lock()
		q.Remove(element)
		blockedClientsMutex.Unlock()
		return helpers.Response{
			Body: nil,
			Type: helpers.NIL,
		}

	}

}

func bLPopCommand(arguments [][]byte,client *Client) helpers.Response {
	if len(arguments) != 2 {
		return  wrongNumberOfArguments("BLOP")
	}

	databaseMutex.RLock()
	data, exists := database[string(arguments[0])]
	databaseMutex.RUnlock()

	if exists {

		if data.Type != LIST {
			return  wrongType()
			}
		

		listData := data.Value.(*List)

		if listData == nil || listData.len == 0 {
			return blockClient(arguments)

		}

		
      listData.listMutex.Lock()
		value := listData.LPop()
		listData.listMutex.Unlock()

		
		markDirty(string(arguments[0]),client)

		return helpers.Response{
			Body: encodeArray([][]byte{arguments[0], value}),
			Type: helpers.ARRAY,
		}
	}

	

	return blockClient(arguments)
}
