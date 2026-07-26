package server

import (
	"CacheDB/app/RESP"
	"CacheDB/app/storage"
)

func ping(client *storage.Client,args [][]byte) RESP.Response{
	   
	    if !client.InSubscribeMode && len(args)==0{
				 return RESP.Response{
					Body: []byte("PONG"),
					Type: RESP.SIMPLE_STRING,
				} 

		 }

		 if len(args)>0{

			 return RESP.Response{
						 Body: RESP.EncodeArray(append([][]byte{[]byte("pong")},args...)),
						 Type: RESP.ARRAY,
					 }
		 }
      

		return RESP.Response{
				Body: RESP.EncodeArray([][]byte{[]byte("pong"),[]byte("")}),
				Type: RESP.ARRAY,
			}
}