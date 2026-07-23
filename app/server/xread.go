package server

import (
	"CacheDB/app/RESP"
	"CacheDB/app/storage"
	"container/list"
	"fmt"
	"strconv"
	"time"
)

func encodeStreams(streams [][]*storage.StreamEntry) []byte {

	var respArray []byte
	count := len(streams)
	respArray = fmt.Appendf(respArray, "*%d\r\n", count)

	for _, entries := range streams {

		if len(entries) == 0 {
			continue
		}

		respArray = fmt.Appendf(respArray, "*2\r\n")
		respArray = fmt.Appendf(respArray, "$%d\r\n%s\r\n", len(entries[0].Stream), entries[0].Stream)
		respArray = fmt.Appendf(respArray, "*%d\r\n", len(entries))

		for _, entry := range entries {

			respArray = fmt.Appendf(respArray, "*2\r\n")
			respArray = fmt.Appendf(respArray, "$%d\r\n%s\r\n", len(entry.ID.String()), entry.ID.String())

			fieldsLen := len(entry.Fields) * 2

			respArray = fmt.Appendf(respArray, "*%d\r\n", fieldsLen)

			for key, value := range entry.Fields {
				respArray = fmt.Appendf(respArray, "$%d\r\n%s\r\n", len(key), key)
				respArray = fmt.Appendf(respArray, "$%d\r\n%s\r\n", len(value), value)
			}

		}
	}

	return respArray

}

func xReadCommand(arguments [][]byte) RESP.Response {

	//map to store key-starting id, incase it is a query of a multiple streams

	keys := make(map[string][]byte)

	mid := len(arguments) / 2

	for i := 0; i < mid; i++ {
		keys[string(arguments[i])] = arguments[i+mid]
	}

	var streams [][]*storage.StreamEntry

	for key, startingId := range keys {

		storage.DatabaseMutex.RLock()
		data, exists := storage.Database[string(key)]
		storage.DatabaseMutex.RUnlock()

		if exists {

			if data.Type != storage.STREAM {

				return RESP.WrongType()
			}

			stream := data.Value.(*storage.Stream)

			stream.StreamMutex.RLock()
			defer stream.StreamMutex.RUnlock()

			startId, err := stream.CreateStreamID(startingId)

			if err != nil {
				return RESP.Response{
					Body: []byte(err.Error()),
					Type: RESP.ERROR,
				}
			}

			s := stream.XRead(startId)

			if len(s) > 0 {

				streams = append(streams, s)
			}

		}
	}

	if len(streams) == 0 {

		return RESP.Response{
			Body: []byte("-1\r\n"),
			Type: RESP.NIL,
		}
	}

	return RESP.Response{
		Body: encodeStreams(streams),
		Type: RESP.ARRAY,
	}

}

func decideTypeOfRead(arguments [][]byte) RESP.Response {

	if len(arguments) < 2 || (len(arguments)-1)%2 != 0 {
		return RESP.WrongNumberOfArguments("XREAD")
	}

	if RESP.CompareBytes(arguments[0], []byte("BLOCK")) {
		return blockingXread(arguments[1:])
	} else {
		return xReadCommand(arguments[1:])
	}

}

func blockingXread(arguments [][]byte) RESP.Response {

	timeout, err := strconv.Atoi(string(arguments[0]))
	if err != nil {
		return RESP.Response{
			Body: []byte(err.Error()),
			Type: RESP.ERROR,
		}
	}

	arguments = arguments[2:]

	if len(arguments) < 2 {
		return RESP.Response{
			Body: []byte("Error: Wrong number of arguments passed to blpop command"),
			Type: RESP.ERROR,
		}
	}

	var streams [][]*storage.StreamEntry

	storage.DatabaseMutex.RLock()
	data, exists := storage.Database[string(arguments[0])]
	storage.DatabaseMutex.RUnlock()
	if exists {
		if data.Type != storage.STREAM {

			return RESP.Response{
				Body: []byte("RESP.WrongType Operation against a key holding the wrong kind of value"),
				Type: RESP.ERROR,
			}
		}

		stream := data.Value.(*storage.Stream)
		stream.StreamMutex.RLock()
		startId, err := stream.CreateStreamID(arguments[1])

		if err != nil {

			return RESP.Response{
				Body: []byte(err.Error()),
				Type: RESP.ERROR,
			}
		}

		s := stream.XRead(startId)
		stream.StreamMutex.RUnlock()
		if len(s) > 0 {

			streams = append(streams, s)
		} else {

			streams = waitForData(stream, timeout, startId, string(arguments[0]))
		}
	} else {

		storage.DatabaseMutex.Lock()
		stream := &storage.Stream{}
		storage.Database[string(arguments[0])] = storage.Data{
			Type:  storage.STREAM,
			Value: stream,
		}
		storage.DatabaseMutex.Unlock()

		startId, err := stream.CreateStreamID(arguments[1])

		if err != nil {
			return RESP.Response{
				Body: []byte(err.Error()),
				Type: RESP.ERROR,
			}
		}

		streams = waitForData(stream, timeout, startId, string(arguments[0]))
	}

	if len(streams) == 0 {

		return RESP.Response{
			Body: []byte("-1"),
			Type: RESP.NIL,
		}
	}

	return RESP.Response{
		Body: encodeStreams(streams),
		Type: RESP.ARRAY,
	}
}

func waitForData(stream *storage.Stream, timeout int, startId storage.StreamID, key string) [][]*storage.StreamEntry {

	var streams [][]*storage.StreamEntry

	ch := make(chan bool, 1)

	storage.WaitingClientsMutex.Lock()

	q, ok := storage.WaitingClients[key]

	if !ok {
		q = list.New()
		storage.WaitingClients[key] = q
	}

	element := q.PushBack(ch)

	storage.WaitingClientsMutex.Unlock()

	if timeout == 0 {

		for {

			<-ch

			stream.StreamMutex.RLock()
			s := stream.XRead(startId)
			stream.StreamMutex.RUnlock()

			if len(s) > 0 {
				streams = append(streams, s)
				storage.WaitingClientsMutex.Lock()
				q.Remove(element)
				storage.WaitingClientsMutex.Unlock()
				break
			}

		}

	} else {

		timer := time.NewTimer(time.Duration(timeout) * time.Millisecond)
		defer timer.Stop()
	WaitLoop:
		for {

			select {

			case <-ch:

				stream.StreamMutex.RLock()
				s := stream.XRead(startId)
				stream.StreamMutex.RUnlock()

				if len(s) > 0 {
					streams = append(streams, s)
					storage.WaitingClientsMutex.Lock()
					q.Remove(element)
					storage.WaitingClientsMutex.Unlock()
					break WaitLoop
				}

			case <-timer.C:
				storage.WaitingClientsMutex.Lock()
				q.Remove(element)
				storage.WaitingClientsMutex.Unlock()
				break WaitLoop

			}

		}
	}

	return streams
}
