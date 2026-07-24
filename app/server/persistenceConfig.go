package server

import (
	aof "CacheDB/app/AOF"
	rdb "CacheDB/app/RDB"
	"CacheDB/app/RESP"
)
func getConfig(args [][]byte,rdbConfig *rdb.RDB,aofConfig *aof.AOF) RESP.Response{
	     if len(args)<2{
				   return RESP.WrongNumberOfArguments("CONFIG")
			}

			

			if RESP.CompareBytes([]byte("GET"),args[0]){
				   switch string(args[1]){
							case "dir":
								return RESP.Response{
											Body: RESP.EncodeArray([][]byte{args[1],[]byte(rdbConfig.Dir)}),
												Type: RESP.ARRAY,
									}
							case "dbfilename":
									return RESP.Response{
								     Body: RESP.EncodeArray([][]byte{args[1],[]byte(rdbConfig.DbFileName)}),
									  Type: RESP.ARRAY,
									 }
							case "appendonly":
								      var appendonly string
								      if aofConfig.AppendOnly{
											  appendonly="yes"
										}else{
											appendonly="no"
										}
										return RESP.Response{
											Body: RESP.EncodeArray([][]byte{args[1],[]byte(appendonly)}),
											Type: RESP.ARRAY,
									 }

							case "appenddirname":
										return RESP.Response{
											Body: RESP.EncodeArray([][]byte{args[1],[]byte(aofConfig.AppendDirName)}),
											Type: RESP.ARRAY,
									 }

							case "appendfilename":
											return RESP.Response{
												Body: RESP.EncodeArray([][]byte{args[1],[]byte(aofConfig.AppendFilename)}),
												Type: RESP.ARRAY,
										}

							case "appendfsync":
										return RESP.Response{
											Body: RESP.EncodeArray([][]byte{args[1],[]byte(aofConfig.AppendFsync)}),
											Type: RESP.ARRAY,
									 }

							default:
								return RESP.Response{
											Body:[]byte("Unknown configuration"),
											Type: RESP.ERROR,
									 }



							

					  }

			}

			return RESP.Response{}
}