package server

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)


var (
	database =make(map[string]Data)
	databaseMutex sync.RWMutex
)

var (
	   expiry=make(map[string] time.Time)
		expiryMutex sync.RWMutex
)

/* 
  findCRLF returns the index of the first '\r' in the first CRLF sequence.
  If no CRLF exists, the request is considered malformed.
*/
func findCRLF(request []byte) int{
	   for i:=0;i<len(request);i++{
			   if request[i]=='\r' && i+1<len(request) && request[i+1]=='\n'{
					     return i
				}
		}
		return -1
}


/*
		Split the RESP array header (e.g. "*2") from the remaining payload.
		The returned body intentionally begins with "\r\n" so every bulk string
		can be parsed using the same logic.
*/


func getHeaderAndBody(request []byte) (header,body []byte){
	      headerEndsAt:=findCRLF(request)

			if headerEndsAt==-1{
				   return nil,nil
			}

			return request[:headerEndsAt],request[headerEndsAt:]
}



/*
	Parse a RESP request into its command arguments.
	Each bulk string is extracted and stored in args for dispatch.
 */

func parseRequest(request []byte) Response{
        if len(request)<1{
			  return Response{
				 Body:nil,
				 Type:NIL,
			  }
		  }   
   
		  

		  header,body:=getHeaderAndBody(request)
       
		  var args [][]byte

	
		  if header==nil{
			   return Response{
					 Body:nil,
					 Type:NIL,
				}
		  }


		  /*Read the array length after '*'.
        Supports multi-digit array sizes such as *12.

		  */


		  index:=1

		  for index<len(header){
			    if header[index]>='0' && header[index]<='9'{

					 index++
				 }else {
					break
				 }
		  }


		  size,err:=strconv.Atoi(string(header[1:index]))

		  if err!=nil{
			   return Response{
					 Body:nil,
					 Type:NIL,
				}
		  }



		  // Extract each RESP bulk string from the request body.
		  
		  for i:=0;i<size;i++{
			  

				if len(body)<5{
					   fmt.Fprint(os.Stderr,"Malformed body\n");
						return Response{
							  Body:nil,
							  Type:NIL,
						}
				 }

				 /* 
				     Find the end of the bulk string length.
				     Example:
				     "\r\n$34\r\nhello..."
				
								 ^
								stop here

					  */

				index:=4

				 
				 for index<len(body){

							   if body[index]>='0' && body[index]<='9'{
									 index++
								}else{
									  break
								}
				 }
				 
				

				 digits:=body[3:index]
				 // Convert the ASCII digits ("34") into an integer.
			    elementSize,err:=strconv.Atoi(string(digits))
				 if err!=nil{
					   fmt.Fprintf(os.Stderr,"Error converting string to integer %s\n",err.Error())
						return Response{
							 Body:nil,
							 Type:NIL,
						}
				 }


				/*

				   Number of bytes before the payload.
					
					For example:
					
					"\r\n$4\r\n"   -> offset = 6
					"\r\n$34\r\n"  -> offset = 7
				
				    The offset grows as the length field gains more digits.

				*/


				offset:=5+len(digits)
				
           
				if elementSize+offset<=len(body){
					// Extract the payload and advance the body to the next bulk string.
               arg:=body[5+len(digits):elementSize+offset]
					args=append(args,arg)
					body=body[offset+elementSize:]
					


				}else{

					  fmt.Fprintf(os.Stderr,"Malformed erro\n")
					  return Response{
						          Body:nil,
									 Type:NIL,
					  }
				}

		  }
        
		  return dispatchCommands(args)
}



func compareBytes(a ,b []byte) bool{
	   if bytes.EqualFold(a,b){
			   return true
		}

		return false
}


