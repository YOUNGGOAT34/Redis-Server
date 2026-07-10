package replication

import "CacheDB/app/RESP"


func ReplConfig(arg [][]byte) RESP.Response{
	   
	     return RESP.Response{
			      Body: []byte("OK"),
					Type: RESP.SIMPLE_STRING,
		  }
	   
}