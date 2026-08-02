package commands

import (
	"CacheDB/app/RESP"
	"CacheDB/app/storage"
	"strconv"
)

func ZRem(args [][]byte) RESP.Response{

	  if len(args)<2{
		   return RESP.WrongNumberOfArguments("ZREM")
	  }
     
	  var count int64=0

	  if data,exists:=storage.Database[string(args[0])];exists{
		    if data.Type!=storage.ZSET{
				  return RESP.WrongType()
			 }

			 zs:=data.Value.(*storage.ZSet)
          
			 for _,member:=range args[1:]{
				  
				 deleted:=zs.ZRem(string(member))
	         
				 if deleted{
					 count++
				 }

			 }


	  }

	  return RESP.Response{
		 Body: []byte(strconv.FormatInt(count,10)),
		 Type: RESP.INTEGER,
	  }

}