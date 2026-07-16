package server

import (
	"CacheDB/app/RESP"
	"strconv"
	"time"
)

func getCommand(arguments [][]byte) RESP.Response {
	if len(arguments) < 1 {
		return RESP.Response{
			Body: []byte("Wrong number of arguments for 'GET' command"),
			Type: RESP.ERROR,
		}
	}

	expiryMutex.Lock()

	expires, exists := expiry[string(arguments[0])]

	if exists {
		if time.Now().After(expires) {
			databaseMutex.Lock()
			delete(database, string(arguments[0]))
			databaseMutex.Unlock()
			delete(expiry, string(arguments[0]))
		}
	}

	expiryMutex.Unlock()

	databaseMutex.RLock()
	dataObject, exists := database[string(arguments[0])]
	databaseMutex.RUnlock()

	if exists {

		if dataObject.Type != STRING {
			return RESP.Response{
				Body: []byte("RESP.WrongType Operation against a key holding the wrong kind of value"),
				Type: RESP.ERROR,
			}
		}

		value := dataObject.Value.(string)
		return RESP.Response{
			Body: []byte(value),
			Type: RESP.BULK_STRING,
		}
	}

	return RESP.Response{
		Body: nil,
		Type: RESP.NIL,
	}
}

func setCommand(arguments [][]byte, client *Client) RESP.Response {
	if len(arguments) < 2 {
		return RESP.WrongNumberOfArguments("SET")
	}

	/*

	  Arguments[0]-->key
	  Arguments[1]-->value

	*/

	expiryMutex.Lock()
	delete(expiry, string(arguments[0]))
	expiryMutex.Unlock()

	if len(arguments) > 2 {
		if RESP.CompareBytes(arguments[2], []byte("EX")) || RESP.CompareBytes(arguments[2], []byte("PX")) {
			if len(arguments) < 4 {
				return RESP.Response{
					Body: []byte(""),
					Type: RESP.BULK_STRING,
				}
			}

			if RESP.CompareBytes(arguments[2], []byte("EX")) {

				timeInSeconds, err := strconv.Atoi(string(arguments[3]))
				if err != nil {
					return RESP.Response{
						Body: []byte("Error invalid expiry time"),
						Type: RESP.ERROR,
					}
				}

				expiryMutex.Lock()

				duration := time.Duration(timeInSeconds) * time.Second
				expiresAt := time.Now().Add(duration)
				expiry[string(arguments[0])] = expiresAt

				expiryMutex.Unlock()

			} else if RESP.CompareBytes(arguments[2], []byte("PX")) {
				timeInMilliSeconds, err := strconv.Atoi(string(arguments[3]))
				if err != nil {
					return RESP.Response{
						Body: []byte("Error invalid expiry time"),
						Type: RESP.ERROR,
					}
				}

				expiryMutex.Lock()
				duration := time.Duration(timeInMilliSeconds) * time.Millisecond
				expiresAt := time.Now().Add(duration)
				expiry[string(arguments[0])] = expiresAt
				expiryMutex.Unlock()
			}
		} else {

			return RESP.Response{

				Body: []byte("Syntax error:Unknown option"),
				Type: RESP.ERROR,
			}

		}
	}

	databaseMutex.Lock()
	database[string(arguments[0])] = Data{
		Type:  STRING,
		Value: string(arguments[1]),
	}
	databaseMutex.Unlock()

	markDirty(string(arguments[0]), client)

	return RESP.Response{
		Body: []byte("OK"),
		Type: RESP.SIMPLE_STRING,
	}

}
