package server

import (
	"CacheDB/app/helpers"
	"fmt"
)

//     |----------------------MULTI COMMAND----------------------|

func multiCommand(arguments [][]byte,client *Client) helpers.Response {
	 if len(arguments)!=0{
		   return wrongNumberOfArguments("MULTI")
	 }

	 if client.InTransaction{
		     return helpers.Response{
                 Body: []byte("ERR MULTI calls cannot be nested"),
					  Type: helpers.ERROR,
			  }
	 }

	 client.InTransaction=true

	 return helpers.Response{
		   Body: []byte("OK"),
			Type: helpers.SIMPLE_STRING,
	 }
}


//     |----------------------EXEC COMMAND----------------------|

func execCommand(arguments [][]byte,client *Client,config *helpers.SERVER) helpers.Response{
      
	if len(arguments)!=0{
		  return wrongNumberOfArguments("EXEC")
	}

	//exec executed without multi
    if !client.InTransaction{
		   return helpers.Response{
			      Body: []byte("ERR EXEC without MULTI"),
					Type: helpers.ERROR,
		  }
	 }

	 queued:=client.Queue

	 client.Queue=nil
	 client.InTransaction=false
	 defer clearWatches(client)
    
	 if client.Dirty{
		  
			 return helpers.Response{
				 Body: []byte("*-1\r\n"),
             Type: helpers.ARRAY,
			 }
	 }

	 var resp []byte
	 resp=fmt.Appendf(resp,"*%d\r\n",len(queued))

	 for _,cmd:=range queued{
		     r:=dispatchCommands(client,cmd.Args,config)
			  resp = append(resp, buildResponse(r)...)
	 }

	
	return  helpers.Response{
		   Body: resp,
			Type: helpers.ARRAY,
	}
}


//     |----------------------DISCARD COMMAND----------------------|

func discardCommand(arguments [][]byte, client *Client) helpers.Response {
	    if len(arguments)!=0{
			   return wrongNumberOfArguments("DISCARD")
		 }

		 if !client.InTransaction{
			    return helpers.Response{
					  Body: []byte("ERR DISCARD without MULTI"),
					  Type: helpers.ERROR,
				 }
		 }

		 client.InTransaction=false
		 client.Queue=nil
       
	   clearWatches(client)
		 return helpers.Response{
			    Body: []byte("OK"),
				 Type: helpers.SIMPLE_STRING,
		 }
}

//     |----------------------WATCH COMMAND----------------------|

func watchCommand(arguments [][]byte,client *Client) helpers.Response{
	   if len(arguments)<1{
			  return wrongNumberOfArguments("WATCH")
		}

		if client.InTransaction{
			  return helpers.Response{ 
				    Body: []byte("ERR WATCH inside MULTI is not allowed"),
					 Type: helpers.ERROR,
			  }
		}

		for _,argument:=range arguments{

			key:=string(argument)
	
			watchedKeysMutex.Lock()
			set,exists:=watchedKeys[key]
			
			if !exists{
				  set=make(map[*Client]struct{})
				  watchedKeys[key]=set
				 
	
			}
			 
			set[client]=struct{}{}
	
			watchedKeysMutex.Unlock()
	
			client.keysWatched[key]=struct{}{}
		}
      
		return helpers.Response{
			 Body: []byte("OK"),
			 Type: helpers.SIMPLE_STRING,
		}
}


func unwatchCommand(arguments [][]byte,client *Client) helpers.Response{
	    if len(arguments)!=0{
			  return wrongNumberOfArguments("UNWATCH")
		 }

		 clearWatches(client)

		 return helpers.Response{
			  Body: []byte("OK"),
			  Type: helpers.SIMPLE_STRING,
		 }
}


func clearWatches(client *Client){
	   watchedKeysMutex.Lock()
		
	  for key:=range client.keysWatched{
		     delete (watchedKeys[key],client)
			  if len(watchedKeys[key])==0{
				   delete(watchedKeys,key)
			  }
	  }

	  watchedKeysMutex.Unlock()

	  client.keysWatched=make(map[string]struct{})
	  client.Dirty=false
}

