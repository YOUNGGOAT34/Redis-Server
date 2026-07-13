package server

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"CacheDB/app/RESP"
	"CacheDB/app/replication"
)


//for the handshake between the master and the replica
type ExpectedResponse int

const (
	ExpectPong ExpectedResponse = iota
	ExpectOK
	ExpectFullResync
)


//identify write commands

func isWrite(command []byte) bool{
	  cmd := strings.ToUpper(string(command))

	  switch cmd{
			case "SET":
				return true
			case "INCR":
				return true
			case "LPUSH":
				return true
			case "LPOP":
				return true
			case "RPUSH":
				return true
			case "XADD":
				return true
	  }

	  return false
}

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

		if len(parsedRequest)>0  && RESP.CompareBytes(parsedRequest[0],[]byte("PSYNC")){

             if err != nil {
						return
					}

			    _,err=conn.Write(RESP.EncodeResponse(RESP.Response{
					Body: replication.EmptyRDB,
					Type: RESP.RDBFILE,
				
				}))

				if err != nil {
						return
					}

				config.ReplicasMutex.Lock()
				config.REPLICAS = append(config.REPLICAS, conn)
				config.ReplicasMutex.Unlock()
				continue
		}
      
		if err != nil {
			return
		}


     if config.Role=="master"{
		//only propagate successful write commands
		   if len(parsedRequest) > 0 && isWrite(parsedRequest[0]) && response.Type!=RESP.ERROR {
   
					replication.PropagateCommands(request[:bytesRead],config)
				
			}
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

      
		config.MASTERCONN=conn
		
		message:="*1\r\n$4\r\nPING\r\n"
		err=handShake(message,conn,ExpectPong)

		if err!=nil{
			  panic(err)
		}


		port:=fmt.Sprintf("%d",config.PORT)
		message=fmt.Sprintf("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$%d\r\n%s\r\n",len(port),port)
		err=handShake(message,conn,ExpectOK)

		if err!=nil{
			  panic(err)
		}

		message="*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n"
	                
		err=handShake(message,conn,ExpectOK)

		if err!=nil{
			  panic(err)
		}


		message="*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n"
		err=handShake(message,conn,ExpectFullResync)

		if err!=nil{
			  panic(err)
		}

		go handleMaster(conn,config)
   
	}

	for {
		
		conn := accept(l)
		if conn != nil {
			go handleClient(conn, config)
		}
	}

}

func handleMaster(conn net.Conn,config *RESP.SERVER) {

	var request = make([]byte, 1024)
   
	defer conn.Close()
	     
	for {
			
				bytesRead, err := conn.Read(request)
			
				if err == io.EOF || (err != nil && strings.Contains(err.Error(), "connection reset")) {

					return

				}

				if err != nil {
					fmt.Fprintf(os.Stderr, "Error reading client request: %s\n", err.Error())
					return
				}

				parsedRequest,err:=RESP.ParseRequest(request[:bytesRead])
			
				// var response RESP.Response

			
					dispatchCommands(&Client{},parsedRequest,config)
				

	}

}




func handShake(message string,conn net.Conn,RES ExpectedResponse) error{
	  response:=make([]byte,128)


		    
		  _,err:=conn.Write([]byte(message))
	
		  if err!=nil{
					return err
		  }
	
		  n,err:=conn.Read(response)
	
		  if err!=nil{
					return err
		  }



		  switch RES{
					case ExpectPong:
							
								if string(response[:n])!="+PONG\r\n"{
										return errors.New("Unexpected Response from the master\n")
								}
					     
					case ExpectOK:
							if string(response[:n])!="+OK\r\n"{
							return errors.New("Unexpected Response from the master\n")
					  }

					case ExpectFullResync:

						if !strings.HasPrefix(string(response[:n]), "+FULLRESYNC"){
								return errors.New("Unexpected Response from the master\n")
							}

		  }


		return nil
}
