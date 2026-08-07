package storage

import (
	"CacheDB/app/RESP"
	"container/list"
	"errors"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"
)

type Command struct {
	Args [][]byte
}

type UserFlags struct {
	NoPass  bool
	Enabled bool
}

type User struct {
	Name               string
	Passwords          [][32]byte
	Flags              UserFlags
	CommandPermissions uint64
	UserMutex          sync.RWMutex
}

// acl bitmap
const (
	//strings
	GET uint64 = 1 << iota
	SET
	DEL
	INCR
	//connection
	PING
	AUTH
	ACL
	//generics
	KEYS
	TYPECMD
	SAVE
	//lists
	RPUSH
	LRANGE
	LPUSH
	LLEN
	BLPOP
	LPOP
	//streams
	XADD
	XRANGE
	XREAD
	//streams
	ZADD
	ZREM
	ZRANGE
	ZSCORE
	ZRANK
	ZCARD
)




var (

	
	AllCommands uint64 = GET | SET | DEL | INCR | PING | AUTH | ACL | KEYS | TYPECMD | SAVE | RPUSH | LRANGE | LPUSH |
	                     LLEN | BLPOP | LPOP | XADD | XRANGE | XREAD | ZADD | ZREM | ZRANGE | ZSCORE | ZRANK
   
	CategoryToPermissions=map[string]uint64{
			"READ": GET | KEYS | TYPECMD | LRANGE | ZRANGE | ZSCORE | ZRANK | XRANGE | XREAD | LLEN,
			"WRITE": SET | DEL | INCR | RPUSH | LPUSH | LPOP | BLPOP | XADD | ZADD | ZREM,
			"ADMIN": ACL | SAVE,
	}

)

var CommandToPermission = map[string]uint64{
	// Strings
	"GET":  GET,
	"SET":  SET,
	"DEL":  DEL,
	"INCR": INCR,

	// Connection
	"PING": PING,
	"AUTH": AUTH,
	"ACL":  ACL,

	// Generic
	"KEYS": KEYS,
	"TYPE": TYPECMD,
	"SAVE": SAVE,

	// Lists
	"RPUSH":  RPUSH,
	"LRANGE": LRANGE,
	"LPUSH":  LPUSH,
	"LLEN":   LLEN,
	"BLPOP":  BLPOP,
	"LPOP":   LPOP,

	// Streams
	"XADD":   XADD,
	"XRANGE": XRANGE,
	"XREAD":  XREAD,

	// Sorted Sets
	"ZADD":   ZADD,
	"ZREM":   ZREM,
	"ZRANGE": ZRANGE,
	"ZSCORE": ZSCORE,
	"ZRANK":  ZRANK,
}

type Client struct {
	Conn               net.Conn
	InTransaction      bool
	Queue              []Command
	Dirty              bool
	KeysWatched        map[string]struct{}
	InSubscribeMode    bool
	SubscribedChannels Set[string]
	User               *User
}

var (
	WatchedKeys      = make(map[string]map[*Client]struct{})
	WatchedKeysMutex sync.RWMutex
)

type TYPE int

const (
	STRING TYPE = iota
	LIST
	STREAM
	ZSET
)

type Data struct {
	Type  TYPE
	Value any
}

type Node struct {
	Data []byte
	Prev *Node
	Next *Node
}

type List struct {
	Head      *Node
	Tail      *Node
	Len       int
	ListMutex sync.RWMutex
}

// for blocking pops
var (
	BlockedClients      = make(map[string]*list.List)
	BlockedClientsMutex sync.RWMutex
)

// for blocking reads(of streams)
var (
	WaitingClients      = make(map[string]*list.List)
	WaitingClientsMutex sync.RWMutex
)

var (
	Database      = make(map[string]Data)
	DatabaseMutex sync.RWMutex
)

var (
	Expiry      = make(map[string]time.Time)
	ExpiryMutex sync.RWMutex
)

