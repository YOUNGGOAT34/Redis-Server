package server

import (
	"strconv"
	"time"
)



func getCommand(arguments [][]byte) Response {
	   if len(arguments)<1{
			  return Response{
				Body:[]byte("Wrong number of arguments for 'GET' command"),
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
			     
			     if dataObject.Type!=STRING{
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
			   return wrongNumberOfArguments("SET")
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
				 }else{

					    return Response{

										     Body:[]byte("Syntax error:Unknown option"),
											  Type:ERROR,
									}
					   
				 }
		}



		databaseMutex.Lock()
		database[string(arguments[0])]=Data{
			   Type: STRING,
				Value: string(arguments[1]),
		}
		databaseMutex.Unlock()

		return Response{
			   Body:[]byte("OK"),
				Type:SIMPLE_STRING,
		  }

} 






