package server

import (
	"CacheDB/app/RESP"
	"strings"
)

func acl(args [][]byte) RESP.Response {

	if len(args)<1{
		  return RESP.WrongNumberOfArguments("ACL")
	}

	switch strings.ToUpper(string(args[0])){
			case "WHOAMI":
				return whoami(args[1:])
			case "GETUSER":
				return getUser(args[1:])
	}

	return RESP.Response{}
	
}

func getUser(args [][]byte) RESP.Response {
	   if len(args)!=1{
			    return RESP.WrongNumberOfArguments("ACL GETUSER")
		}

		return RESP.Response{
			Type: RESP.ARRAY,
			Array: []RESP.Response{
				     {
						  Type: RESP.BULK_STRING,
						  Body: []byte("flags"),
					  },

					  {
						   Type: RESP.ARRAY,
							Array:[]RESP.Response{},
					  },
			},
		}
}

func whoami(args [][]byte) RESP.Response {
	if len(args)!=0{
		 return RESP.WrongNumberOfArguments("ACL WHOAMI")
	}
	return RESP.Response{
		Body: []byte("default"),
		Type: RESP.BULK_STRING,
	}

}
