package server

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

//responsible for resp encoding
func buildResponse(res Response) []byte{

   body:=res.Body
    
	switch res.Type{
		   
				case  ERROR:
						return fmt.Appendf(nil, "-%s\r\n",body)
				case SIMPLE_STRING:
					   return fmt.Appendf(nil, "+%s\r\n",body)

				case NIL:
					   
						return fmt.Appendf(nil, "$-1\r\n")

				case BULK_STRING:
					   
						return fmt.Appendf(nil, "$%d\r\n%s\r\n",len(body),body)
				case INTEGER:
						return fmt.Appendf(nil, ":%s\r\n",body)
				case ARRAY:
					   //a resp array is already encoded from the parser
						return res.Body
					   

				default :
				      
						panic("Unknown Response type")
	}
	 
}



func handleClient(conn net.Conn){
	     var request=make([]byte,1024)

		  defer conn.Close()

		  for{

			  bytesRead,err:= conn.Read(request)
				
				  if err==io.EOF || (err!=nil && strings.Contains(err.Error(),"connection reset")){
	 
					 return
					
				}
	 
				if err!=nil{
					  fmt.Fprintf(os.Stderr,"Error reading client request: %s\n",err.Error())
					  return
				}

				 
	         response:=parseRequest(request[:bytesRead])

	
				_,err=conn.Write(buildResponse(response))

				if err!=nil{
					  return
				}
		  }
	   
}


func accept(listener net.Listener) net.Conn{
	   conn,err := listener.Accept()

		if err != nil {
				fmt.Fprintf(os.Stderr,"Error accepting connection: %s\r\n", err.Error())
				return nil
			}

			return conn

}

func StartServer(){

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}


	for{
         conn:=accept(l)
			if conn!=nil{
				go handleClient(conn)
			}
	}
	
}