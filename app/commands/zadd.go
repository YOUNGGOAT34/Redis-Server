package commands

import (
	"CacheDB/app/RESP"
	"CacheDB/app/storage"
	"fmt"
	"strconv"
)


func ZaddCommand(args [][]byte) RESP.Response{
	    if len(args)<3 || (len(args)-1)%2!=0{
			  return RESP.WrongNumberOfArguments("ZADD")
		 }

		 key:=string(args[0])

		 var totalCount int=0
	
		if data,exists:=storage.Database[key];exists{
			  if data.Type!=storage.ZSET{
				  return RESP.WrongType()
			  }

			  zs:=data.Value.(*storage.ZSet)
         
		     count,err:=add(args[1:],zs)

			  if err!=nil{
						return RESP.Response{
								Body: []byte(err.Error()),
								Type: RESP.ERROR,
						}
				}

				totalCount=count

		}else{
			  
			zs:=&storage.ZSet{
					Dict: make(map[string]*storage.SkipNode),
					List: storage.NewSkipList(),
			}
			
			count,err:=add(args[1:],zs)

			if err!=nil{
						return RESP.Response{
								Body: []byte(err.Error()),
								Type: RESP.ERROR,
						}
			}

			totalCount=count
	
			storage.Database[key]=storage.Data{
					 Value: zs,
					 Type: storage.ZSET,
			}
		}

		return RESP.Response{
				      Body: fmt.Appendf([]byte{},"%d",totalCount),
						Type: RESP.INTEGER,
			  }

}


func add(args [][]byte,zs *storage.ZSet) (int,error){

         var count int=0
	      for i:=0;i<len(args);i+=2{
				   
				   score,err:=strconv.ParseFloat(string(args[i]),64)

					if err!=nil{
					     return 0,err
					}
				  
					node:=&storage.SkipNode{
							Member: string(args[i+1]),
							Score: score,
		          }

					//if deleted is true it means an existing member was updated
					deleted:=zs.Add(node)
					if !deleted{
						count++
					}
			  }

			  return count,nil
}