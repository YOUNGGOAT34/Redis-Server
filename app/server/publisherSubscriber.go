package server

import (
	"CacheDB/app/RESP"
	"CacheDB/app/config"
	"CacheDB/app/storage"
	"fmt"
	"net"
	"strconv"
)

func encodeSubResponse(channel []byte, count int) []byte {
	var resp []byte
	resp = fmt.Appendf(resp, "*3\r\n")
	resp = fmt.Appendf(resp, "$%d\r\nsubscribe\r\n", len("subscribe"))
	resp = fmt.Appendf(resp, "$%d\r\n%s\r\n", len(channel), channel)
	resp = fmt.Appendf(resp, ":%d\r\n", count)
	return resp
}

func sub(serverConfig *config.SERVER,client *storage.Client, args [][]byte) RESP.Response {
	if len(args) < 2 {
		return RESP.WrongNumberOfArguments("SUBSCRIBE")
	}

	serverConfig.PubSub.ChannelMutex.Lock()
   
	if !client.InSubscribeMode{

		client.InSubscribeMode=true
	}


   var response []byte

	for _,channel :=range args[1:]{
		    
		   channelName:=string(channel)

		   if _,exists:=serverConfig.PubSub.Channels[channelName];!exists{
				       serverConfig.PubSub.Channels[channelName]=storage.NewSet[net.Conn]()
			}
			serverConfig.PubSub.Channels[channelName].Add(client.Conn)
			client.SubscribedChannels.Add(channelName)
			
			response = append(response, encodeSubResponse(channel,client.SubscribedChannels.Len())...)
	}

	serverConfig.PubSub.ChannelMutex.Unlock()



	return RESP.Response{
		Body: response,
		Type: RESP.ARRAY,
	}
}



func pub(serverConfig *config.SERVER,args [][]byte,) RESP.Response{
	    if len(args)!=2{
			  return RESP.WrongNumberOfArguments("PUBLISH")
		 }

	
	//copy the subscribed clients to avoid holding the lock while sending messages
	
   serverConfig.PubSub.ChannelMutex.RLock()
	subscribedclients:=make([]net.Conn,0,serverConfig.PubSub.Channels[string(args[0])].Len())

	for client:=range serverConfig.PubSub.Channels[string(args[0])]{
		    subscribedclients = append(subscribedclients, client)
	} 
	serverConfig.PubSub.ChannelMutex.RUnlock()

   
	 response:=RESP.EncodeArray([][]byte{[]byte("message"),args[0],args[1]})

    for _,client:=range subscribedclients{
		     _,err:=client.Write(response)

			  if err!=nil{

			  }
	 }
	

	return RESP.Response{
		      Body: []byte(strconv.FormatInt(int64(len(subscribedclients)),10)),
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
