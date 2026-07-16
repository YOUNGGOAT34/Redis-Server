package rdb

import "CacheDB/app/RESP"


type RDB struct{
	   Dir string
		DbFileName string
}

func (rdbfile *RDB) RdbConfig(args [][]byte) RESP.Response{
	      if len(args)<2{
				   return RESP.WrongNumberOfArguments("CONFIG")
			}

			if RESP.CompareBytes([]byte("GET"),args[0]){
				   if RESP.CompareBytes([]byte("dir"),args[1]){
						    return RESP.Response{
								     Body: RESP.EncodeArray([][]byte{args[1],[]byte(rdbfile.Dir)}),
									   Type: RESP.ARRAY,
							 }
					}else if RESP.CompareBytes(args[1],[]byte("dbfilename")){
						     return RESP.Response{
								     Body: RESP.EncodeArray([][]byte{args[1],[]byte(rdbfile.DbFileName)}),
									  Type: RESP.ARRAY,
							 }
					}
			}

			return RESP.Response{}
}