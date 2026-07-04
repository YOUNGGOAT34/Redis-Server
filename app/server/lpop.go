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

func lPopCommand(arguments [][]byte) Response {

	if len(arguments) < 1 {
		return Response{
			Body: []byte("Error: Wrong number of arguments passed to lpop command"),
			Type: ERROR,
		}
	}

	databaseMutex.Lock()
	defer databaseMutex.Unlock()

	data, exists := database[string(arguments[0])]

	if exists {

		if data.Type != LIST {

			return Response{
				Body: []byte("WRONGTYPE Operation against a key holding the wrong kind of value"),
				Type: ERROR,
			}
		}

		list := data.Value.(*List)

		if len(arguments) == 1 {
			body := list.LPop()

			if body != nil {

				if list.len == 0 {
					delete(database, string(arguments[0]))
				}

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
