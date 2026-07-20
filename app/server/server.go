package server

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"io"
	"net"
	"os"
	"strings"

	rdb "CacheDB/app/RDB"
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

func handleClient(conn net.Conn, replConfig *RESP.SERVER,rdbConfig *rdb.RDB) {
	var request []byte
	var temp=make([]byte,1024)
   
	defer conn.Close()

	client := &Client{
		Conn:        conn,
		keysWatched: make(map[string]struct{}),
	}


	for {


     
		bytesRead, err := conn.Read(temp)
       
		if err == io.EOF || (err != nil && strings.Contains(err.Error(), "connection reset")) {

			return

		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading client request: %s\n", err.Error())
			return
		}

		request = append(request, temp[:bytesRead]...)

      for{
			parsedRequest,bytesConsumed,err:=RESP.ParseRequest(request)

			
			var response RESP.Response
	
			if err!=nil{

				   if errors.Is(err,RESP.ErrIncomplete){
						  break
					}

					response=RESP.Response{
							Body: []byte(err.Error()),
							Type: RESP.ERROR,
					}
	
			}else{
				  

				response= dispatchCommands(client,parsedRequest,replConfig,rdbConfig)
			}

			commandBytes:=request[:bytesConsumed]
			
			request=request[bytesConsumed:]
	
	
			_, err = conn.Write(RESP.EncodeResponse(response))
	
			if len(parsedRequest)>0  && RESP.CompareBytes(parsedRequest[0],[]byte("PSYNC")){
	
					 if err != nil {
							return
						}
	           
					data,err:=os.ReadFile(rdbConfig.Dir+"/"+rdbConfig.DbFileName)
					if err!=nil{
						   fmt.Fprintf(os.Stderr,"Error reading the rdb file in the master: %s\r\n",err.Error())
							return
					}

					 _,err=conn.Write(RESP.EncodeResponse(RESP.Response{
						Body: data,
						Type: RESP.RDBFILE,
					
					}))
	
					if err != nil {
						   fmt.Fprintf(os.Stderr,"Error sending an rdb file to the replica: %s\r\n",err.Error())
							return
						}

					replica:=&RESP.REPLICA{
						  Conn:conn,
					}

					replica.Offset.Store(-1)
					replConfig.ReplicasMutex.Lock()
					replConfig.REPLICAS = append(replConfig.REPLICAS,replica)
					replConfig.ReplicasMutex.Unlock()
					continue
			}
			
			if err != nil {
				return
			}
	
	
		  if replConfig.Role=="master"{
			//only propagate successful write commands
				if len(parsedRequest) > 0 && isWrite(parsedRequest[0]) && response.Type!=RESP.ERROR {
		
						replication.PropagateCommands(commandBytes,replConfig)
						replConfig.MASTERREPLOFFSET.Add(int32(bytesConsumed))
				}
		  }
		}
		
      

	}

}


