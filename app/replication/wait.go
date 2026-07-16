package replication

import (
	"CacheDB/app/RESP"
	"strconv"
	"time"
)




func WaitCommand(args [][]byte,config *RESP.SERVER) RESP.Response{

	  if len(args)<2{
		      
				return RESP.Response{
					Body: []byte("Wrong number of arguments for 'WAIT' command",),
					Type: RESP.ERROR,
				}
	  }

	  targetOffset:=config.MASTERREPLOFFSET.Load()
	  ack := RESP.EncodeResponse(RESP.Response{
						Body:EncodeArray([][]byte{
							[]byte("REPLCONF"),
							[]byte("GETACK"),
							[]byte("*"),
						}),
						Type: RESP.ARRAY,
					})

		config.ReplicasMutex.RLock()

		replicas:=append([]*RESP.REPLICA(nil),config.REPLICAS...)

		config.ReplicasMutex.RUnlock()


	   for _,replica:=range replicas{
			    _,err:=replica.Conn.Write(ack)
				 
					  if err!=nil{
					  //if the write fails remove the replica
					 
                 config.ReplicasMutex.Lock()
					   for j,r:=range config.REPLICAS{
                     if r==replica{
								config.REPLICAS[j].Conn.Close()
								config.REPLICAS = append(config.REPLICAS[:j],config.REPLICAS[j+1:]...)
								break
							}

						}
					  config.ReplicasMutex.Unlock()
					  
				}
			
				
		}

		requiredReplicas,err:=strconv.Atoi(string(args[0]))
		
		if err!=nil{
			    return RESP.Response{
					   Body: []byte(err.Error()),
						Type: RESP.ERROR,
				 }
		}

     config.ReplicasMutex.RLock()
		if requiredReplicas==0 || len(config.REPLICAS)==0{
			  config.ReplicasMutex.RUnlock()
			  return RESP.Response{
				 Body:[]byte("0"),
				 Type: RESP.INTEGER,
			  }
		}

      if requiredReplicas>len(config.REPLICAS){
			   requiredReplicas=len(config.REPLICAS)
		}

		config.ReplicasMutex.RUnlock()

		timeout,err:=strconv.Atoi(string(args[1]))

		if err!=nil{
			   return RESP.Response{
					   Body: []byte(err.Error()),
						Type: RESP.ERROR,
				 }
		}

		deadline:=time.Now().Add(time.Duration(timeout)*time.Millisecond)


		var totalCount int64

		for{
         
			config.ReplicasMutex.RLock()

			replicas:=append([]*RESP.REPLICA(nil),config.REPLICAS...)

			config.ReplicasMutex.RUnlock()

			  count:=0

			  for _,replica:=range replicas{
				     if int32(replica.Offset.Load())>=targetOffset{
						   count++
					  }
			  }

			  if count>=requiredReplicas{
				totalCount=int64(count)
				   break
			  }

			  if time.Now().After(deadline){
				 totalCount=int64(count)
				  break
			  }

			  //sleep for a millisecond to avoid busy spinning
			  time.Sleep(time.Millisecond)
		}
	  
	   return RESP.Response{
			    Body: []byte(strconv.FormatInt(totalCount,10)),
				 Type: RESP.INTEGER,
		}
}