package rdb

import "CacheDB/app/RESP"

func KeysCommand(args [][]byte,rdbConfig *RDB) RESP.Response{

	   keys,err:=ReadRDBFile(rdbConfig)

		if err!=nil{
			  return RESP.Response{
				    Body: []byte(err.Error()),
					 Type: RESP.ERROR,
			  }
		}
	   
	    return RESP.Response{
			  Body: RESP.EncodeArray(keys),
			  Type: RESP.ARRAY,
		 }
}