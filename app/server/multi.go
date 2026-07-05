package server


func multiCommand(arguments [][]byte) Response {
	 if len(arguments)!=0{
		   return Response{
							Body:[]byte("WRONGTYPE Operation against a key holding the wrong kind of value"),
							Type:ERROR,
			  }
	 }

	 return Response{
		   Body: []byte("OK"),
			Type: SIMPLE_STRING,
	 }
}