// for debugging
func typeToString(t TYPE) string {
	switch t {
	case STRING:
		return "STRING"
	case LIST:
		return "LIST"
	case STREAM:
		return "STREAM"
	case ZSET:
		return "ZSET"
	default:
		return "UNKNOWN"
	}
}

func (list *List) PushFront(value []byte) {

	node := &Node{
		Data: value,
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
	list.Len++

}

func (list *List) PushBack(value []byte) {
	node := &Node{
		Data: value,
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

	list.Len++
}

func (list *List) LPop() []byte {

	if list == nil || list.Len == 0 {
		return nil
	}

	tmp := list.Head
	list.Head = tmp.Next
	if list.Head != nil {
		list.Head.Prev = nil
	} else {
		list.Tail = nil
	}
	list.Len--

	return tmp.Data
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
	Stream string //name of the associated stream
	Fields map[string][]byte
}

type Stream struct {
	Entries []*StreamEntry
	// Tree *Radix
	LastID      StreamID
	StreamMutex sync.RWMutex
	Len         int
}

func (stream *Stream) CreateStreamID(id []byte) (StreamID, error) {

	if RESP.CompareBytes(id, []byte("-")) {
		return stream.Entries[0].ID, nil
	}

	if RESP.CompareBytes(id, []byte("+")) {
		return stream.LastID, nil
	}

	if RESP.CompareBytes(id, []byte("$")) {
		return stream.LastID, nil
	}

	hyphenIndex := -1
	for index, char := range id {
		if char == '-' {
			hyphenIndex = index
			break
		}
	}

	if hyphenIndex == -1 {

		return StreamID{}, errors.New("invalid stream Id")
	}

	milliseconds, err := strconv.ParseUint(string(id[0:hyphenIndex]), 10, 64)

	if err != nil {
		return StreamID{}, err
	}
	sequence, err := strconv.ParseUint(string(id[hyphenIndex+1:]), 10, 64)

	if err != nil {
		return StreamID{}, err
	}

	return StreamID{
		Milliseconds: milliseconds,
		Sequence:     sequence,
	}, err
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

func (stream *Stream) GenerateSequence(userSpecifiedId []byte) (StreamID, error) {

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

//converts a storage.storage.StreamID into a string

func (id StreamID) String() string {
	return strconv.FormatUint(id.Milliseconds, 10) + "-" + strconv.FormatUint(id.Sequence, 10)
}

func (stream *Stream) binarySearch(startId StreamID, inclusive bool) int {
	startIndex := sort.Search(stream.Len, func(i int) bool {
		current := stream.Entries[i].ID

		if current.Milliseconds > startId.Milliseconds {
			return true
		}

		if current.Milliseconds < startId.Milliseconds {
			return false
		}

		if inclusive {

			return current.Sequence >= startId.Sequence
		} else {
			return current.Sequence > startId.Sequence
		}
	})

	return startIndex
}

// find all entries in a given range
func (stream *Stream) XRange(startId StreamID, endId StreamID) []*StreamEntry {
	if stream.Len == 0 {
		return nil
	}

	startIndex := stream.binarySearch(startId, true)

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

func (stream *Stream) XRead(startId StreamID) []*StreamEntry {
	if stream.Len == 0 {
		return nil
	}

	startIndex := stream.binarySearch(startId, false)

	var entries []*StreamEntry

	for i := startIndex; i < stream.Len; i++ {

		entries = append(entries, stream.Entries[i])

	}

	return entries

}

//custom set

type Set[T comparable] map[T]struct{}

func NewSet[T comparable]() Set[T] {
	return make(Set[T])
}

func (s Set[T]) Add(element T) {
	s[element] = struct{}{}
}

func (s Set[T]) Remove(element T) {
	delete(s, element)
}

func (s Set[T]) Len() int {
	return len(s)
}

func (s Set[T]) Contains(element T) bool {
	_, exists := s[element]

	return exists
}

func (s Set[T]) Clear() {
	clear(s)
}

var (
	Users      = make(map[string]*User)
	UsersMutex sync.RWMutex
)

// //converts a string version of stream id into []bytes
// func(id storage.storage.StreamID) Bytes() []byte{
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
