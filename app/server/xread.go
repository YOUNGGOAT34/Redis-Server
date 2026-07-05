package server

import (
	"container/list"
	"fmt"
	"strconv"
	"time"
)


func encodeStreams(streams [][]*StreamEntry) []byte{
	
	
		var respArray []byte
		count := len(streams)
		respArray = fmt.Appendf(respArray, "*%d\r\n", count)

		for _,entries:=range streams{
         
			 if len(entries)==0{
				  continue
			 }

			 respArray = fmt.Appendf(respArray, "*2\r\n")
			 respArray = fmt.Appendf(respArray, "$%d\r\n%s\r\n", len(entries[0].stream),entries[0].stream)
			 respArray = fmt.Appendf(respArray, "*%d\r\n",len(entries))

			for _, entry := range entries {
				
				respArray = fmt.Appendf(respArray, "*2\r\n")
				respArray = fmt.Appendf(respArray, "$%d\r\n%s\r\n", len(entry.ID.String()), entry.ID.String())
	
				fieldsLen := len(entry.Fields) * 2
	
				respArray = fmt.Appendf(respArray, "*%d\r\n", fieldsLen)
	
				for key, value := range entry.Fields {
					respArray = fmt.Appendf(respArray, "$%d\r\n%s\r\n", len(key), key)
					respArray = fmt.Appendf(respArray, "$%d\r\n%s\r\n", len(value), value)
				}
	
			}
		}


		return respArray

}

func xReadCommand(arguments [][]byte) Response {
	   
        

       //map to store key-starting id, incase it is a query of a multiple streams
       
		 keys:=make(map[string][]byte)
 
		 mid:=len(arguments)/2
		  
		 for i:=0;i<mid;i++{
			     keys[string(arguments[i])]=arguments[i+mid]
		 }

		 
		

		var streams [][]*StreamEntry
     
		for key,startingId:=range keys{

			databaseMutex.RLock()
			data,exists:=database[string(key)];
			databaseMutex.RUnlock()

			if exists{
	           
				  if data.Type!=STREAM{
					     
						  return Response{
									Body: []byte("WRONGTYPE Operation against a key holding the wrong kind of value"),
									Type: ERROR,
								}
				  }
				  
				  stream:=data.Value.(*Stream)

				  stream.streamMutex.RLock()
				  defer stream.streamMutex.RUnlock()
	
				  startId,err:=stream.createStreamID(startingId)
	
				  if err!=nil{
						return Response{
								Body: []byte(err.Error()),
								Type: ERROR,
						}
				  }

				  s:=stream.xRead(startId)
	
				  if len(s)>0{
	
					  streams = append(streams,s)
				  }
					
			}
		}


	 if len(streams)==0{

		 return Response{
					  Body: []byte("-1\r\n"),
					  Type: NIL,
		 }
	 }

	 return Response{
		   Body: encodeStreams(streams),
			Type: ARRAY,
	 }

}




func decideTypeOfRead(arguments [][]byte) Response {

	if len(arguments)<2 || (len(arguments)-1)%2!=0{
			   return Response{
						Body: []byte("Wrong number of arguments for 'XREAD' command"),
						Type: ERROR,
					}
		}

		if compareBytes(arguments[0],[]byte("BLOCK")){
           return blockingXread(arguments[1:])
		}else{
			 return xReadCommand(arguments[1:])
		}

}



func blockingXread(arguments [][]byte) Response {
	
	  timeout, err := strconv.Atoi(string(arguments[0]))
	  if err!=nil{
		     return Response{
				    Body: []byte(err.Error()),
					 Type: ERROR,
			  }
	  }


	  arguments=arguments[2:]

	  if len(arguments)<2{
		     return Response{
						Body: []byte("Error: Wrong number of arguments passed to blpop command"),
						Type: ERROR,
						}
	  }


	  var streams [][]*StreamEntry

	  databaseMutex.RLock()
     data,exists:=database[string(arguments[0])];
	  databaseMutex.RUnlock()
	  if exists{
		     if data.Type!=STREAM{
				    
					return Response{
							Body: []byte("WRONGTYPE Operation against a key holding the wrong kind of value"),
							Type: ERROR,
					}
			  }


			     stream:=data.Value.(*Stream)
	           stream.streamMutex.RLock()
				  startId,err:=stream.createStreamID(arguments[1])
	
				  if err!=nil{
						return Response{
								Body: []byte(err.Error()),
								Type: ERROR,
						}
				  }

				  s:=stream.xRead(startId)
	           stream.streamMutex.RUnlock()
				  if len(s)>0{
	
					  streams = append(streams,s)
				  }else{

					   
					
					    streams=waitForData(stream,timeout,startId,string(arguments[0]))
				  }
	  }else{
		      databaseMutex.Lock()
				stream:=&Stream{}
				database[string(arguments[0])]=Data{
					    Type: STREAM,
						 Value: stream,
				}
				databaseMutex.Unlock()

				startId,err:=stream.createStreamID(arguments[1])

				if err!=nil{
					   return Response{
								Body: []byte(err.Error()),
								Type: ERROR,
						}
				}

		     streams=waitForData(stream,timeout,startId,string(arguments[0]))
	  }


   if len(streams)==0{

		return Response{
				Body: []byte("-1"),
				Type: NIL,
		}
	}


	return Response{
						Body: encodeStreams(streams),
						Type: ARRAY,
				}
}



func waitForData(stream *Stream,timeout int,startId StreamID,key string) [][]*StreamEntry{

	  var streams [][]*StreamEntry
	  
	  ch:=make(chan bool,1)

		waitingClientsMutex.Lock()
			
		q, ok := waitingClients[key]

		if !ok {
			q = list.New()
			waitingClients[key] = q
		}

		element:=q.PushBack(ch)

		waitingClientsMutex.Unlock()


		timer:=time.NewTimer(time.Duration(timeout)*time.Millisecond)
		defer timer.Stop()
		WaitLoop:
				for{
					
						select{
						case <-ch:

									stream.streamMutex.RLock()
									s:=stream.xRead(startId)
									stream.streamMutex.RUnlock()
									
									if len(s)>0{
										streams=append(streams, s)
										waitingClientsMutex.Lock()
										q.Remove(element)
										waitingClientsMutex.Unlock()
										break WaitLoop
									}

								case <-timer.C:
									waitingClientsMutex.Lock()
									q.Remove(element)
									waitingClientsMutex.Unlock()
									break WaitLoop

						}
									
				}


				return streams
}
