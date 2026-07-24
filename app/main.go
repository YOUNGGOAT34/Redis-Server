package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"CacheDB/app/AOF"
	"CacheDB/app/RDB"
	"CacheDB/app/RESP"
	"CacheDB/app/server"
)

func main() {
    
	replConfig := &RESP.SERVER{}
	rdbFileConfig:=&rdb.RDB{}
	aofFileConfig:=&aof.AOF{}

	PORT := flag.Int("port", 6379, "server port")
	replicaof := flag.String("replicaof", "", "master and host port")

	dir:=flag.String("dir",".","rdb file directory")
	dbfilename:=flag.String("dbfilename","rdbfile.db","rdb filename")

	flag.Parse()

	//replication configuration
    
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


	//rdb file config
	rdbFileConfig.Dir=*dir
	rdbFileConfig.DbFileName=*dbfilename

	//aof file config

	currentWorkingDir,err:=os.Getwd()
	
	if err!=nil{
		fmt.Fprintf(os.Stderr,"Error:%s\r\n",err.Error())
		return
	}
	
	aofFileConfig.AppendDirName=currentWorkingDir
	aofFileConfig.AppendFilename="appendnoly.aof"
	aofFileConfig.AppendOnly=false
	aofFileConfig.AppendDirName="appendonlydir"
	aofFileConfig.AppendFsync="everysec"

	server.StartServer(replConfig,rdbFileConfig,aofFileConfig)

}
