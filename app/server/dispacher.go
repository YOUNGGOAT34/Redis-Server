package server

import "strings"


func dispatchCommands(args [][]byte) Response{
	   
		  if len(args)<1{
			    return Response{
					Body:nil,
					Type:NIL,
				}
		  }

		  command:=args[0]

		  //convert to a string and make it case insensitive so that it can be used in a switch case
		  cmd:=strings.ToUpper(string(command))


		  switch cmd{
			
					case "ECHO":
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

					case "PING":
								
								return  Response{
								Body:[]byte("PONG"),
								Type:SIMPLE_STRING,
							}

					case "SET":

						    if len(args)<2{
							   return Response{
									Body:nil,
									Type:NIL,
								}
						 }
			          return setCommand(args[1:])

					case "GET":
						return getCommand(args[1:])
					case "RPUSH":
						return rPushCommand(args[1:])

					case "LRANGE":
						return lRangeCommand(args[1:])
					case "LPUSH":
						return lpushCommand(args[1:])
						
               case "LLEN":
						 return llenCommand(args[1:])

					case "LPOP":
						 return lpopCommand(args[1:])

					case "BLPOP":
						  return blpopCommand(args[1:])
					case "TYPE":
						    return typeCommand(args[1:])
					case "XADD":
						    return xaddCommand(args[1:])
					default:
						return Response{
                          Body:[]byte("Error: Unknown command"),
								  Type: ERROR,
						}

		  }
}








