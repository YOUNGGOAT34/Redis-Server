package server

import (
	"CacheDB/app/RESP"
	"CacheDB/app/storage"
	"math"
	"strconv"
)

func incrCommand(arguments [][]byte, client *storage.Client) RESP.Response {

	if len(arguments) != 1 {
		return RESP.WrongNumberOfArguments("INCR")
	}

	key := string(arguments[0])

	storage.DatabaseMutex.Lock()
	defer storage.DatabaseMutex.Unlock()

	var intValue int64
	var err error

	data, exists := storage.Database[key]
	if exists {
		if data.Type != storage.STRING {
			return RESP.WrongType()
		}

		value := data.Value.([]byte)

		intValue, err = strconv.ParseInt(string(value), 10, 64)

		if err != nil {
			return RESP.Response{
				Body: []byte("ERR value is not an integer or out of range"),
				Type: RESP.ERROR,
			}
		} else {

			if intValue == math.MaxInt64 {
				return RESP.Response{
					Body: []byte("ERR increment or decrement would overflow"),
					Type: RESP.ERROR,
				}
			}

			intValue++
		}

	} else {

		intValue = 1

	}

	strValue := strconv.FormatInt(intValue, 10)

	storage.Database[key] = storage.Data{
		Type:  storage.STRING,
		Value: []byte(strValue),
	}

	markDirty(key, client)

	return RESP.Response{
		Body: []byte(strValue),
		Type: RESP.INTEGER,
	}
}
