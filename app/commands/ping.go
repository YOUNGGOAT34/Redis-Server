package commands

import (
	"CacheDB/app/RESP"
	"CacheDB/app/storage"
)

func Ping(client *storage.Client, args [][]byte) RESP.Response {

	if !client.InSubscribeMode && len(args) == 0 {
		return RESP.Response{
			Body: []byte("PONG"),
			Type: RESP.SIMPLE_STRING,
		}

	}

	if len(args) > 0 {

		responses := make([]RESP.Response, 0, len(args)+1)

		responses = append(responses, RESP.Response{
			Type: RESP.BULK_STRING,
			Body: []byte("pong"),
		})

		for _, arg := range args {
			responses = append(responses, RESP.Response{
				Type: RESP.BULK_STRING,
				Body: arg,
			})
		}

		return RESP.Response{
			Type:  RESP.ARRAY,
			Array: responses,
		}
	}

	return RESP.Response{
		Array: []RESP.Response{
			{Body: []byte("pong"), Type: RESP.BULK_STRING},
			{Body: []byte(""), Type: RESP.BULK_STRING},
		},
		Type: RESP.ARRAY,
	}
}
