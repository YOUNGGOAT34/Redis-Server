package server

import (
	"container/list"
	"errors"
	"sort"
	"strconv"
	"sync"
	"time"
)

/*

 */

type TYPE int

const (
	STRING TYPE = iota
	LIST
	STREAM
)

type Data struct {
	Type  TYPE
	Value any
}

type ResponseType int

const (
	ERROR ResponseType = iota
	SIMPLE_STRING
	BULK_STRING
	NIL
	INTEGER
	ARRAY
)

type Response struct {
	Body []byte
	Type ResponseType
}

type Node struct {
	data []byte
	Prev *Node
	Next *Node
}

type List struct {
	Head *Node
	Tail *Node
	len  int
}

var (
	blockedClients      = make(map[string]*list.List)
	blockedClientsMutex sync.RWMutex
)

var (
	database      = make(map[string]Data)
	databaseMutex sync.RWMutex
)

var (
	expiry      = make(map[string]time.Time)
	expiryMutex sync.RWMutex
)

func (list *List) PushFront(value []byte) {

	node := &Node{
		data: value,
	}

	if list.Head != nil {

		tmp := list.Head
		node.Next = tmp
		tmp.Prev = node
		list.Head = node
	} else {

		list.Head = node
		list.Tail = node
	}
	list.len++

}

func (list *List) PushBack(value []byte) {
	node := &Node{
		data: value,
	}

	if list.Head != nil {
		tmp := list.Tail
		tmp.Next = node
		list.Tail = node
		node.Prev = tmp

	} else {
		list.Head = node
		list.Tail = node
	}

	list.len++
}

