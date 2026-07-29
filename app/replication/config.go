package replication

import (
	"CacheDB/app/RESP"
	"CacheDB/app/config"
	"fmt"
	"net"
	"strconv"
)


func ReplConfig(args [][]byte,config *config.SERVER,conn net.Conn) RESP.Response{
	    
	    if RESP.CompareBytes(args[0],[]byte("GETACK")){
			   return RESP.Response{
						Type: RESP.ARRAY,
						Array: []RESP.Response{
							{
								Type: RESP.BULK_STRING,
								Body: []byte("REPLCONF"),
							},
							{
								Type: RESP.BULK_STRING,
								Body: []byte("ACK"),
							},
							{
								Type: RESP.BULK_STRING,
								Body: []byte(strconv.Itoa(int(config.MASTERREPLOFFSET.Load()))),
							},
						},
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

func Psync(_args [][]byte,config *config.SERVER) RESP.Response{
	  
		 message:=fmt.Sprintf("FULLRESYNC %s 0",config.MASTERREPLID)

		 return RESP.Response{
			  Body: []byte(message),
			  Type: RESP.SIMPLE_STRING,
		 }
}
