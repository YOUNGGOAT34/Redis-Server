package commands

import (
	"CacheDB/app/RESP"
	"CacheDB/app/storage"
	"fmt"
	"strconv"
)


func ZaddCommand(args [][]byte) RESP.Response{
	    if len(args)<3{
			  return RESP.WrongNumberOfArguments("ZADD")
		 }

		 key:=string(args[0])
		 score,err:=strconv.ParseFloat(string(args[1]),64)


		 if err!=nil{
			  return RESP.Response{
				   Body: []byte(err.Error()),
					Type: RESP.ERROR,
			  }
		 }

		 node:=&storage.SkipNode{
			     Member: string(args[2]),
              Score: score,
		 }

		if data,exists:=storage.Database[key];exists{
			  if data.Type!=storage.ZSET{
				  return RESP.WrongType()
			  }

			  sortedSet:=data.Value.(*storage.ZSet)
         

			  //if deleted is true it means an existing member was updated
			  deleted:=sortedSet.Add(node)
			  count:=1

			  if deleted{
				  count=0
			  }

			  return RESP.Response{
				      Body: fmt.Appendf([]byte{},"%d",count),
						Type: RESP.INTEGER,
			  }
		}


		zs:=&storage.ZSet{
			   Dict: make(map[string]*storage.SkipNode),
				List: storage.NewSkipList(),
		}
		
		zs.Add(node)

		storage.Database[key]=storage.Data{
			    Value: zs,
				 Type: storage.ZSET,
		}

		return RESP.Response{
				      Body: fmt.Appendf([]byte{},"%d",1),
						Type: RESP.INTEGER,
			  }

}