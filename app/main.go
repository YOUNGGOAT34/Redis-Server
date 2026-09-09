package main

import (
	"CacheDB/app/AOF"
	"CacheDB/app/RDB"
	"CacheDB/app/config"
	"CacheDB/app/server"
	"CacheDB/app/storage"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {

	currentWorkingDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error:%s\r\n", err.Error())
		return
	}

	replConfig := &config.SERVER{
		Database: make(map[string]storage.Data),
		Expiry:   make(map[string]time.Time),
	}

	PORT := flag.Int("port", 6379, "server port")
	replicaof := flag.String("replicaof", "", "master and host port")

	dir := flag.String("dir", currentWorkingDir, "rdb file directory")
	dbfilename := flag.String("dbfilename", "rdbfile.db", "rdb filename")

	appendonly := flag.String("appendonly", "no", "yes or no")

	appenddirname := flag.String("appenddirname", "appendonly", "appendonly directory")

	appendfilename := flag.String("appendfilename", "appendonly.aof", "appendonly filename")

	appendfsync := flag.String("appendfsync", "everysec", "i.e everysec")

	flag.Parse()

	// replication configuration
	replConfig.PORT = *PORT

	if len(*replicaof) > 0 {
		parts := strings.Fields(*replicaof)
		if len(parts) < 2 {
			fmt.Fprintf(os.Stderr, "replicaof expected master and host port\n")
			return
		}

		masterPort, err := strconv.Atoi(parts[1])

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			return
		}

		replConfig.Role = "slave"
		replConfig.MasterPort = masterPort
		replConfig.MasterHost = parts[0]

	} else {
		replConfig.Role = "master"
	}

	replConfig.MASTERREPLID = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
	replConfig.MASTERREPLOFFSET.Store(0)

	replConfig.PubSub.Channels = make(map[string]storage.Set[net.Conn])

	rdbFileConfig, aofFileConfig := buildPersistenceConfig(
		*dir,
		*dbfilename,
		*appendonly,
		*appenddirname,
		*appendfilename,
		*appendfsync,
	)

	server.StartServer(replConfig, rdbFileConfig, aofFileConfig)
}

// buildPersistenceConfig wires the --dir/--dbfilename/--appendonly/... flags into
// the RDB and AOF configuration structs.
//
// --dir is the single base persistence directory for this server instance: the RDB
// file is written directly inside it, and the AOF directory (--appenddirname) is
// created as a subdirectory of it, so a coherent layout looks like:
//
//	<dir>/
//	├── <dbfilename>
//	└── <appenddirname>/
//	    ├── <appendfilename>.manifest
//	    └── <appendfilename>.<n>.incr.aof
//
// Extracted from main() so the wiring itself (not just the underlying packages)
// is unit-testable — see main_test.go.
func buildPersistenceConfig(dir, dbfilename, appendonly, appenddirname, appendfilename, appendfsync string) (*rdb.RDB, *aof.AOF) {
	rdbFileConfig := &rdb.RDB{
		Dir:        dir,
		DbFileName: dbfilename,
	}

	aofFileConfig := &aof.AOF{
		Dir:            dir,
		AppendOnly:     appendonly,
		AppendDirName:  appenddirname,
		AppendFilename: appendfilename,
		AppendFsync:    appendfsync,
	}

	return rdbFileConfig, aofFileConfig
}
