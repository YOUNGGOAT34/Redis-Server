package server

import (
	"container/list"
	"errors"
	"strconv"
	"sync"
	"time"
)

/*

 */

 

type TYPE int

const(
	  STRING TYPE=iota
	  LIST
	  STREAM
)

type Data struct{

	    Type TYPE
		 Value any
}




type ResponseType int

const (
	  ERROR ResponseType=iota
	  SIMPLE_STRING
	  BULK_STRING
	  NIL
	  INTEGER
	  ARRAY
)

type Response struct{
	   Body []byte
		Type ResponseType
}


type Node struct{
	   data []byte
		Prev *Node
		Next *Node
}


type List struct{
	    Head *Node
		 Tail *Node
		 len int
}


var (
	blockedClients =make(map[string]*list.List)
	blockedClientsMutex sync.RWMutex
)


var (
	database =make(map[string]Data)
	databaseMutex sync.RWMutex
)

var (
	   expiry=make(map[string] time.Time)
		expiryMutex sync.RWMutex
)


func (list *List) PushFront(value []byte){
	     
	     node:=&Node{
								  data:value,  
							}

							if list.Head!=nil{

								tmp:=list.Head
								node.Next=tmp
								tmp.Prev=node
								list.Head=node
							}else{

								  list.Head=node
								  list.Tail=node
				}
			list.len++
			
}

func (list *List) PushBack(value []byte){
	   node:=&Node{
			   data:value,
		}

		if list.Head!=nil{
			  tmp:=list.Tail
			  tmp.Next=node
			  list.Tail=node
			  node.Prev=tmp

		}else{
			  list.Head=node
			  list.Tail=node
		}

		list.len++
}
 
func (list *List) LPop() []byte{
  
	  if list==nil || list.len==0{
		    return nil
	  }

	  tmp:=list.Head
	  list.Head=tmp.Next
	  if list.Head!=nil{
		  list.Head.Prev=nil 
	  }else{
		   list.Tail=nil
	  }
	  list.len--

	  return tmp.data
}




/*
    Stream
	 fields:
	   Id
		Entries(map)

	Id:
	   Time in milliseconds
		Sequence of that exact time :i.e 0,1,2,3
	          
*/


type StreamID struct{
	   Milliseconds uint64
		Sequence uint64
}

type StreamEntry struct{
	    ID StreamID
		 Fields map[string][]byte
}


type Stream struct{
	   Entries []*StreamEntry
		LastID StreamID
		Len int
}


//auto generate the full id
func (stream *Stream) NextID() StreamID{
	   now:=uint64(time.Now().UnixMilli())

		if now>stream.LastID.Milliseconds{
			    stream.LastID.Milliseconds=now
				 stream.LastID.Sequence=0
		}else{
			  stream.LastID.Sequence++
		}

		return stream.LastID
}

//auto generate the sequence number

func (stream *Stream) generateSequence(userSpecifiedId []byte) (StreamID,error){
	 
	     hyphenIndex:=0

		  for index,char :=range userSpecifiedId{
			        if char=='-'{
						   hyphenIndex=index
							break
					  }
		  }

		  if hyphenIndex==0{
			    return StreamID{},errors.New("Invalid stream id")
		  }


		 milliseconds,err:=strconv.ParseUint(string(userSpecifiedId[0:hyphenIndex]),10,64)

		 if err!=nil{
			   return StreamID{},err
		 }

		 return StreamID{
			     Milliseconds: milliseconds,
				  Sequence: uint64(stream.Len)+1,
		 },err
}




func (id StreamID) String()string{
	   return  strconv.FormatUint(id.Milliseconds,10)+"-"+strconv.FormatUint(id.Sequence,10)
}
