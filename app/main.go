package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
)


func main() {



	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	connection,err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}


	var request=make([]byte,1024)
 
	connection.Read(request)
   
	if bytes.Equal(request,[]byte("PING")){
		  connection.Write([]byte("PONG"))
	}
}
