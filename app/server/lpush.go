package server

import "strconv"

func lpushCommand(arguments [][]byte) Response{
	     if len(arguments)<2{
			    return Response{
					   Body:[]byte("Wrong number of arguments for 'LPUSH' command"),
						Type:ERROR,
				 }
		  }

		  key:=string(arguments[0])


		  databaseMutex.Lock()
		  defer databaseMutex.Unlock()

		  data,exists:=database[key]

		  if exists{
             
			    if data.Type!=LIST{

								return Response{
									Body:[]byte("WRONGTYPE Operation against a key holding the wrong kind of value"),
									Type:ERROR,
						}
				 }


				  list:=data.Value.(*List)

				  for _,argument:=range arguments[1:]{
					     list.PushFront(argument)
				  }

				   var buf [32]byte

				  return Response{
					   Body:strconv.AppendInt(buf[:0],int64(list.len),10),
						Type:INTEGER,
				  }

				  
  
		  }

		  

		  node:=&Node{
			   data:arguments[1],	
				
		  }

		  list:=&List{
			    Tail: node,
				 Head:node,
				 len:1,
		  }

		

			  for _,argument:=range arguments[2:]{
                 list.PushFront(argument)
			  }


			
		  var dataObject Data
		  dataObject.Type=LIST
		  dataObject.Value=list
		  database[key]=dataObject

		  var buf [32]byte
		  

        return Response{
		  
		   Body:strconv.AppendInt(buf[:0],int64(list.len),10),
			Type:INTEGER,
			
	}

}