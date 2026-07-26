package server

import (
	"CacheDB/app/RESP"
	"CacheDB/app/storage"
	"fmt"
)

func encodePubSubResponse(channel []byte, count int) []byte {
	var resp []byte
	resp = fmt.Appendf(resp, "*3\r\n")
	resp = fmt.Appendf(resp, "$%d\r\nsubscribe\r\n", len("subscribe"))
	resp = fmt.Appendf(resp, "$%d\r\n%s\r\n", len(channel), channel)
	resp = fmt.Appendf(resp, ":%d\r\n", count)
	return resp
}

func sub(client *storage.Client, args [][]byte) RESP.Response {
	if len(args) < 2 {
		return RESP.WrongNumberOfArguments("SUBSCRIBE")
	}


	for _,channel :=range args[1:]{
		
			client.SubscribedChannels.Add(string(channel))  
	}

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
