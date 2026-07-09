package server

import (
	"CacheDB/app/helpers"
	"fmt"
)

func encodeEntries(entries []*StreamEntry) []byte {
	if len(entries) == 0 {
		return []byte("*0\r\n")
	}
	var respArray []byte
	count := len(entries)
	respArray = fmt.Appendf(respArray, "*%d\r\n", count)

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

	return respArray
}

func xRangeCommand(arguments [][]byte) helpers.Response {

	if len(arguments) != 3 {
		return wrongNumberOfArguments("XRANGE")
	}

	var entries []*StreamEntry

	databaseMutex.RLock()
	data, exists := database[string(arguments[0])];
	databaseMutex.RUnlock()
	if  exists {

		if data.Type != STREAM {
			return wrongType()
		}

		stream := data.Value.(*Stream)
		stream.streamMutex.RLock()
		defer stream.streamMutex.RUnlock()

		/*
			    This guard will prevent against accessing invalid memory when the use queries with -
				 Since for empty entries the stream.createStreamID function will never be called
				 Inside the stream.Entities entities[0] can be safely accessed ,with a guarantee that there is data inside the stream
		*/
		if stream.Len == 0 {
			return helpers.Response{
				Body: encodeEntries(stream.Entries),
				Type: helpers.ARRAY,
			}
		}

		startId, err := stream.createStreamID(arguments[1])

		if err != nil {

			return helpers.Response{
				Body: []byte(err.Error()),
				Type: helpers.ERROR,
			}
		}

		endId, err := stream.createStreamID(arguments[2])

		if err != nil {
			return helpers.Response{
				Body: []byte(err.Error()),
				Type: helpers.ERROR,
			}
		}

		entries = stream.xRange(startId, endId)

	}

	return helpers.Response{
		Body: encodeEntries(entries),
		Type: helpers.ARRAY,
	}

}
