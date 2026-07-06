package server

func typeCommand(arguments [][]byte) Response {
	  if len(arguments)!=1{

		    return wrongNumberOfArguments("TYPE")
	  }

	  databaseMutex.Lock()
	  defer databaseMutex.Unlock()


	  data,exists:=database[string(arguments[0])]

	  if exists{
		    _type:=typeToByte(data.Type)
			 return Response{
				     Body: _type,
					  Type: SIMPLE_STRING,
			 }
	  }


	  return Response{
		     Body: []byte("none"),
			  Type: SIMPLE_STRING,
	  }

}

func typeToByte(_type TYPE) []byte{
	    switch _type{
		 case STRING:
			   return []byte("string")

		 case LIST:
			    return []byte("list")

		 default:
			   panic("Unkown type")
		 }
}


