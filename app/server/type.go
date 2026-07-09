package server

import "CacheDB/app/helpers"

func typeCommand(arguments [][]byte) helpers.Response {
	  if len(arguments)!=1{

		    return wrongNumberOfArguments("TYPE")
	  }

	  databaseMutex.Lock()
	  defer databaseMutex.Unlock()


	  data,exists:=database[string(arguments[0])]

	  if exists{
		    _type:=typeToByte(data.Type)
			 return helpers.Response{
				     Body: _type,
					  Type: helpers.SIMPLE_STRING,
			 }
	  }


	  return helpers.Response{
		     Body: []byte("none"),
			  Type: helpers.SIMPLE_STRING,
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


