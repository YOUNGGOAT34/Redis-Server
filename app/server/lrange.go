package server

import (
	"CacheDB/app/RESP"
	"CacheDB/app/storage"
	"fmt"
	"os"
	"strconv"
)

func encodeRespArray(list *storage.List, startIndex int, endIndex int) []byte {

	var respArray []byte

	count := endIndex - startIndex + 1

	respArray = fmt.Appendf(respArray, "*%d\r\n", count)

	/*
		     To find the starting node
			  A small optimization is:
			  if the distance between the head and startIndex is less than
			  the distance between Tail and startIndex
			  find the startIndex by traversing the list from head
			  otherwise find the start index by traversing from the tail backwards since it is a doubly linked list
	*/

	currentNode := list.Head
	currentIndex := 0

	if startIndex <= (list.Len-1)-startIndex {

		for currentNode != nil && currentIndex != startIndex {
			currentNode = currentNode.Next
			currentIndex += 1
		}

	} else {

		currentIndex = list.Len - 1
		currentNode = list.Tail
		for currentNode != nil && currentIndex != startIndex {
			currentNode = currentNode.Prev
			currentIndex--
		}

	}

	for i := startIndex; i <= endIndex; i++ {
		element := currentNode.Data

		respArray = fmt.Appendf(respArray, "$%d\r\n%s\r\n", len(element), element)
		currentNode = currentNode.Next
	}
	return respArray
}

func lRangeCommand(arguments [][]byte) RESP.Response {

	if len(arguments) != 3 {
		return RESP.WrongNumberOfArguments("LRANGE")

	}

	key := string(arguments[0])

	startIndex, err := strconv.Atoi(string(arguments[1]))
	if err != nil {

		fmt.Fprintf(os.Stderr, "Error converting string to integer %s\n", err.Error())
		return RESP.Response{
			Body: []byte("Invalid start index"),
			Type: RESP.ERROR,
		}
	}

	endIndex, err := strconv.Atoi(string(arguments[2]))

	if err != nil {

		fmt.Fprintf(os.Stderr, "Error converting string to integer %s\n", err.Error())

		return RESP.Response{
			Body: []byte("Invalid end index"),
			Type: RESP.ERROR,
		}
	}

	storage.DatabaseMutex.RLock()
	data, exists := storage.Database[key]
	storage.DatabaseMutex.RUnlock()

	if exists {

		if data.Type != storage.LIST {
			return RESP.WrongType()
		}

		list := data.Value.(*storage.List)

		list.ListMutex.RLock()
		defer list.ListMutex.RUnlock()

		if list == nil || list.Len == 0 {
			return RESP.Response{
				Body: []byte("*0\r\n"),
				Type: RESP.ARRAY,
			}
		}

		if startIndex < 0 {
			startIndex = list.Len + startIndex
		}

		if startIndex >= list.Len {

			return RESP.Response{
				Body: []byte("*0\r\n"),
				Type: RESP.ARRAY,
			}

		}

		if endIndex >= list.Len {
			endIndex = list.Len - 1
		}

		if endIndex < 0 {
			endIndex = list.Len + endIndex
		}

		if startIndex < 0 {
			startIndex = 0
		}

		if endIndex < 0 {
			endIndex = 0
		}

		if startIndex > endIndex {
			return RESP.Response{
				Body: []byte("*0\r\n"),
				Type: RESP.ARRAY,
			}
		}

		return RESP.Response{

			Body: encodeRespArray(list, startIndex, endIndex),
			Type: RESP.ARRAY,
		}

	}

	return RESP.Response{
		Body: []byte("*0\r\n"),
		Type: RESP.ARRAY,
	}
}
