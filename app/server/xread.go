package server

import "fmt"


func encodeStreams(streams [][]*StreamEntry) []byte{
	
	  if len(streams) == 0 {
		return []byte("*0\r\n")
		}


		var respArray []byte
		count := len(streams)
		respArray = fmt.Appendf(respArray, "*%d\r\n", count)

		for _,entries:=range streams{
         
			 respArray = fmt.Appendf(respArray, "*%d\r\n", len(entries))
			
			for index, entry := range entries {
				if index==0{
					   respArray = fmt.Appendf(respArray, "*%d\r\n%s\r\n", len(entry.stream),entry.stream)
				}
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
	   if len(arguments)!=2{
			   return Response{
						Body: []byte("Wrong number of arguments for 'XREAD' command"),
						Type: ERROR,
					}
		}

		key:=arguments[0]
		startingId:=arguments[1]

		databaseMutex.RLock()
		defer databaseMutex.RUnlock()

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

			  
			  streams = append(streams, stream.xRead(startId))

			  return Response{
				        Body: encodeStreams(streams),
						  Type: ARRAY,
			  }
			    
		}

		return Response{
			       Body: []byte("*0\r\n"),
					 Type: ARRAY,
		}
}
