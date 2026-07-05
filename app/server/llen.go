package server

import "strconv"

func llenCommand(arguments [][]byte) Response {
	  if len(arguments)!=1{
		    return Response{
						   Body:[]byte("Error: Wrong number of arguments passed to lrange command"),
							Type: ERROR,
					}
	  }

	  databaseMutex.RLock()
	  data,exists:=database[string(arguments[0])]
	  databaseMutex.RUnlock()

	  

	  if exists{
		    if data.Type!=LIST{
				    return Response{
									Body:[]byte("WRONGTYPE Operation against a key holding the wrong kind of value"),
									Type:ERROR,
						       }
			 }


			list:=data.Value.(*List)
			
			list.listMutex.RLock()
			defer list.listMutex.RUnlock()

			if list==nil{
				   return Response{
									Body:[]byte("0"),
									Type:INTEGER,
						   }
			}

			var buf [32]byte
          
			return Response{
				  Body:strconv.AppendInt(buf[:0],int64(list.len),10),
				  Type: INTEGER,

			 }
	  }

	  return Response{
		    Body:[]byte("0"),
			 Type: INTEGER,
	  }

}