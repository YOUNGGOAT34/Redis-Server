package server

import (
	"CacheDB/app/RESP"
	"fmt"
)

func wrongType() RESP.Response {
	return RESP.Response{
		Body: []byte("WRONGTYPE Operation against a key holding the wrong kind of value"),
		Type: RESP.ERROR,
	}
}

func wrongNumberOfArguments(command string) RESP.Response {
	errMessage := fmt.Sprintf("Wrong number of arguments for '%s' command", command)
	return RESP.Response{
		Body: []byte(errMessage),
		Type: RESP.ERROR,
	}
}

// this lets a client know that a key in a transaction was modified
func markDirty(key string, writer *Client) {

	watchedKeysMutex.Lock()
	defer watchedKeysMutex.Unlock()

	if set, exists := watchedKeys[key]; exists {

		for client := range set {
			if client != writer {
				client.Dirty = true
			}
		}
	}
}
