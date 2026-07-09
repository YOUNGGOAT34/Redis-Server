package server

import (
	"CacheDB/app/helpers"
	"strconv"
	"time"
)

func getCommand(arguments [][]byte) helpers.Response {
	if len(arguments) < 1 {
		return helpers.Response{
			Body: []byte("Wrong number of arguments for 'GET' command"),
			Type: helpers.ERROR,
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
			return helpers.Response{
				Body: []byte("WRONGTYPE Operation against a key holding the wrong kind of value"),
				Type: helpers.ERROR,
			}
		}

		value := dataObject.Value.(string)
		return helpers.Response{
			Body: []byte(value),
			Type: helpers.BULK_STRING,
		}
	}

	return helpers.Response{
		Body: nil,
		Type: helpers.NIL,
	}
}

func setCommand(arguments [][]byte, client *Client) helpers.Response {
	if len(arguments) < 2 {
		return wrongNumberOfArguments("SET")
	}

	/*

	  Arguments[0]-->key
	  Arguments[1]-->value

	*/

	expiryMutex.Lock()
	delete(expiry, string(arguments[0]))
	expiryMutex.Unlock()

	if len(arguments) > 2 {
		if helpers.CompareBytes(arguments[2], []byte("EX")) || helpers.CompareBytes(arguments[2], []byte("PX")) {
			if len(arguments) < 4 {
				return helpers.Response{
					Body: []byte(""),
					Type: helpers.BULK_STRING,
				}
			}

			if helpers.CompareBytes(arguments[2], []byte("EX")) {

				timeInSeconds, err := strconv.Atoi(string(arguments[3]))
				if err != nil {
					return helpers.Response{
						Body: []byte("Error invalid expiry time"),
						Type: helpers.ERROR,
					}
				}

				expiryMutex.Lock()

				duration := time.Duration(timeInSeconds) * time.Second
				expiresAt := time.Now().Add(duration)
				expiry[string(arguments[0])] = expiresAt

				expiryMutex.Unlock()

			} else if helpers.CompareBytes(arguments[2], []byte("PX")) {
				timeInMilliSeconds, err := strconv.Atoi(string(arguments[3]))
				if err != nil {
					return helpers.Response{
						Body: []byte("Error invalid expiry time"),
						Type: helpers.ERROR,
					}
				}

				expiryMutex.Lock()
				duration := time.Duration(timeInMilliSeconds) * time.Millisecond
				expiresAt := time.Now().Add(duration)
				expiry[string(arguments[0])] = expiresAt
				expiryMutex.Unlock()
			}
		} else {

			return helpers.Response{

				Body: []byte("Syntax error:Unknown option"),
				Type: helpers.ERROR,
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

	return helpers.Response{
		Body: []byte("OK"),
		Type: helpers.SIMPLE_STRING,
	}

}
