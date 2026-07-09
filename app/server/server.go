package server

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"CacheDB/app/helpers"
)

//responsible for resp encoding
func buildResponse(res helpers.Response) []byte{

   body:=res.Body
    
	switch res.Type{
		   
				case  helpers.ERROR:
						return fmt.Appendf(nil, "-%s\r\n",body)
				case helpers.SIMPLE_STRING:
					   return fmt.Appendf(nil, "+%s\r\n",body)

				case helpers.NIL:
					   
						return fmt.Appendf(nil, "$-1\r\n")

				case helpers.BULK_STRING:
					  
						return fmt.Appendf(nil, "$%d\r\n%s\r\n",len(body),body)
				case helpers.INTEGER:
						return fmt.Appendf(nil, ":%s\r\n",body)
				case helpers.ARRAY:
					   //a resp array is already encoded from the parser
						return res.Body
					   
      
				default :
				      
						panic("Unknown Response type")
	}
	 
}



func handleClient(conn net.Conn,config *helpers.SERVER){
	     var request=make([]byte,1024)

		  defer conn.Close()

         client:=&Client{
				   Conn: conn,
					keysWatched: make(map[string]struct{}),
			}

		  for{

			  bytesRead,err:= conn.Read(request)
				
				  if err==io.EOF || (err!=nil && strings.Contains(err.Error(),"connection reset")){
	 
					 return
					
				}
	 
				if err!=nil{
					  fmt.Fprintf(os.Stderr,"Error reading client request: %s\n",err.Error())
					  return
				}

				 
	         response:=parseRequest(client,request[:bytesRead],config)
             
				
	
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

func StartServer(config *helpers.SERVER){
   address:=fmt.Sprintf("0.0.0.0:%d",config.PORT)
	l, err := net.Listen("tcp",address)
	if err != nil {
		fmt.Printf("Failed to bind to port %d\n",config.PORT)
		os.Exit(1)
	}


	for{
         conn:=accept(l)
			if conn!=nil{
				go handleClient(conn,config)
			}
	}
	
}