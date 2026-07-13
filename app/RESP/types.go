package RESP

import (
	"net"
	"sync"
)

type ResponseType int

type SERVER struct {
	Role             string
	MasterHost       string
	MasterPort       int
	PORT             int
	MASTERREPLID     string
	MASTERREPLOFFSET int
	MASTERCONN net.Conn

	REPLICAS []net.Conn
	ReplicasMutex sync.Mutex
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
