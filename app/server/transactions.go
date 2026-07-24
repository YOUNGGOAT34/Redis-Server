package server

import (
	aof "CacheDB/app/AOF"
	rdb "CacheDB/app/RDB"
	"CacheDB/app/RESP"
	"CacheDB/app/storage"
	"fmt"
)

//     |----------------------MULTI COMMAND----------------------|

func multiCommand(arguments [][]byte, client *storage.Client) RESP.Response {
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

func execCommand(arguments [][]byte, client *storage.Client, replConfig *RESP.SERVER,aofConfig *aof.AOF) RESP.Response {

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
		r := dispatchCommands(client, cmd.Args, replConfig, &rdb.RDB{},aofConfig)
		resp = append(resp, RESP.EncodeResponse(r)...)
	}

	return RESP.Response{
		Body: resp,
		Type: RESP.ARRAY,
	}
}

//     |----------------------DISCARD COMMAND----------------------|

func discardCommand(arguments [][]byte, client *storage.Client) RESP.Response {
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

func watchCommand(arguments [][]byte, client *storage.Client) RESP.Response {
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

		storage.WatchedKeysMutex.Lock()
		set, exists := storage.WatchedKeys[key]

		if !exists {
			set = make(map[*storage.Client]struct{})
			storage.WatchedKeys[key] = set

		}

		set[client] = struct{}{}

		storage.WatchedKeysMutex.Unlock()

		client.KeysWatched[key] = struct{}{}
	}

	return RESP.Response{
		Body: []byte("OK"),
		Type: RESP.SIMPLE_STRING,
	}
}

func unwatchCommand(arguments [][]byte, client *storage.Client) RESP.Response {
	if len(arguments) != 0 {
		return RESP.WrongNumberOfArguments("UNWATCH")
	}

	clearWatches(client)

	return RESP.Response{
		Body: []byte("OK"),
		Type: RESP.SIMPLE_STRING,
	}
}

func clearWatches(client *storage.Client) {
	storage.WatchedKeysMutex.Lock()

	for key := range client.KeysWatched {
		delete(storage.WatchedKeys[key], client)
		if len(storage.WatchedKeys[key]) == 0 {
			delete(storage.WatchedKeys, key)
		}
	}

	storage.WatchedKeysMutex.Unlock()

	client.KeysWatched = make(map[string]struct{})
	client.Dirty = false
}
