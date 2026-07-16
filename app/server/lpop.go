package server

import (
	"CacheDB/app/RESP"
	"fmt"
	"os"
	"strconv"
)

func lPopCommand(arguments [][]byte, client *Client) RESP.Response {

	if len(arguments) < 1 {
		return RESP.WrongNumberOfArguments("LPOP")
	}

	databaseMutex.Lock()
	data, exists := database[string(arguments[0])]
	databaseMutex.Unlock()

	if exists {

		if data.Type != LIST {

			return RESP.WrongType()
		}

		list := data.Value.(*List)

		list.listMutex.Lock()
		defer list.listMutex.Unlock()

		if len(arguments) == 1 {
			body := list.LPop()

			if body != nil {

				if list.len == 0 {
					delete(database, string(arguments[0]))

				}

				markDirty(string(arguments[0]), client)

				return RESP.Response{
					Body: body,
					Type: RESP.BULK_STRING,
				}
			}
		} else {
			res := make([][]byte, 0)
			numberOfElements, err := strconv.Atoi(string(arguments[1]))

			if err != nil {
				fmt.Fprintf(os.Stderr, "Error converting string to integer %s\n", err.Error())
				return RESP.Response{
					Body: []byte("ERR value is not an integer, or out of range"),
					Type: RESP.ERROR,
				}
			}

			if numberOfElements < 0 {

				return RESP.Response{
					Body: []byte("ERR value is out of range, must be positive"),
					Type: RESP.ERROR,
				}
			}

			for i := 0; i < numberOfElements; i++ {
				poppedElement := list.LPop()

				if poppedElement == nil {
					break
				}

				res = append(res, poppedElement)

			}

			if list.len == 0 {
				delete(database, string(arguments[0]))
			}

			if len(res) > 0 {
				markDirty(string(arguments[0]), client)
			}

			return RESP.Response{
				Body: RESP.EncodeArray(res),
				Type: RESP.ARRAY,
			}

		}

	}

	return RESP.Response{
		Body: nil,
		Type: RESP.NIL,
	}

}
