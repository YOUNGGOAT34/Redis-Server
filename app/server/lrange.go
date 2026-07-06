package server

import (

	"fmt"
	"os"
	"strconv"
)


func encodeRespArray(list *List,startIndex int,endIndex int) []byte{

	 var respArray []byte

	 count:=endIndex-startIndex+1

	 respArray=fmt.Appendf(respArray,"*%d\r\n",count)  
	  
	 /*
	     To find the starting node 
		  A small optimization is:
		  if the distance between the head and startIndex is less than 
		  the distance between Tail and startIndex
		  find the startIndex by traversing the list from head 
		  otherwise find the start index by traversing from the tail backwards since it is a doubly linked list
	 */
	 
    currentNode:=list.Head
    currentIndex:=0

	 if startIndex<=(list.len-1)-startIndex{
        
			 for currentNode!=nil && currentIndex!=startIndex{
				      currentNode=currentNode.Next
				      currentIndex+=1
			 }

	 }else{

		    currentIndex=list.len-1
			 currentNode=list.Tail
			 for currentNode!=nil && currentIndex!=startIndex{
				      currentNode=currentNode.Prev
				      currentIndex--
			 }
		   
	 }


	 for i:=startIndex;i<=endIndex;i++{
		    element:=currentNode.data
			 
			 respArray=fmt.Appendf(respArray,"$%d\r\n%s\r\n",len(element),element)
			 currentNode=currentNode.Next
	 }
	 return respArray
}


func lRangeCommand(arguments [][]byte) Response {

	   
	   if len(arguments)!=3{
			      return  wrongNumberOfArguments("LRANGE")
		
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
		data,exists:=database[key]
		databaseMutex.RUnlock()

		if exists{
           
			    if data.Type!=LIST{
								return wrongType()
				 }

				  
				 list:=data.Value.(*List)

				 list.listMutex.RLock()
				 defer list.listMutex.RUnlock()
             
				 if list==nil || list.len==0{
					    return Response{
										Body:[]byte("*0\r\n"),
										Type:ARRAY,
							}
				 }


				 if startIndex<0{
					   startIndex=list.len+startIndex
				 }

				
				 if startIndex>=list.len {
					            
									return Response{
										Body:[]byte("*0\r\n"),
										Type:ARRAY,
							}

							
				 }

				 
				 if endIndex>=list.len{
					     endIndex=list.len-1
				 }
      
				
				 if endIndex<0{
					   endIndex=list.len+endIndex
				 }


				 if startIndex<0{
					   startIndex=0
				 }

				 if endIndex<0{
					   endIndex=0
				 }


				 if startIndex>endIndex{
					     return Response{
										Body:[]byte("*0\r\n"),
										Type:ARRAY,
							}
				 }

				 
				 return Response{
					     
					      Body:encodeRespArray(list,startIndex,endIndex),
							Type:ARRAY,
				  }
		 
		}


	return Response{
					Body:[]byte("*0\r\n"),
					Type:ARRAY,
							}
}