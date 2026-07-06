package server

import (
	
	"strconv"
)


func wakeUpWaitingClients(key string,values *[][]byte){
	      blockedClientsMutex.Lock() 
			

			for len(*values)>0{


				   q,ok:=blockedClients[key]

					if!ok{

						 break
					}
				   
				   if front:=q.Front(); front!=nil{
						   ch:=front.Value.(chan []byte)
							q.Remove(front)
							if q.Len()==0{
								   delete(blockedClients,key)
							}
							blockedClientsMutex.Unlock()
							res:=(*values)[0]
							*values=(*values)[1:]
							ch<-res
                  
					}else{
					    delete(blockedClients,key)
						  break
					} 


					blockedClientsMutex.Lock()
					
					
			}

			blockedClientsMutex.Unlock()
}


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
		    return wrongNumberOfArguments("RPUSH")
	}


	values:=arguments[1:]

	databaseMutex.Lock()
	data,exists:=database[string(key)]
	databaseMutex.Unlock()
	
	

	if exists{

		   if data.Type!=LIST{
				  
				  return wrongType()
			}

			list:=data.Value.(*List)
			list.listMutex.Lock()
			defer list.listMutex.Unlock()

			wakeUpWaitingClients(string(arguments[0]),&values)


			for _,value:=range values{
							list.PushBack(value)
			}

			
			var buf [32]byte
			return Response{
		  
		   Body:strconv.AppendInt(buf[:0],int64(list.len),10),
			Type:INTEGER,
			
	}
			 
	}



	wakeUpWaitingClients(string(arguments[0]),&values)



	if len(values)==0{
		  var buf [32]byte

		  return Response{
				
					Body:strconv.AppendInt(buf[:0],0,10),
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