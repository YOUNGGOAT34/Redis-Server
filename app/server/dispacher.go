package server



func dispatchCommands(args [][]byte) Response{
	   
		  if len(args)<1{
			    return Response{
					Body:nil,
					Type:NIL,
				}
		  }


        
		  command:=args[0]


		  if compareBytes(command,[]byte("ECHO")){
			      if len(args)<2{
						  return Response{
							Body:nil,
							Type:NIL,
						}
					}

					return Response{
						Body:args[1],
						Type:BULK_STRING,
					}
		  }else if compareBytes(command,[]byte("PING")){
			       return  Response{
						Body:[]byte("PONG"),
						Type:SIMPLE_STRING,
					 }
		  }else if compareBytes(command,[]byte("SET")){
			          if len(args)<2{
							   return Response{
									Body:nil,
									Type:NIL,
								}
						 }
			          return setCommand(args[1:])
		  }else if compareBytes(command,[]byte("GET")){
			       return getCommand(args[1:])
		  }else if compareBytes(command,[]byte("RPUSH")){
			       return rPushCommand(args[1:])
		  }
   
		  return Response{
			   Body:nil,
				Type:NIL,
			}
			    
}


