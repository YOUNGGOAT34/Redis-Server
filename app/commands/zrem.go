package commands

import (
	"CacheDB/app/RESP"
	"CacheDB/app/storage"
	"strconv"
)

func ZRem(args [][]byte) RESP.Response{

	  if len(args)!=2{
		   return RESP.WrongNumberOfArguments("ZREM")
	  }

	  if data,exists:=storage.Database[string(args[0])];exists{
		    if data.Type!=storage.ZSET{
				  return RESP.WrongType()
			 }

			 zs:=data.Value.(*storage.ZSet)

			 deleted:=zs.ZRem(string(args[0]))

			 if deleted{
				  return RESP.Response{
						Body: []byte(strconv.FormatInt(1,10)),
						Type: RESP.INTEGER,
					}
			 }

	  }

	  return RESP.Response{
		 Body: []byte(strconv.FormatInt(0,10)),
		 Type: RESP.INTEGER,
	  }

}