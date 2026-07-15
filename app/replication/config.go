package replication

import (
	"CacheDB/app/RESP"
	"fmt"
	"net"
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

func ReplConfig(args [][]byte,config *RESP.SERVER,conn net.Conn) RESP.Response{
	    
	    if RESP.CompareBytes(args[0],[]byte("GETACK")){
			   return RESP.Response{
					     Body: EncodeArray([][]byte{[]byte("REPLCONF"),[]byte("ACK"),[]byte(strconv.Itoa(int(config.MASTERREPLOFFSET.Load())))}),
						  Type: RESP.ARRAY,
				}
		 }


		 if RESP.CompareBytes(args[0],[]byte("ACK")){
			   
			   offset,err:=strconv.Atoi(string(args[1]))

				if err==nil{
					   config.ReplicasMutex.RLock()
					   for _,replica:=range config.REPLICAS{
							    if replica.Conn==conn{
									   replica.Offset.Store(int64(offset))
										break
								 }
						}

						config.ReplicasMutex.RUnlock()
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
