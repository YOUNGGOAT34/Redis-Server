package server

import "fmt"


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
	   if len(arguments)<2 || len(arguments)%2!=0{
			   return Response{
						Body: []byte("Wrong number of arguments for 'XREAD' command"),
						Type: ERROR,
					}
		}


       //map to store key-starting id, incase it is a query of a multiple streams
       
		 keys:=make(map[string][]byte)
 
		 mid:=len(arguments)/2
		  
		 for i:=0;i<mid;i++{
			     keys[string(arguments[i])]=arguments[i+mid]
		 }

		 
		databaseMutex.RLock()
		defer databaseMutex.RUnlock()

		for key,startingId:=range keys{

			if data,exists:=database[string(key)];exists{
	
				  if data.Type!=STREAM{
						  return Response{
									Body: []byte("WRONGTYPE Operation against a key holding the wrong kind of value"),
									Type: ERROR,
								}
				  }
				  
				  stream:=data.Value.(*Stream)
	
				  startId,err:=stream.createStreamID(startingId)
	
				  if err!=nil{
						return Response{
								Body: []byte(err.Error()),
								Type: ERROR,
						}
				  }
	
				  var streams [][]*StreamEntry
	
				  s:=stream.xRead(startId)
	
				  if len(s)>0{
	
					  streams = append(streams, stream.xRead(startId))
	
						 return Response{
							  Body: encodeStreams(streams),
							  Type: ARRAY,
				  }
				  }
					
			}
		}


		return Response{
			       Body: []byte("-1\r\n"),
					 Type: NIL,
		}
}
