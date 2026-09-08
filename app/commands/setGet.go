package commands

import (
	"CacheDB/app/RESP"
	"CacheDB/app/config"
	"CacheDB/app/storage"
	"strconv"
	"time"
)

func GetCommand(arguments [][]byte, replconfig *config.SERVER) RESP.Response {
	if len(arguments) < 1 {
		return RESP.Response{
			Body: []byte("Wrong number of arguments for 'GET' command"),
			Type: RESP.ERROR,
		}
	}

	// Lazy expiry: check under ExpiryMutex, but do NOT nest DatabaseMutex
	// inside it. Elsewhere (RDB.SaveRDB/SaveRDBLocked, and the PSYNC
	// handler's snapshotAndRegisterReplica) DatabaseMutex is locked before
	// ExpiryMutex; acquiring them in the opposite order here (as this used
	// to) is a real, reachable AB-BA deadlock between a GET on an expired
	// key and a concurrent SAVE/PSYNC. Releasing ExpiryMutex before ever
	// touching DatabaseMutex avoids the ordering question entirely instead
	// of trying to enforce a global lock order by convention.
	replconfig.ExpiryMutex.Lock()
	expires, exists := replconfig.Expiry[string(arguments[0])]
	isExpired := exists && time.Now().After(expires)
	if isExpired {
		delete(replconfig.Expiry, string(arguments[0]))
	}
	replconfig.ExpiryMutex.Unlock()

	if isExpired {
		replconfig.DatabaseMutex.Lock()
		delete(replconfig.Database, string(arguments[0]))
		replconfig.DatabaseMutex.Unlock()
	}

	replconfig.DatabaseMutex.RLock()
	dataObject, exists := replconfig.Database[string(arguments[0])]
	replconfig.DatabaseMutex.RUnlock()

	if exists {

		if dataObject.Type != storage.STRING {
			return RESP.Response{
				Body: []byte("RESP.WrongType Operation against a key holding the wrong kind of value"),
				Type: RESP.ERROR,
			}
		}

		value := dataObject.Value.([]byte)
		return RESP.Response{
			Body: value,
			Type: RESP.BULK_STRING,
		}
	}

	return RESP.Response{
		Body: nil,
		Type: RESP.NIL,
	}
}

func SetCommand(arguments [][]byte, client *storage.Client, replconfig *config.SERVER) RESP.Response {
	if len(arguments) < 2 {
		return RESP.WrongNumberOfArguments("SET")
	}

	/*

	  Arguments[0]-->key
	  Arguments[1]-->value

	*/

	replconfig.ExpiryMutex.Lock()
	delete(replconfig.Expiry, string(arguments[0]))
	replconfig.ExpiryMutex.Unlock()

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

				replconfig.ExpiryMutex.Lock()

				duration := time.Duration(timeInSeconds) * time.Second
				expiresAt := time.Now().Add(duration)
				replconfig.Expiry[string(arguments[0])] = expiresAt

				replconfig.ExpiryMutex.Unlock()

			} else if RESP.CompareBytes(arguments[2], []byte("PX")) {
				timeInMilliSeconds, err := strconv.Atoi(string(arguments[3]))
				if err != nil {
					return RESP.Response{
						Body: []byte("Error invalid expiry time"),
						Type: RESP.ERROR,
					}
				}

				replconfig.ExpiryMutex.Lock()
				duration := time.Duration(timeInMilliSeconds) * time.Millisecond
				expiresAt := time.Now().Add(duration)
				replconfig.Expiry[string(arguments[0])] = expiresAt
				replconfig.ExpiryMutex.Unlock()
			}
		} else {

			return RESP.Response{

				Body: []byte("Syntax error:Unknown option"),
				Type: RESP.ERROR,
			}

		}
	}

	replconfig.DatabaseMutex.Lock()
	replconfig.Database[string(arguments[0])] = storage.Data{
		Type:  storage.STRING,
		Value: arguments[1],
	}
	replconfig.DatabaseMutex.Unlock()

	markDirty(string(arguments[0]), client)

	return RESP.Response{
		Body: []byte("OK"),
		Type: RESP.SIMPLE_STRING,
	}

}
