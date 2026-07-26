package server

import (
	"CacheDB/app/RESP"
	"CacheDB/app/config"
	"CacheDB/app/storage"
	"fmt"
	"net"
	"strconv"
)

func encodeSubResponse(channel []byte, count int,isSub bool) []byte {
	var resp []byte
	resp = fmt.Appendf(resp, "*3\r\n")
	if isSub{
		resp = fmt.Appendf(resp, "$%d\r\nsubscribe\r\n", len("subscribe"))

	}else{
		resp = fmt.Appendf(resp, "$%d\r\nunsubscribe\r\n", len("unsubscribe"))
	}
	resp = fmt.Appendf(resp, "$%d\r\n%s\r\n", len(channel), channel)
	resp = fmt.Appendf(resp, ":%d\r\n", count)
	return resp
}

func sub(replConfig *config.SERVER,client *storage.Client, args [][]byte) RESP.Response {
	if len(args) < 1{
		return RESP.WrongNumberOfArguments("SUBSCRIBE")
	}

	replConfig.PubSub.ChannelMutex.Lock()
	defer replConfig.PubSub.ChannelMutex.Unlock()
   
	client.InSubscribeMode=true
  
   var response []byte

	for _,channel :=range args{
		    
		   channelName:=string(channel)

		   if _,exists:=replConfig.PubSub.Channels[channelName];!exists{
				       replConfig.PubSub.Channels[channelName]=storage.NewSet[net.Conn]()
			}
			replConfig.PubSub.Channels[channelName].Add(client.Conn)
			client.SubscribedChannels.Add(channelName)
			
			response = append(response, encodeSubResponse(channel,client.SubscribedChannels.Len(),true)...)
	}

	

	return RESP.Response{
		Body: response,
		Type: RESP.ARRAY,
	}
}



func unSub(replConfig *config.SERVER,client *storage.Client, args [][]byte) RESP.Response{

	  if len(args)<1{
		       return RESP.WrongNumberOfArguments("UNSUBSCRIBE")
	  }
	   
	  replConfig.PubSub.ChannelMutex.Lock()
	  defer replConfig.PubSub.ChannelMutex.Unlock()

	  var response []byte 

	  for _,channel:=range args{
		      channelName:=string(channel)

				client.SubscribedChannels.Remove(channelName)

				if client.SubscribedChannels.Len()==0{
					   client.InSubscribeMode=false
				}

				  subscribers:=replConfig.PubSub.Channels[channelName]
						subscribers.Remove(client.Conn)
						if subscribers.Len()==0{
							   delete(replConfig.PubSub.Channels,channelName)
					}
       
		     response = append(response, encodeSubResponse(channel,client.SubscribedChannels.Len(),false)...)		
	  }

	  return  RESP.Response{
		  Body: response,
		  Type: RESP.ARRAY,
	  }
}



func pub(replConfig *config.SERVER,args [][]byte,) RESP.Response{
	    if len(args)!=2{
			  return RESP.WrongNumberOfArguments("PUBLISH")
		 }

	
	//copy the subscribed clients to avoid holding the lock while sending messages
	
   replConfig.PubSub.ChannelMutex.RLock()
	subscribedclients:=make([]net.Conn,0,replConfig.PubSub.Channels[string(args[0])].Len())

	for client:=range replConfig.PubSub.Channels[string(args[0])]{
		    subscribedclients = append(subscribedclients, client)
	} 
	replConfig.PubSub.ChannelMutex.RUnlock()

   
	 response:=RESP.EncodeArray([][]byte{[]byte("message"),args[0],args[1]})

	 //keep track of the subscribers who got messages

	 count:=0

    for _,client:=range subscribedclients{
		     _,err:=client.Write(response)

			  if err!=nil{
						channel:=string(args[0])
					//remove dead clients
				      replConfig.PubSub.ChannelMutex.Lock()
				      subscribers:=replConfig.PubSub.Channels[channel]
						subscribers.Remove(client)
						if subscribers.Len()==0{
							   delete(replConfig.PubSub.Channels,channel)
						}
						replConfig.PubSub.ChannelMutex.Unlock()


						continue	
			  }

			  count+=1
	 }
	

	return RESP.Response{
		      Body: []byte(strconv.FormatInt(int64(count),10)),
				Type: RESP.INTEGER,
	}
		 
}

//determines whether a command is legal in subscribe mode
func isLegal(command string) bool{

	  switch command{
	       case "SUBSCRIBE","UNSUBSCRIBE", "PSUBSCRIBE" ,"PUNSUBSCRIBE" , "PING" , "QUIT":
				return true
	  }
	return false
	 
}
