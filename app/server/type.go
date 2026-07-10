package server

import "CacheDB/app/RESP"

func typeCommand(arguments [][]byte) RESP.Response {
	if len(arguments) != 1 {

		return wrongNumberOfArguments("TYPE")
	}

	databaseMutex.Lock()
	defer databaseMutex.Unlock()

	data, exists := database[string(arguments[0])]

	if exists {
		_type := typeToByte(data.Type)
		return RESP.Response{
			Body: _type,
			Type: RESP.SIMPLE_STRING,
		}
	}

	return RESP.Response{
		Body: []byte("none"),
		Type: RESP.SIMPLE_STRING,
	}

}

func typeToByte(_type TYPE) []byte {
	switch _type {
	case STRING:
		return []byte("string")

	case LIST:
		return []byte("list")

	default:
		panic("Unkown type")
	}
}
