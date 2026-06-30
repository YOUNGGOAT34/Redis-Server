package server

import (
	
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

	databaseMutex.Lock()
	defer databaseMutex.Unlock()

	data,exists:=database[string(key)]

	if exists{


		   

		   if data.Type!=LIST{
				  return Response{
					      Body:[]byte("WRONGTYPE Operation against a key holding the wrong kind of value"),
							Type:ERROR,
				  }
			}

			list:=data.Value.(*List)

         
			for _,value:=range values{
				   list.PushBack(value)
			}


			var buf [32]byte

			return Response{
		  
		   Body:strconv.AppendInt(buf[:0],int64(list.len),10),
			Type:INTEGER,
			
	}
			 
	}


	node:=&Node{
        data:values[0],
	}

	list :=&List{ 
		   Head:node,
			Tail: node,
			len:1,
	}


  

	for _,value:=range values[1:]{
		  list.PushBack(value)
	}

	var dataObject Data
	dataObject.Type=LIST
	dataObject.Value=list

	database[string(key)]=dataObject
	
	var buf [32]byte

   return Response{
		  
		   Body:strconv.AppendInt(buf[:0],int64(list.len),10),
			Type:INTEGER,
			
	}
   
}