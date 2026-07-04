package server


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

			  stream.xRead(startId)
			    
		}

		return Response{
			       Body: []byte("*0\r\n"),
					 Type: ARRAY,
		}
}