package server

import (
	"CacheDB/app/RESP"
	"errors"
	"strconv"
)

func createStreamID(id []byte) (StreamID, error) {
	//find the hyphen in the user's id

	hyphenIndex := -1
	for index, char := range id {
		if char == '-' {
			hyphenIndex = index
			break
		}
	}

	if hyphenIndex == -1 {
		return StreamID{}, errors.New("invalid stream Id")
	}

	milliseconds, err := strconv.ParseUint(string(id[0:hyphenIndex]), 10, 64)
	if err != nil {
		return StreamID{}, err
	}
	sequence, err := strconv.ParseUint(string(id[hyphenIndex+1:]), 10, 64)

	if err != nil {
		return StreamID{}, err
	}

	return StreamID{
		Milliseconds: milliseconds,
		Sequence:     sequence,
	}, err
}

func xAddCommand(arguments [][]byte, client *Client) RESP.Response {
	if len(arguments) < 4 {

		return RESP.WrongNumberOfArguments("XADD")
	}

	if len(arguments[2:])%2 != 0 {
		return RESP.Response{
			Body: []byte("Error wrong number of field-value arguments"),
			Type: RESP.ERROR,
		}
	}

	var stream *Stream

	key := string(arguments[0])

	databaseMutex.Lock()
	data, exists := database[key]
	databaseMutex.Unlock()

	if exists {
		if data.Type != STREAM {
			return RESP.WrongType()
		}

		stream = data.Value.(*Stream)
		stream.streamMutex.Lock()
		defer stream.streamMutex.Unlock()

	} else {

		stream = &Stream{
			// Tree:NewRadix(),
		}

		database[string(arguments[0])] = Data{
			Type:  STREAM,
			Value: stream,
		}

	}

	var Id StreamID

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

		if hasWildCard(arguments[1]) {
			Id, err = stream.generateSequence(arguments[1])
		} else {

			Id, err = createStreamID(arguments[1])

			if err != nil && RESP.CompareBytes(arguments[1], []byte("-")) {

			}

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

	entry := &StreamEntry{
		ID:     Id,
		Fields: fields,
	}

	entry.stream = key

	stream.LastID = Id
	stream.Entries = append(stream.Entries, entry)
	stream.Len++

	waitingClientsMutex.Lock()
	defer waitingClientsMutex.Unlock()

	if q, ok := waitingClients[key]; ok {

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

func hasWildCard(userSpecifiedId []byte) bool {
	for _, char := range userSpecifiedId {
		if char == '*' {
			return true
		}
	}
	return false
}
