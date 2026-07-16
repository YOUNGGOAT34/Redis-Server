package server

import (
	rdb "CacheDB/app/RDB"
	"CacheDB/app/RESP"
	"fmt"
)

//     |----------------------MULTI COMMAND----------------------|

func multiCommand(arguments [][]byte, client *Client) RESP.Response {
	if len(arguments) != 0 {
		return RESP.WrongNumberOfArguments("MULTI")
	}

	if client.InTransaction {
		return RESP.Response{
			Body: []byte("ERR MULTI calls cannot be nested"),
			Type: RESP.ERROR,
		}
	}

	client.InTransaction = true

	return RESP.Response{
		Body: []byte("OK"),
		Type: RESP.SIMPLE_STRING,
	}
}

//     |----------------------EXEC COMMAND----------------------|

func execCommand(arguments [][]byte, client *Client, replConfig *RESP.SERVER) RESP.Response {

	if len(arguments) != 0 {
		return RESP.WrongNumberOfArguments("EXEC")
	}

	//exec executed without multi
	if !client.InTransaction {
		return RESP.Response{
			Body: []byte("ERR EXEC without MULTI"),
			Type: RESP.ERROR,
		}
	}

	queued := client.Queue

	client.Queue = nil
	client.InTransaction = false
	defer clearWatches(client)

	if client.Dirty {

		return RESP.Response{
			Body: []byte("*-1\r\n"),
			Type: RESP.ARRAY,
		}
	}

	var resp []byte
	resp = fmt.Appendf(resp, "*%d\r\n", len(queued))

	for _, cmd := range queued {
		r := dispatchCommands(client, cmd.Args, replConfig, &rdb.RDB{})
		resp = append(resp, RESP.EncodeResponse(r)...)
	}

	return RESP.Response{
		Body: resp,
		Type: RESP.ARRAY,
	}
}

//     |----------------------DISCARD COMMAND----------------------|

func discardCommand(arguments [][]byte, client *Client) RESP.Response {
	if len(arguments) != 0 {
		return RESP.WrongNumberOfArguments("DISCARD")
	}

	if !client.InTransaction {
		return RESP.Response{
			Body: []byte("ERR DISCARD without MULTI"),
			Type: RESP.ERROR,
		}
	}

	client.InTransaction = false
	client.Queue = nil

	clearWatches(client)
	return RESP.Response{
		Body: []byte("OK"),
		Type: RESP.SIMPLE_STRING,
	}
}

//     |----------------------WATCH COMMAND----------------------|

func watchCommand(arguments [][]byte, client *Client) RESP.Response {
	if len(arguments) < 1 {
		return RESP.WrongNumberOfArguments("WATCH")
	}

	if client.InTransaction {
		return RESP.Response{
			Body: []byte("ERR WATCH inside MULTI is not allowed"),
			Type: RESP.ERROR,
		}
	}

	for _, argument := range arguments {

		key := string(argument)

		watchedKeysMutex.Lock()
		set, exists := watchedKeys[key]

		if !exists {
			set = make(map[*Client]struct{})
			watchedKeys[key] = set

		}

		set[client] = struct{}{}

		watchedKeysMutex.Unlock()

		client.keysWatched[key] = struct{}{}
	}

	return RESP.Response{
		Body: []byte("OK"),
		Type: RESP.SIMPLE_STRING,
	}
}

func unwatchCommand(arguments [][]byte, client *Client) RESP.Response {
	if len(arguments) != 0 {
		return RESP.WrongNumberOfArguments("UNWATCH")
	}

	clearWatches(client)

	return RESP.Response{
		Body: []byte("OK"),
		Type: RESP.SIMPLE_STRING,
	}
}

func clearWatches(client *Client) {
	watchedKeysMutex.Lock()

	for key := range client.keysWatched {
		delete(watchedKeys[key], client)
		if len(watchedKeys[key]) == 0 {
			delete(watchedKeys, key)
		}
	}

	watchedKeysMutex.Unlock()

	client.keysWatched = make(map[string]struct{})
	client.Dirty = false
}
