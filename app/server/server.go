package server

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"CacheDB/app/RESP"
)

// responsible for resp encoding


func handleClient(conn net.Conn, config *RESP.SERVER) {
	var request = make([]byte, 1024)

	defer conn.Close()

	client := &Client{
		Conn:        conn,
		keysWatched: make(map[string]struct{}),
	}

	for {

		bytesRead, err := conn.Read(request)

		if err == io.EOF || (err != nil && strings.Contains(err.Error(), "connection reset")) {

			return

		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading client request: %s\n", err.Error())
			return
		}

		// fmt.Printf("%q", request[:bytesRead])
     
		parsedRequest,err:=RESP.ParseRequest(request[:bytesRead])

		var response RESP.Response

		if err!=nil{
			   response=RESP.Response{
					   Body: []byte(err.Error()),
                  Type: RESP.ERROR,
				}

		}else{
			  
			response= dispatchCommands(client,parsedRequest,config)
		}

		_, err = conn.Write(RESP.EncodeResponse(response))
      
		if err != nil {
			return
		}
	}

}

func accept(listener net.Listener) net.Conn {
	conn, err := listener.Accept()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error accepting connection: %s\r\n", err.Error())
		return nil
	}

	return conn

}

func StartServer(config *RESP.SERVER) {
	address := fmt.Sprintf("0.0.0.0:%d", config.PORT)
	l, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Printf("Failed to bind to port %d\n", config.PORT)
		os.Exit(1)
	}


	//sync with the master if this server is a replica

	if config.Role == "slave" {
		address := net.JoinHostPort(config.MasterHost, fmt.Sprintf("%d", config.MasterPort))
		conn, err := net.Dial("tcp", address)

		if err != nil {
			panic(err)
		}
     

		response:=make([]byte,128)

		//  message:=fmt.Sprintf("*1\r\n$4\r\nPING\r\n")
		_,err=conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
      
		if err!=nil{
			   panic(err)
				
		}


		n,err:=conn.Read(response)

		if err!=nil{
			    panic(err)
		}


		if string(response[:n])!="+PONG\r\n"{
			     fmt.Fprintf(os.Stderr,"Unexpected Response from the master\n")
				  return
		}

   
		_,err=conn.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$4\r\n6380\r\n"))

      if err!=nil{
			    panic(err)
		}

		n,err=conn.Read(response)

		if err!=nil{
			    panic(err)
		}

		if string(response[:n])!="+OK\r\n"{
			     fmt.Fprintf(os.Stderr,"Unexpected Response from the master\n")
				  return
		}
	}

	for {
		conn := accept(l)
		if conn != nil {
			go handleClient(conn, config)
		}
	}

}
