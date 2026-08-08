package config

import (
	"CacheDB/app/storage"
	"net"
	"sync"
	"sync/atomic"
	"time"
)



type REPLICA struct{
	  Conn net.Conn
	  Offset atomic.Int64
}



type PubSub struct{
	    ChannelMutex sync.RWMutex
		 Channels map[string]storage.Set[net.Conn]
}

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

	Database map[string]storage.Data
	DatabaseMutex sync.RWMutex

	Expiry     map[string]time.Time
	ExpiryMutex sync.RWMutex

	PubSub PubSub
}








