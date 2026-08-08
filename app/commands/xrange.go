package commands

import (
	"CacheDB/app/RESP"
	"CacheDB/app/config"
	"CacheDB/app/storage"
	"fmt"
)

func encodeEntries(entries []*storage.StreamEntry) []byte {
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

func XRangeCommand(arguments [][]byte,replconfig *config.SERVER) RESP.Response {

	if len(arguments) != 3 {
		return RESP.WrongNumberOfArguments("XRANGE")
	}

	var entries []*storage.StreamEntry

	replconfig.DatabaseMutex.RLock()
	data, exists := replconfig.Database[string(arguments[0])]
	replconfig.DatabaseMutex.RUnlock()
	if exists {

		if data.Type != storage.STREAM {
			return RESP.WrongType()
		}

		stream := data.Value.(*storage.Stream)
		stream.StreamMutex.RLock()
		defer stream.StreamMutex.RUnlock()

		/*
			    This guard will prevent against accessing invalid memory when the use queries with -
				 Since for empty entries the stream.createstorage.storage.StreamID function will never be called
				 Inside the stream.Entities entities[0] can be safely accessed ,with a guarantee that there is data inside the stream
		*/
		if stream.Len == 0 {
			return RESP.Response{
				Body: encodeEntries(stream.Entries),
				Type: RESP.STREAM,
			}
		}

		startId, err := stream.CreateStreamID(arguments[1])

		if err != nil {

			return RESP.Response{
				Body: []byte(err.Error()),
				Type: RESP.ERROR,
			}
		}

		endId, err := stream.CreateStreamID(arguments[2])

		if err != nil {
			return RESP.Response{
				Body: []byte(err.Error()),
				Type: RESP.ERROR,
			}
		}

		entries = stream.XRange(startId, endId)

	}

	return RESP.Response{
		Body: encodeEntries(entries),
		Type: RESP.STREAM,
	}

}
