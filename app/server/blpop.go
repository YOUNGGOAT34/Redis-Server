package server

import (
	"container/list"
	"strconv"
	"time"
)




func blockClient(arguments [][]byte) Response{
	
	
	   ch:=make(chan []byte,1)
		blockedClientsMutex.Lock()
		q,ok:=blockedClients[string(arguments[0])]

		if !ok{
			q=list.New()
			blockedClients[string(arguments[0])]=q
		}

		element:=q.PushBack(ch)

		blockedClientsMutex.Unlock()

		timeout,err:=strconv.Atoi(string(arguments[1]))




		if err!=nil{
			return Response{

					Body: []byte("Error timeout should be an integer"),
					Type: ERROR,
			}
		}


		if timeout==0{
			 value:=<-ch

			 return Response{
									Body: encodeArray([][]byte{arguments[0],value}),
									Type: ARRAY,
							
							
				}

		}


		select{
				case value:=<-ch:
					       
								return Response{
									Body: encodeArray([][]byte{arguments[0],value}),
									Type: ARRAY,
							
							
				}

			case <-time.After(time.Duration(timeout)*time.Second):
				    
					blockedClientsMutex.Lock()
					q.Remove(element)
					blockedClientsMutex.Unlock()
					return Response{
								Body: nil,
								Type: NIL,
					}
					
				
			}

}


func blpopCommand(arguments [][]byte)Response{
	   if len(arguments)!=2{
			   return Response{
						   Body:[]byte("Error: Wrong number of arguments passed to blpop command"),
							Type: ERROR,
					}
		}

		databaseMutex.RLock()
		data,exists:=database[string(arguments[0])]
		if exists {

			  if data.Type!=LIST{
				     databaseMutex.RUnlock()
				    return Response{
									Body:[]byte("WRONGTYPE Operation against a key holding the wrong kind of value"),
									Type:ERROR,

					 }
			  }
			    
			    listData:=data.Value.(*List)



				 if listData==nil || listData.len==0{
               
					     databaseMutex.RUnlock()
                   
						  return blockClient(arguments)
						 
				 }

				
				databaseMutex.RUnlock()

            databaseMutex.Lock()

				value:=listData.LPop()

				databaseMutex.Unlock()
  
				 
              
				 return Response{
					       Body: encodeArray([][]byte{arguments[0],value}),
							 Type: ARRAY,
				 }
		}

		databaseMutex.RUnlock()
     
      return blockClient(arguments)
}