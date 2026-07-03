package server

import "fmt"


func encodeEntries(entries []*StreamEntry) []byte{
	     if len(entries)==0{
			        return []byte("*0\r\n")
		  }
	     var respArray []byte
		  count:=len(entries)
	     respArray=fmt.Appendf(respArray,"*%d\r\n",count)  

		  for _,entry:=range entries{
			    respArray=fmt.Appendf(respArray,"*2\r\n") 
			    respArray=fmt.Appendf(respArray,"$%d\r\n%s\r\n",len(entry.ID.String()),entry.ID.String())

				 fieldsLen:=len(entry.Fields)*2

				 respArray=fmt.Appendf(respArray,"*%d\r\n",fieldsLen)
				 
				 for key,value:=range entry.Fields{
					     respArray=fmt.Appendf(respArray,"$%d\r\n%s\r\n",len(key),key)
						  respArray=fmt.Appendf(respArray,"$%d\r\n%s\r\n",len(value),value)
				 }
             
		  }

		  return respArray
}




func xrangeCommand(arguments [][]byte) Response {

	   if len(arguments)!=3{
			    return Response{
						Body:[]byte("Wrong number of arguments for 'XRANGE' command"),
						Type:ERROR,

				         }
		}

		var entries []*StreamEntry

		databaseMutex.RLock()
		defer databaseMutex.RUnlock()
		if data,exists:=database[string(arguments[1])];exists{
			    
			    if data.Type!=STREAM{
					    return Response{
							Body:[]byte("WRONGTYPE Operation against a key holding the wrong kind of value"),
							Type:ERROR,
						  }
				 }

				 stream:=data.Value.(*Stream)
				
				 startId,err:=createStreamID(arguments[1])

				 if err!=nil{
					    return Response{
							   Body: []byte(err.Error()),
								Type: ERROR,
						 }
				 }

				 endId,err:=createStreamID(arguments[1])

				 if err!=nil{
					    return Response{
							   Body: []byte(err.Error()),
								Type: ERROR,
						 }
				 }

				 entries=stream.xRange(startId,endId)
           
		}


		return Response{
			   Body: encodeEntries(entries),
				Type: ARRAY,
		}
 
}