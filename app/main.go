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


	currentWorkingDir,err:=os.Getwd()
	
	if err!=nil{
		fmt.Fprintf(os.Stderr,"Error:%s\r\n",err.Error())
		return
	}
    
	replConfig := &RESP.SERVER{}
	rdbFileConfig:=&rdb.RDB{}
	aofFileConfig:=&aof.AOF{}

	PORT := flag.Int("port", 6379, "server port")
	replicaof := flag.String("replicaof", "", "master and host port")

	dir:=flag.String("dir",currentWorkingDir,"rdb file directory")
	dbfilename:=flag.String("dbfilename","rdbfile.db","rdb filename")

	appendonly:=flag.String("appendonly","no","yes or no")

	appenddirname:=flag.String("appenddirname",currentWorkingDir,"appendonly directory")

	appendfilename:=flag.String("appendfilename","appendnoly.aof","appendonly filename")

	appendfsync:=flag.String("appendfsync","everysec","i.e everysec")



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

	
	
	aofFileConfig.AppendDirName=*dir
	aofFileConfig.AppendFilename=*appendfilename
	aofFileConfig.AppendOnly=*appendonly
	aofFileConfig.AppendDirName=*appenddirname
	aofFileConfig.AppendFsync=*appendfsync

	server.StartServer(replConfig,rdbFileConfig,aofFileConfig)

}
