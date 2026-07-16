package server

import (
	"CacheDB/app/RESP"
	"math"
	"strconv"
)

func incrCommand(arguments [][]byte, client *Client) RESP.Response {

	if len(arguments) != 1 {
		return RESP.WrongNumberOfArguments("INCR")
	}

	key := string(arguments[0])

	databaseMutex.Lock()
	defer databaseMutex.Unlock()

	var intValue int64
	var err error

	data, exists := database[key]
	if exists {
		if data.Type != STRING {
			return RESP.WrongType()
		}

		value := data.Value.(string)

		intValue, err = strconv.ParseInt(value, 10, 64)

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

	database[key] = Data{
		Type:  STRING,
		Value: strValue,
	}

	markDirty(key, client)

	return RESP.Response{
		Body: []byte(strValue),
		Type: RESP.INTEGER,
	}
}
