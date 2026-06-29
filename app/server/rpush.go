package server

import (
	"fmt"
	"os"
	"strconv"
)


func rPushCommand(arguments [][]byte) Response {
	if len(arguments)==0{
		  return Response{
			    Body:[]byte("Wrong number of arguments for 'RPUSH' command"),
				 Type:ERROR,

		  }
	}

	/*

	  arguments[0]-->key
	  arguments[1:]--->value(s)

	*/

	key:=arguments[0]
	
	//in a case where there was a key but had no values
	if len(arguments)<2{
		    return Response{
			    Body:[]byte("Wrong number of arguments for 'RPUSH' command"),
				 Type:ERROR,

		  }
	}


	values:=arguments[1:]

	
   var data [][]byte

	databaseMutex.Lock()
	defer databaseMutex.Unlock()

	list,exists:=database[string(key)]

	if exists{


		   

		   if list.Type!="List"{
				  return Response{
					      Body:[]byte("WRONGTYPE Operation against a key holding the wrong kind of value"),
							Type:ERROR,
				  }
			}


			 var ok bool
		    data,ok=list.Value.([][]byte)
			 if !ok{
				   fmt.Fprintf(os.Stderr,"Database corruption: Expected [][]byte for a list\n")
				   return Response{
						 Body:[]byte("Internal error"),
						 Type:ERROR,
					}
			 }
			 data=append(data, values...)
			
			 
	}else {
		  
		  data=values
	}



	var dataObject Data
	dataObject.Type="List"
	
	dataObject.Value=data
	database[string(key)]=dataObject

	
	var buf [32]byte

   return Response{
		  
		   Body:strconv.AppendInt(buf[:0],int64(len(data)),10),
			Type:INTEGER,
			
	}
   
}