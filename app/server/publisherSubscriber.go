package server

import (
	"CacheDB/app/RESP"
	"CacheDB/app/config"
	"CacheDB/app/storage"
	"fmt"
	"net"
)

func encodePubSubResponse(channel []byte, count int) []byte {
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

	for _,channel :=range args[1:]{
		  
		   if _,exists:=serverConfig.PubSub.Channels[string(channel)];!exists{
				       serverConfig.PubSub.Channels[string(channel)]=storage.NewSet[net.Conn]()
			}
			serverConfig.PubSub.Channels[string(channel)].Add(client.Conn)
			client.SubscribedChannels.Add(string(channel))  
	}

	serverConfig.PubSub.ChannelMutex.RUnlock()

	if !client.InSubscribeMode{

		client.InSubscribeMode=true
	}


	return RESP.Response{
		Body: encodePubSubResponse(args[1], client.SubscribedChannels.Len()),
		Type: RESP.ARRAY,
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
