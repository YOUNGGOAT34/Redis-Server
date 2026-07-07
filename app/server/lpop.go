package server

import (
	"fmt"
	"os"
	"strconv"
)

func encodeArray(body [][]byte) []byte {
	var respArray []byte
	respArray = fmt.Appendf(respArray, "*%d\r\n", len(body))

	for _, value := range body {
		respArray = fmt.Appendf(respArray, "$%d\r\n%s\r\n", len(value), value)
	}

	return respArray
}

func lPopCommand(arguments [][]byte,client *Client) Response {

	if len(arguments) < 1 {
		return wrongNumberOfArguments("LPOP")
	}

	databaseMutex.Lock()
	data, exists := database[string(arguments[0])]
	databaseMutex.Unlock()


	if exists {

		if data.Type != LIST {

			return  wrongType()
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

				markDirty(string(arguments[0]),client)

				return Response{
					Body: body,
					Type: BULK_STRING,
				}
			}
		} else {
			res := make([][]byte, 0)
			numberOfElements, err := strconv.Atoi(string(arguments[1]))

			if err != nil {
				fmt.Fprintf(os.Stderr, "Error converting string to integer %s\n", err.Error())
				return Response{
					Body: []byte("ERR value is not an integer, or out of range"),
					Type: ERROR,
				}
			}

			if numberOfElements < 0 {

				return Response{
					Body: []byte("ERR value is out of range, must be positive"),
					Type: ERROR,
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

			if len(res)>0{
				  	markDirty(string(arguments[0]),client)
			}

			return Response{
				Body: encodeArray(res),
				Type: ARRAY,
			}

		}

	}

	return Response{
		Body: nil,
		Type: NIL,
	}

}
