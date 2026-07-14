package replication

import (
	"CacheDB/app/RESP"
	"fmt"
	"strconv"
)


func EncodeArray(body [][]byte) []byte{
	   var res []byte
		res=fmt.Appendf(res,"*%d\r\n",len(body))

		for _,element:=range body{
           res=fmt.Appendf(res,"$%d\r\n%s\r\n",len(element),element)
		}

		return res
}

func ReplConfig(args [][]byte,config *RESP.SERVER) RESP.Response{
	    
	    if RESP.CompareBytes(args[0],[]byte("GETACK")){
			   return RESP.Response{
					     Body: EncodeArray([][]byte{[]byte("REPLCONF"),[]byte("ACK"),[]byte(strconv.Itoa(config.MASTERREPLOFFSET))}),
						  Type: RESP.ARRAY,
				}
		 }

	     return RESP.Response{
			      Body: []byte("OK"),
					Type: RESP.SIMPLE_STRING,
		  }
	   
}

func Psync(_args [][]byte,config *RESP.SERVER) RESP.Response{
	  
		 message:=fmt.Sprintf("FULLRESYNC %s 0",config.MASTERREPLID)

		 return RESP.Response{
			  Body: []byte(message),
			  Type: RESP.SIMPLE_STRING,
		 }
}
