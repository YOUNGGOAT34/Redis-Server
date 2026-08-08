package commands

import (
	"CacheDB/app/RESP"
	"CacheDB/app/config"
	"CacheDB/app/storage"
	"fmt"
	"os"
	"strconv"
)

func LPopCommand(arguments [][]byte, client *storage.Client,replconfig *config.SERVER) RESP.Response {

	if len(arguments) < 1 {
		return RESP.WrongNumberOfArguments("LPOP")
	}

	replconfig.DatabaseMutex.Lock()
	data, exists := replconfig.Database[string(arguments[0])]
	replconfig.DatabaseMutex.Unlock()

	if exists {

		if data.Type != storage.LIST {

			return RESP.WrongType()
		}

		list := data.Value.(*storage.List)

		list.ListMutex.Lock()
		defer list.ListMutex.Unlock()

		if len(arguments) == 1 {
			body := list.LPop()

			if body != nil {

				if list.Len == 0 {
					delete(replconfig.Database, string(arguments[0]))

				}

				markDirty(string(arguments[0]), client)

				return RESP.Response{
					Body: body,
					Type: RESP.BULK_STRING,
				}
			}
		} else {
			res := make([]RESP.Response, 0)
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

				res = append(res, RESP.Response{
					Body: poppedElement,
					Type: RESP.BULK_STRING,
				})

			}

			if list.Len == 0 {
				delete(replconfig.Database, string(arguments[0]))
			}

			if len(res) > 0 {
				markDirty(string(arguments[0]), client)
			}

			return RESP.Response{
				Array: res,
				Type:  RESP.ARRAY,
			}

		}

	}

	return RESP.Response{
		Body: nil,
		Type: RESP.NIL,
	}

}
