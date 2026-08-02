package commands

import (
	"CacheDB/app/RESP"
	"CacheDB/app/storage"
	"strconv"
)

func Zcard(args [][]byte) RESP.Response{
	   if len(args)!=1{
			  return RESP.WrongNumberOfArguments("ZCARD")
		}

		if data,exists:=storage.Database[string(args[0])];exists{
			 if data.Type!=storage.ZSET{
				  return RESP.WrongType()
			 }

			 zs:=data.Value.(*storage.ZSet)

			 return RESP.Response{
				 Body: []byte(strconv.FormatInt(int64(zs.List.Length),10)),
				 Type: RESP.INTEGER,
			 }
		}

		return RESP.Response{
				 Body: []byte(strconv.FormatInt(int64(0),10)),
				 Type: RESP.INTEGER,
			 }

}