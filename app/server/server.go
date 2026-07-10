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

	if config.Role == "slave" {
		address := fmt.Sprintf("%s:%d", config.MasterHost, config.MasterPort)
		conn, err := net.Dial("tcp", address)

		if err != nil {
			panic(err)
		}

		//  message:=fmt.Sprintf("*1\r\n$4\r\nPING\r\n")
		conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))

	}

	for {
		conn := accept(l)
		if conn != nil {
			go handleClient(conn, config)
		}
	}

}
