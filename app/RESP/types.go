package RESP

import (
	"net"
	"sync"
	"sync/atomic"
)


type REPLICA struct{
	  Conn net.Conn
	  Offset atomic.Int64
}

type ResponseType int
type SERVER struct {
	Role             string
	MasterHost       string
	MasterPort       int
	PORT             int
	MASTERREPLID     string
	MASTERREPLOFFSET atomic.Int32
	MASTERCONN net.Conn

	REPLICAS []*REPLICA
	ReplicasMutex sync.RWMutex
}

const (
	ERROR ResponseType = iota
	SIMPLE_STRING
	BULK_STRING
	NIL
	INTEGER
	ARRAY
	RDBFILE
)

type Response struct {
	Body []byte
	Type ResponseType
}
