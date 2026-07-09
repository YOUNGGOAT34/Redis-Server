package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"CacheDB/app/helpers"
	"CacheDB/app/server"
)





func main() {

	   config:=&helpers.SERVER{}

	    PORT:=flag.Int("port",6379,"server port")
       replicaof:=flag.String("replicaof","","master and host port")
		 
		 flag.Parse()

		 config.PORT=*PORT

		 if len(*replicaof)>0{
			    parts:=strings.Fields(*replicaof)
				 if len(parts)<2{
					  fmt.Fprintf(os.Stderr,"replicaof expected master and host port\n")
					  return
				 }
        
				 masterPort,err:=strconv.Atoi(parts[1])

				 if err!=nil{
					  fmt.Fprintf(os.Stderr,"%s\n",err.Error())
					  return
				 }

				 config.Role="slave"
				 config.MasterPort=masterPort
				 config.MasterHost= parts[0]


		 }else{
			   config.Role="master"
		 }

		 config.MASTERREPLID="8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
		 config.MASTERREPLOFFSET=0
	    

	    server.StartServer(config)

}
