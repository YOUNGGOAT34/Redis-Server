package replication

import "CacheDB/app/helpers"


func InfoCommand(args [][]byte) helpers.Response{
	   if len(args)>0{
			   if helpers.CompareBytes(args[0],[]byte("replication")){
					    
					     return helpers.Response{
									Body: []byte("#Replication\r\nrole:master\r\n"),
									Type: helpers.BULK_STRING,
						  }
				}
		}
       
		return helpers.Response{}
}