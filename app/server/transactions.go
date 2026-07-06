package server

import "fmt"


func multiCommand(arguments [][]byte,client *Client) Response {
	 if len(arguments)!=0{
		   return wrongNumberOfArguments("MULTI")
	 }

	 if client.InTransaction{
		     return Response{
                 Body: []byte("ERR MULTI calls cannot be nested"),
					  Type: ERROR,
			  }
	 }

	 client.InTransaction=true

	 return Response{
		   Body: []byte("OK"),
			Type: SIMPLE_STRING,
	 }
}

func execCommand(arguments [][]byte,client *Client) Response{
      
	if len(arguments)!=0{
		  return wrongNumberOfArguments("EXEC")
	}

	//exec executed without multi
    if !client.InTransaction{
		   return Response{
			      Body: []byte("ERR EXEC without MULTI"),
					Type: ERROR,
		  }
	 }

	 queued:=client.Queue
	 client.Queue=nil
	 client.InTransaction=false


	 var resp []byte
	 resp=fmt.Appendf(resp,"*%d\r\n",len(queued))

	 for _,cmd:=range queued{
		     r:=dispatchCommands(client,cmd.Args)
			  resp = append(resp, buildResponse(r)...)
	 }

	return  Response{
		   Body: resp,
			Type: ARRAY,
	}
}



func discardCommand(arguments [][]byte, client *Client) Response {
	    if len(arguments)!=0{
			   return wrongNumberOfArguments("DISCARD")
		 }

		 if !client.InTransaction{
			    return Response{
					  Body: []byte("ERR DISCARD without MULTI"),
					  Type: ERROR,
				 }
		 }

		 client.InTransaction=false
		 client.Queue=nil

		 return Response{
			    Body: []byte("OK"),
				 Type: SIMPLE_STRING,
		 }
}


