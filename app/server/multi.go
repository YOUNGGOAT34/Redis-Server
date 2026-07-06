package server


func multiCommand(arguments [][]byte) Response {
	 if len(arguments)!=0{
		   return wrongNumberOfArguments("MULTI")
	 }

	 return Response{
		   Body: []byte("OK"),
			Type: SIMPLE_STRING,
	 }
}