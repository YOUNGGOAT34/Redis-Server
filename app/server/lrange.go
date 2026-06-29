package server

import (

	"fmt"
	"os"
	"strconv"
)


func encodeRespArray(elements [][]byte) []byte{

	 var respArray []byte
	 respArray=fmt.Appendf(respArray,"*%d\r\n",len(elements))   
	 for i:=0;i<len(elements);i++{
		    element:=elements[i]
			 
			 respArray=fmt.Appendf(respArray,"$%d\r\n%s\r\n",len(element),element)
	 }
	 return respArray
}


func lRangeCommand(arguments [][]byte) Response {

	   
	   if len(arguments)!=3{
			      return Response{
						   Body:[]byte("Error: Wrong number of arguments passed to lrange command"),
							Type: ERROR,
					}
		
		}


		key:=string(arguments[0])

		startIndex,err:=strconv.Atoi(string(arguments[1]))
		if err!=nil{
			  
			   fmt.Fprintf(os.Stderr,"Error converting string to integer %s\n",err.Error())
				return Response{
					  Body:[]byte("Invalid start index"),
					  Type: ERROR,
				}
		}

		endIndex,err:=strconv.Atoi(string(arguments[2]))

		if err!=nil{
			   
			   fmt.Fprintf(os.Stderr,"Error converting string to integer %s\n",err.Error())
			
				return Response{
					  Body:[]byte("Invalid end index"),
					  Type: ERROR,
				}
		}


		databaseMutex.RLock()
		defer databaseMutex.RUnlock()
		data,exists:=database[key]

		if exists{
           
		
			    if data.Type!="List"{
								return Response{
									Body:[]byte("WRONGTYPE Operation against a key holding the wrong kind of value"),
									Type:ERROR,
						       }
				 }

				  
				 elements:=data.Value.([][]byte)


				 if len(elements)==0{
					    return Response{
										Body:encodeRespArray([][]byte{}),
										Type:ARRAY,
							}
				 }


				 if startIndex<0{
					   startIndex=len(elements)+startIndex
				 }

				
				 if startIndex>=len(elements) {
					            
									return Response{
										Body:encodeRespArray([][]byte{}),
										Type:ARRAY,
							}

							
				 }

				 
				 if endIndex>=len(elements){
					     endIndex=len(elements)-1
				 }
      
				
				 if endIndex<0{
					   endIndex=len(elements)+endIndex
				 }


				 if startIndex<0{
					   startIndex=0
				 }

				 if endIndex<0{
					   endIndex=0
				 }


				 if startIndex>endIndex{
					     return Response{
										Body:encodeRespArray([][]byte{}),
										Type:ARRAY,
							}
				 }

				 
				 return Response{
					     
					      Body:encodeRespArray(elements[startIndex:endIndex+1]),
							Type:ARRAY,
				  }
		 
		}


	return Response{
					      Body:encodeRespArray([][]byte{}),
							Type:ARRAY,
				  }
}