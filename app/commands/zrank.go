package commands

import (
	"CacheDB/app/RESP"
	"CacheDB/app/storage"
	"strconv"
)

func Zrank(args [][]byte) RESP.Response{
	   if len(args)!=2{
			 return RESP.WrongNumberOfArguments("ZRANK")
		}

		data,exists:=storage.Database[string(args[0])]

		if exists{
			  if data.Type!=storage.ZSET{
				 return RESP.WrongType()
			  }

			  zs:=data.Value.(*storage.ZSet)
			  
			  node,exists:=zs.Dict[string(args[1])]
			  if exists{
				   target,rank:=zs.List.Search(node)
					if target!=nil{
						
						return RESP.Response{
							 Body: []byte(strconv.FormatInt(int64(rank),10)),
							 Type: RESP.INTEGER,
						}
					}
			  }
		}

		return RESP.Response{
			 Body: []byte{},
			 Type: RESP.NIL,
		}
}