func dispatchCommands(args [][]byte) Response{
	   
		  if len(args)<1{
			    return Response{
					Body:nil,
					Type:NIL,
				}
		  }


        
		  command:=args[0]


		  if compareBytes(command,[]byte("ECHO")){
			      if len(args)<2{
						  return Response{
							Body:nil,
							Type:NIL,
						}
					}

					return Response{
						Body:args[1],
						Type:BULK_STRING,
					}
		  }else if compareBytes(command,[]byte("PING")){
			       return  Response{
						Body:[]byte("PONG"),
						Type:SIMPLE_STRING,
					 }
		  }else if compareBytes(command,[]byte("SET")){
			          if len(args)<2{
							   return Response{
									Body:nil,
									Type:NIL,
								}
						 }
			          return setCommand(args[1:])
		  }else if compareBytes(command,[]byte("GET")){
			       return getCommand(args[1:])
		  }
   
		  return Response{
			   Body:nil,
				Type:NIL,
			}
			    
}

func getCommand(arguments [][]byte) Response {
	   if len(arguments)<1{
			  return Response{
				Body:[]byte("Wrong number of arguments"),
				Type:ERROR,
			  }
		}


		expiryMutex.Lock()

		expires,exists:=expiry[string(arguments[0])];

		if exists{
			    if time.Now().After(expires){
					   databaseMutex.Lock()
						delete (database,string(arguments[0]))
						databaseMutex.Unlock()
						delete (expiry,string(arguments[0]))
				 }
		}

		expiryMutex.Unlock() 


		databaseMutex.RLock()
      dataObject,exists:=database[string(arguments[0])]
		databaseMutex.RUnlock()

		if exists{
			     
			     if dataObject.Type!="String"{
					     return Response{
							Body:[]byte("WRONGTYPE Operation against a key holding the wrong kind of value"),
							Type:ERROR,
						  }
				  }

				  value:=dataObject.Value.(string)
			     return Response{
					    Body:[]byte(value),
						 Type:BULK_STRING,
				  }
		}

      
		return Response{
			  Body:nil,
			  Type:NIL,
		}
}


func setCommand(arguments [][]byte) Response {
	   if len(arguments)<2 {
			   return Response{
					Body:[]byte("Wrong number of arguments"),
					Type:ERROR,
				}
		}

	
		/*

		  Arguments[0]-->key
		  Arguments[1]-->value

		*/
   

		expiryMutex.Lock()
		delete (expiry,string(arguments[0]))
		expiryMutex.Unlock() 


       
		if len(arguments)>2{
			    if compareBytes(arguments[2],[]byte("EX")) || compareBytes(arguments[2],[]byte("PX")){
					     if len(arguments)<4{
							  return Response{
								     Body:[]byte(""),
									  Type:BULK_STRING,
							     }
						  } 
						  
						  if compareBytes(arguments[2],[]byte("EX")){
							   
							    timeInSeconds,err:=strconv.Atoi(string(arguments[3]))
								 if err!=nil{
									 return Response{
										   Body:[]byte("Error invalid expiry time"),
											Type:ERROR,
									 }
									}


                           expiryMutex.Lock()

									duration:=time.Duration(timeInSeconds)*time.Second
									expiresAt:=time.Now().Add(duration)
									expiry[string(arguments[0])]=expiresAt

									expiryMutex.Unlock()

						  }else if compareBytes(arguments[2],[]byte("PX")){
								timeInMilliSeconds,err:=strconv.Atoi(string(arguments[3]))
								if err!=nil{
									return Response{
										     Body:[]byte("Error invalid expiry time"),
											  Type:ERROR,
									}
								}
  
								 expiryMutex.Lock()
								 duration:=time.Duration(timeInMilliSeconds)*time.Millisecond
								 expiresAt:=time.Now().Add(duration)
								 expiry[string(arguments[0])]=expiresAt
								 expiryMutex.Unlock()
						  }
				 }
		}


		var dataObject Data

		dataObject.Type="String"
		dataObject.Value=string(arguments[1])


		databaseMutex.Lock()
		database[string(arguments[0])]=dataObject
		databaseMutex.Unlock()

		return Response{
			   Body:[]byte("OK"),
				Type:SIMPLE_STRING,
		  }

} 






