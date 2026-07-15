package replication

import "CacheDB/app/RESP"

func WaitCommand(args [][]byte) RESP.Response{
	  
	   return RESP.Response{
			    Body: []byte("0"),
				 Type: RESP.INTEGER,
		}
}