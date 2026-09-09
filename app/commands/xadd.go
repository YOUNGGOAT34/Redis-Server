package commands

import (
	"CacheDB/app/RESP"
	"CacheDB/app/config"
	"CacheDB/app/storage"
	"errors"
	"strconv"
)

func createStreamID(id []byte) (storage.StreamID, error) {
	//find the hyphen in the user's id

	hyphenIndex := -1
	for index, char := range id {
		if char == '-' {
			hyphenIndex = index
			break
		}
	}

	if hyphenIndex == -1 {
		return storage.StreamID{}, errors.New("invalid stream Id")
	}

	milliseconds, err := strconv.ParseUint(string(id[0:hyphenIndex]), 10, 64)
	if err != nil {
		return storage.StreamID{}, err
	}
	sequence, err := strconv.ParseUint(string(id[hyphenIndex+1:]), 10, 64)

	if err != nil {
		return storage.StreamID{}, err
	}

	return storage.StreamID{
		Milliseconds: milliseconds,
		Sequence:     sequence,
	}, err
}

func XAddCommand(arguments [][]byte, client *storage.Client, replconfig *config.SERVER) RESP.Response {
	if len(arguments) < 4 {

		return RESP.WrongNumberOfArguments("XADD")
	}

	if len(arguments[2:])%2 != 0 {
		return RESP.Response{
			Body: []byte("Error wrong number of field-value arguments"),
			Type: RESP.ERROR,
		}
	}

	var stream *storage.Stream

	key := string(arguments[0])

	// Protect both lookup and first creation of the stream. The previous
	// split lock/unlock allowed concurrent XADDs to observe a missing key,
	// publish competing stream values, and access a stream while it was being
	// published. Acquire the stream lock before releasing the database lock so
	// every caller observes a fully initialized stream and all mutations of one
	// stream are serialized by its owning mutex.
	replconfig.DatabaseMutex.Lock()
	data, exists := replconfig.Database[key]

	if exists {
		if data.Type != storage.STREAM {
			replconfig.DatabaseMutex.Unlock()
			return RESP.WrongType()
		}

		stream = data.Value.(*storage.Stream)
	} else {
		stream = &storage.Stream{}
		replconfig.Database[key] = storage.Data{
			Type:  storage.STREAM,
			Value: stream,
		}
	}

	stream.StreamMutex.Lock()
	replconfig.DatabaseMutex.Unlock()
	defer stream.StreamMutex.Unlock()

	var Id storage.StreamID

	/*
		     The id format is millisecondsTime-sequence

		      If the given id is just a * --> auto generate both the millisecondsTime portion and the  sequence portion
				else if it is millisecondsTime-* --> auto generate the sequence number
				else ---> use the specified id
	*/

	if string(arguments[1]) == "*" {

		Id = stream.NextID()

	} else {
		var err error

		if exists, _ := hasWildCard(arguments[1], '*'); exists {
			Id, err = stream.GenerateSequence(arguments[1])
		} else {

			Id, err = createStreamID(arguments[1])
		}

		if err != nil {

			return RESP.Response{
				Body: []byte("Invalid stream Id"),
				Type: RESP.ERROR,
			}
		}

		/*
			Id validation
			0-0 is invalid

			millisecondsTime portion of the new Id must be greater or equal to the last entry's  millisecondsTime
			If millisecondsTime values are equal the sequence number of the new id must be greater than the last entry's sequence number

		*/

		if Id.Milliseconds == 0 && Id.Sequence == 0 {
			return RESP.Response{
				Body: []byte("ERR The ID specified in XADD must be greater than 0-0"),
				Type: RESP.ERROR,
			}
		}

		if Id.Milliseconds < stream.LastID.Milliseconds {

			return RESP.Response{
				Body: []byte("ERR The ID specified in XADD is equal or smaller than the target stream top item"),
				Type: RESP.ERROR,
			}
		}

		if Id.Milliseconds == stream.LastID.Milliseconds {

			if Id.Sequence <= stream.LastID.Sequence {
				return RESP.Response{
					Body: []byte("ERR The ID specified in XADD is equal or smaller than the target stream top item"),
					Type: RESP.ERROR,
				}
			}
		}

	}

	fields := make(map[string][]byte)

	for i := 2; i < len(arguments); i += 2 {
		fields[string(arguments[i])] = arguments[i+1]
	}

	entry := &storage.StreamEntry{
		ID:     Id,
		Fields: fields,
	}

	entry.Stream = key

	stream.LastID = Id
	stream.Entries = append(stream.Entries, entry)
	stream.Len++

	storage.WaitingClientsMutex.Lock()
	defer storage.WaitingClientsMutex.Unlock()

	if q, ok := storage.WaitingClients[key]; ok {

		for element := q.Front(); element != nil; element = element.Next() {
			ch := element.Value.(chan bool)

			select {
			case ch <- true:
			default:

			}
		}
	}

	markDirty(key, client)

	return RESP.Response{
		Body: []byte(Id.String()),
		Type: RESP.BULK_STRING,
	}

}
