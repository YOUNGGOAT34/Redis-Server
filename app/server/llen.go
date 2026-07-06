package server

import "strconv"

func llenCommand(arguments [][]byte) Response {
	  if len(arguments)!=1{
		    return wrongNumberOfArguments("LLEN")
	  }

	  databaseMutex.RLock()
	  data,exists:=database[string(arguments[0])]
	  databaseMutex.RUnlock()

	  

	  if exists{
		    if data.Type!=LIST{
				    return wrongType()
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