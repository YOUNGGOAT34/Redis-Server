package server

import "CacheDB/app/RESP"

func keys(args [][]byte) RESP.Response{
	    if len(args)!=1{
			   return RESP.WrongNumberOfArguments("KEYS")
		 }


		 keys:=make([][]byte,0,len(database))

		 for key := range database{
			   keys = append(keys, []byte(key))
		 }
		 
		 if RESP.CompareBytes(args[0],[]byte("*")){
			   return RESP.Response{
					    Body: RESP.EncodeArray(keys),
						 Type: RESP.ARRAY,
				}
		 }

		 return RESP.Response{}
}