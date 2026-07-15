package replication

import (
	"CacheDB/app/RESP"
	"strconv"
)

func WaitCommand(args [][]byte,config *RESP.SERVER) RESP.Response{
	  
	   return RESP.Response{
			    Body: []byte(strconv.FormatInt(int64(len(config.REPLICAS)),10)),
				 Type: RESP.INTEGER,
		}
}