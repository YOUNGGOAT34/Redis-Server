package commands

import (
	"CacheDB/app/RESP"
	"CacheDB/app/config"
	"CacheDB/app/storage"
)

func TypeCommand(arguments [][]byte,replconfig *config.SERVER) RESP.Response {
	if len(arguments) != 1 {

		return RESP.WrongNumberOfArguments("TYPE")
	}

	replconfig.DatabaseMutex.RLock()
	defer replconfig.DatabaseMutex.RUnlock()

	data, exists := replconfig.Database[string(arguments[0])]

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

func typeToByte(_type storage.TYPE) []byte {
	switch _type {
	case storage.STRING:
		return []byte("string")

	case storage.LIST:
		return []byte("list")

	default:
		panic("Unkown type")
	}
}
