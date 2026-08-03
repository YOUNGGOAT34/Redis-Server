package commands

import (
	"CacheDB/app/RESP"
	"CacheDB/app/storage"
	"fmt"
	"strconv"
)

func Zrange(args [][]byte) RESP.Response{
	   
	   if len(args)!=3{
			  return RESP.WrongNumberOfArguments("ZRANGE")
		}

		data,exists:=storage.Database[string(args[0])]

		if !exists{
			   return RESP.Response{
					  Body: []byte{},
					  Type: RESP.ARRAY,
				}
		}

		if data.Type!=storage.ZSET{
			 return RESP.WrongType()
		}

		startIndex,err:=strconv.Atoi(string(args[1]))
		if err!=nil{
			  return stringToIntError(err)
		}

		stopIndex,err:=strconv.Atoi(string(args[2]))

		if err!=nil{
			  return stringToIntError(err)
		}
       
		zs:=data.Value.(*storage.ZSet)

		if startIndex<0{
			  startIndex=zs.List.Length+startIndex
		}

		if stopIndex<0{
			  stopIndex=zs.List.Length+stopIndex
		}

		fmt.Printf("start: %d ,end:%d\r\n",startIndex,stopIndex)
		
		if startIndex>=zs.List.Length || startIndex>stopIndex{
			  return RESP.Response{
				  Body: []byte{},
				  Type: RESP.ARRAY,
			  }
		}

		if stopIndex>=zs.List.Length{
			   stopIndex=zs.List.Length-1
		}


      res:=getElementsInRange(zs,startIndex,stopIndex)

		return RESP.Response{
			  Array: res,
			  Type: RESP.ARRAY,
		}

}

func getElementsInRange(zs *storage.ZSet, startIndex, stopIndex int) []RESP.Response{
	   
	   currentIndex:=0
		current:=zs.List.Head.Forward[0]

		//find the starting node

		for current!=nil && currentIndex!=startIndex{
			  current=current.Forward[0]
			  currentIndex++
		}

		res:=make([]RESP.Response,0,stopIndex-startIndex+1)

		for i:=startIndex;i<=stopIndex;i++{
			   res = append(res, RESP.Response{
					  Body: []byte(current.Member),
					  Type: RESP.BULK_STRING,
				})

				current=current.Forward[0]
		}

		return res
}

func stringToIntError(err error) RESP.Response{
	  return RESP.Response{
				   Body: []byte(err.Error()),
					Type: RESP.ERROR,
			  }
}