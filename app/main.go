package main

import (
	"flag"

	"CacheDB/app/server"
)




func main() {

	    PORT:=flag.Int("port",6379,"server port")

	    flag.Parse()

	    server.StartServer(*PORT)

}
