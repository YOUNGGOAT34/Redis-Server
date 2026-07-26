package server

import (
	"CacheDB/app/RESP"
	"fmt"
)




func encodePubSubResponse(channel []byte,count int) []byte{
	   var resp []byte
		resp=fmt.Appendf(resp,"*3\r\n")
		resp=fmt.Appendf(resp,"$%d\r\nsubscribe\r\n",len("subscribe"))
		resp=fmt.Appendf(resp,"$%d\r\n%s\r\n",len(channel),channel)
		resp=fmt.Appendf(resp,":%d\r\n",count)
		return resp
}

func sub(args [][]byte) RESP.Response{
	   if len(args)!=2{
			    
			   return RESP.WrongNumberOfArguments("SUBSCRIBE")
		}

		return RESP.Response{
			   Body: encodePubSubResponse(args[1],1),
				Type: RESP.ARRAY,
		}
}