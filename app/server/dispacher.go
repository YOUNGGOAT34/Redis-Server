package server

import (
	"strings"
)


func dispatchCommands(client *Client,args [][]byte) Response {

	if len(args) < 1 {
		return Response{
			Body: nil,
			Type: NIL,
		}
	}

	command := args[0]

	//convert to a string and make it case insensitive so that it can be used in a switch case
	cmd := strings.ToUpper(string(command))

	switch cmd{

		 case "MULTI":
				return multiCommand(args[1:],client)
			case "EXEC":
				return execCommand(args[1:],client)

			case "DISCARD":
				return discardCommand(args[1:],client)
			case "WATCH":
				 return watchCommand(args[1:],client)
			
		  
	}

  if client.InTransaction{
	     client.Queue = append(client.Queue, 
		               Command{
                          Args: args,
							})

			return Response{
				      Body: []byte("QUEUED"),
						Type: SIMPLE_STRING,
			}
  }


	switch cmd {

			case "ECHO":
				if len(args) < 2 {
					return Response{
						Body: nil,
						Type: NIL,
					}
				}

				return Response{
					Body: args[1],
					Type: BULK_STRING,
				}

			case "PING":

				return Response{
					Body: []byte("PONG"),
					Type: SIMPLE_STRING,
				}

			case "SET":

				if len(args) < 2 {
					return Response{
						Body: nil,
						Type: NIL,
					}
				}
				return setCommand(args[1:],client)

			case "GET":
				return getCommand(args[1:])
			case "RPUSH":
				return rPushCommand(args[1:],client)

			case "LRANGE":
				return lRangeCommand(args[1:])
			case "LPUSH":
				return lPushCommand(args[1:],client)

			case "LLEN":
				return llenCommand(args[1:])

			case "LPOP":
				return lPopCommand(args[1:],client)

			case "BLPOP":
				return bLPopCommand(args[1:],client)
			case "TYPE":
				return typeCommand(args[1:])
			case "XADD":
				return xAddCommand(args[1:],client)
			case "XRANGE":
				return xRangeCommand(args[1:])
			case "XREAD":
				return decideTypeOfRead(args[1:])
			case "INCR":
				return incrCommand(args[1:],client)
			
			case "UNWATCH":
				 return unwatchCommand(args[1:],client)
			
			default:
				return Response{
					Body: []byte("Error: Unknown command"),
					Type: ERROR,
				}

	}
}