func (list *List) LPop() []byte {

	if list == nil || list.len == 0 {
		return nil
	}

	tmp := list.Head
	list.Head = tmp.Next
	if list.Head != nil {
		list.Head.Prev = nil
	} else {
		list.Tail = nil
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

type StreamID struct {
	Milliseconds uint64
	Sequence     uint64
}

type StreamEntry struct {
	ID     StreamID
	Fields map[string][]byte
}

type Stream struct {
	Entries []*StreamEntry
	// Tree *Radix
	LastID StreamID
	Len    int
}

func (stream *Stream) createStreamID(id []byte) (StreamID, error) {
	    
	    if compareBytes(id,[]byte("-")){
			   return stream.Entries[0].ID,nil
		 }


		 if compareBytes(id,[]byte("+")){
			  return stream.Entries[stream.Len-1].ID,nil
		 }

	    hyphenIndex:=-1
		 for index,char:=range id{
			    if char=='-'{
					   hyphenIndex=index
						break
				 }
		 }


		 if hyphenIndex==-1{
			   
			   return StreamID{},errors.New("invalid stream Id")
		 }
      

		 milliseconds,err:=strconv.ParseUint(string(id[0:hyphenIndex]),10,64)
		 if err!=nil{
			   return StreamID{},err
		 }
		 sequence,err:=strconv.ParseUint(string(id[hyphenIndex+1:]),10,64)

		 if err!=nil{
			   return StreamID{},err
		 }

		 return StreamID{
			     Milliseconds:milliseconds,
				  Sequence:sequence,
		 },err
}

// auto generate the full id
func (stream *Stream) NextID() StreamID {
	now := uint64(time.Now().UnixMilli())

	if now > stream.LastID.Milliseconds {
		stream.LastID.Milliseconds = now
		stream.LastID.Sequence = 0
	} else {
		stream.LastID.Sequence++
	}

	return stream.LastID
}

//auto generate the sequence number

func (stream *Stream) generateSequence(userSpecifiedId []byte) (StreamID, error) {

	hyphenIndex := 0

	for index, char := range userSpecifiedId {
		if char == '-' {
			hyphenIndex = index
			break
		}
	}

	if hyphenIndex == 0 {
		return StreamID{}, errors.New("Invalid stream id")
	}

	milliseconds, err := strconv.ParseUint(string(userSpecifiedId[0:hyphenIndex]), 10, 64)

	if err != nil {
		return StreamID{}, err
	}

	return StreamID{
		Milliseconds: milliseconds,
		Sequence:     uint64(stream.LastID.Sequence) + 1,
	}, err
}

//converts a streamId into a string

func (id StreamID) String() string {
	return strconv.FormatUint(id.Milliseconds, 10) + "-" + strconv.FormatUint(id.Sequence, 10)
}

// find all entries in a given range
func (stream *Stream) xRange(startId StreamID, endId StreamID) []*StreamEntry {
	if stream.Len == 0 {
		return nil
	}



	startIndex := sort.Search(stream.Len, func(i int) bool {
		current := stream.Entries[i].ID

		if current.Milliseconds > startId.Milliseconds {
			return true
		}

		if current.Milliseconds < startId.Milliseconds {
			return false
		}

		return current.Sequence >= startId.Sequence
	})

	var entries []*StreamEntry

	for i := startIndex; i < stream.Len; i++ {
		current := stream.Entries[i].ID

		if current.Milliseconds > endId.Milliseconds || (current.Milliseconds == endId.Milliseconds && current.Sequence > endId.Sequence) {

			break
		}

		entries = append(entries, stream.Entries[i])

	}

	return entries
}

// //converts a string version of stream id into []bytes
// func(id StreamID) Bytes() []byte{
// 	   return []byte(id.String())
// }

// type Radix struct {
//     Root *RadixNode
// }

// type RadixNode struct {
//     Children map[byte]*RadixNode
//     IsId  bool
// 	 Label []byte
// 	 Entry *StreamEntry
// }

// func NewRadix() *Radix{

// 	return&Radix{

// 	   Root:&RadixNode{
// 			  Children: make(map[byte]*RadixNode),
// 		},
// 	}

// }

// func commonPrefix(a ,b []byte) int{

// 	  index:=0

// 	  for i:=0;i<len(a)&& i<len(b);i++{
// 		    if a[i]!=b[i]{
// 					break
// 			 }

// 			 index++
// 	  }

// 	  return index

// }

// func (t *Radix) Insert(entry *StreamEntry){
// 	   current:=t.Root

// 		remaining:=entry.ID.Bytes()

// 		for{
// 			 b:=remaining[0]

// 			 child,exists:=current.Children[b]

// 			 if exists{

// 				prefix:=commonPrefix(child.Label,remaining)

// 			   if len(child.Label)==prefix{
// 					  remaining=remaining[prefix:]
// 					  current=child
// 				}else if prefix<len(child.Label){
// 					   remainingLabel:=child.Label[prefix:]
// 						remaining=remaining[prefix:]
// 						child.Label=child.Label[:prefix]
// 						node1:=&RadixNode{
// 							   Label: remainingLabel,
// 								Children: child.Children,
// 								IsId: child.IsId,
// 								Entry: child.Entry,
// 						}

// 						child.Children=make(map[byte]*RadixNode)

// 						if len(remaining)>0{

// 							node2:=&RadixNode{
// 									Label: remaining,
// 									Children: make(map[byte]*RadixNode),
// 									Entry: entry,
// 									IsId: true,
// 							}

// 							child.IsId=false

// 							child.Children[node2.Label[0]]=node2
// 						}else{
// 							   child.Entry=entry
// 							   child.IsId=true
// 						}

// 						child.Children[node1.Label[0]]=node1

// 						return
// 				}

// 				if len(remaining)==0{
// 					  current.Entry=entry
// 					  current.IsId=true
// 					  return
// 				}

// 			 }else{
// 				  node:=&RadixNode{
// 					     Label: remaining,
// 					     Children: make(map[byte]*RadixNode),
// 						  Entry: entry,
// 						  IsId: true,

// 				  }

// 				  current.Children[b]=node

// 				 return
// 			 }
// 		}

// }

// func (t *Radix) Search(Id []byte) *StreamEntry{
// 	   current:=t.Root
// 		remaining:=Id
// 		for {

// 			b:=remaining[0]

// 			node,exists:=current.Children[b]

// 			if exists{
//               prefix:=commonPrefix(node.Label,remaining)
// 				  if prefix!=len(node.Label){
// 					  return nil
// 				  }
// 				  current=node
// 				  remaining=remaining[prefix:]
// 				  if len(remaining)==0{
// 					 break
// 				  }

// 			}else{
// 				  return nil
// 			}

// 		}

// 		if current.IsId{

// 			return current.Entry
// 		}

// 		return nil
// }
