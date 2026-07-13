package replication

import (
	"CacheDB/app/RESP"
	"fmt"
)


func ReplConfig(arg [][]byte) RESP.Response{
	    
	     return RESP.Response{
			      Body: []byte("OK"),
					Type: RESP.SIMPLE_STRING,
		  }
	   
}

func Psync(_args [][]byte,config *RESP.SERVER) RESP.Response{
	    //+FULLRESYNC 8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb 0\r\n

		 message:=fmt.Sprintf("FULLRESYNC %s 0",config.MASTERREPLID)

		 return RESP.Response{
			  Body: []byte(message),
			  Type: RESP.SIMPLE_STRING,
		 }
}