//for replicas
func handleMaster(conn net.Conn,replConfig *RESP.SERVER) {

	var request []byte
	temp:=make([]byte,1024)

	defer conn.Close()
	     
	for {
			   
				

				bytesRead, err := conn.Read(temp)
            
				if err == io.EOF || (err != nil && strings.Contains(err.Error(), "connection reset")) {
					return
				}

				if err != nil {
					fmt.Fprintf(os.Stderr, "Error reading client request: %s\n", err.Error())
					return
				}

				request = append(request, temp[:bytesRead]...)

				for{

					  parsedRequest,bytesConsumed,err:=RESP.ParseRequest(request)
                 

					  if err!=nil{

						  if errors.Is(err,RESP.ErrIncomplete){
								 break
						  }

						  if RESP.CompareBytes([]byte("+OK\r\n"),request){
                            request=request[5:]
									 continue
						  }

						    
						   fmt.Fprintf(os.Stderr,"Parse error %v\n",err)
							return
                  
					  }

				      request=request[bytesConsumed:]
						
						response:=dispatchCommands(&Client{},parsedRequest,replConfig,&rdb.RDB{})
                   
						if len(parsedRequest)>0 && RESP.CompareBytes(parsedRequest[0],[]byte("REPLCONF")){
							
							   _,err=conn.Write(RESP.EncodeResponse(response))

								if err!=nil{
									 return
								}
						}

						replConfig.MASTERREPLOFFSET.Add(int32(bytesConsumed))
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

func StartServer(replConfig *RESP.SERVER,rdbConfig *rdb.RDB) {
	address := fmt.Sprintf("0.0.0.0:%d", replConfig.PORT)
	l, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Printf("Failed to bind to port %d\n", replConfig.PORT)
		os.Exit(1)
	}

    

    //load rdb file from memory
	 dataEntries,err:=rdb.ReadRDBFile(rdbConfig)
	 if err!=nil{
		    fmt.Fprintf(os.Stderr,"Error loading an rdb file :%s\r\n",err.Error())
	 }

	 for _,dataEntry:=range dataEntries{
		   database[string(dataEntry.Key)]=Data{
				   Value: dataEntry.Value,
					Type: TYPE(dataEntry.Type),
			}
	 }



	//sync with the master if this server is a replica

	if replConfig.Role == "slave" {
		address := net.JoinHostPort(replConfig.MasterHost, fmt.Sprintf("%d", replConfig.MasterPort))
		conn, err := net.Dial("tcp", address)

		if err != nil {
			panic(err)
		}

      
		replConfig.MASTERCONN=conn
		
		message:="*1\r\n$4\r\nPING\r\n"
		err=handShake(message,conn,ExpectPong)

		if err!=nil{
			  panic(err)
		}


		port:=fmt.Sprintf("%d",replConfig.PORT)
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
		rdbBytes,err:=receiveFullResync(message,conn)


		if err!=nil{
			  panic(err)
		}


	path := rdbConfig.Dir + "/" + rdbConfig.DbFileName

	_, err = os.Stat(path)

	if os.IsNotExist(err) {
		err = os.WriteFile(path, rdbBytes, 0644)

		if err != nil {

			panic(err)
		}

	}else{
		
		err=os.WriteFile(path,rdbBytes,0644)
		   
	}

   
	 if err!=nil{
		   panic(err)
	 }


		go handleMaster(conn,replConfig)
   
	}

	for {
		
		conn := accept(l)
		if conn != nil {
			go handleClient(conn, replConfig,rdbConfig)
		}
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
		  }

		return nil
}


func receiveFullResync(message string,conn net.Conn) ([]byte,error){
	    temp:=make([]byte,1024)
		 var buffer []byte

		

		  _,err:=conn.Write([]byte(message))
	
		  if err!=nil{
					return nil,err
		  }

		 

		 firstNewLinePos:=0

		 //read the fullresync line
		
		 for{

			   n,err:=conn.Read(temp)

				if err!=nil{
					  return nil,err
				}

				buffer = append(buffer, temp[:n]...)

				firstNewLinePos=bytes.Index(buffer[:n],[]byte("\r\n"))

				if firstNewLinePos==-1{
					
					  continue
				}

				
				fullyResync:=buffer[:firstNewLinePos]

            if !strings.HasPrefix(string(fullyResync), "+FULLRESYNC"){
								return nil,errors.New("Unexpected Response from the master\n")
					}

				buffer=buffer[firstNewLinePos+2:]
				
				break
		 }

		 //read the rdb length
		 secondNewLine:=0
		 rdbLength:=0
		 for{
			secondNewLine=bytes.Index(buffer,[]byte("\r\n"))

			if secondNewLine!=-1{
				  lengthLine:=buffer[:secondNewLine]
				  //skip the header(+2 for the crlf)
				  buffer=buffer[secondNewLine+2:]

				  if len(lengthLine)==0 || lengthLine[0]!='$'{
					   return nil,errors.New("Expected RDB length")
				  }

				  rdbLength,err=strconv.Atoi(string(lengthLine[1:]))

				  if err!=nil{
					 return nil,err
				  }

				  break
				 
				  
			}

			n,err:=conn.Read(temp)

			if err!=nil{
				  return nil,err
			}


			   buffer = append(buffer, temp[:n]...)
		 }


		 rdbData:=make([]byte,rdbLength)
		 /*
		   There might be some rdb files in the buffer ,so copy them and get the number of bytes we copied
		 */
		 copiedBytes:=copy(rdbData,buffer)
		 buffer=buffer[copiedBytes:]


		 for copiedBytes<rdbLength{
			  n,err:=conn.Read(temp)
    
			  if err!=nil{
				 return nil,err
			  }

			  bytesNeeded:=rdbLength-copiedBytes


			  if n<=bytesNeeded{
				  copy(rdbData[copiedBytes:],temp[:n])
				  copiedBytes+=n
			  }else{
				   
				    copy(rdbData[copiedBytes:],temp[:bytesNeeded])
					 buffer=append(buffer, temp[bytesNeeded:n]...)
					 copiedBytes=bytesNeeded

			  }
		 }



		 return rdbData,nil

}