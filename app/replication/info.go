package replication

import (
	"CacheDB/app/helpers"
	"fmt"
)


func InfoCommand(args [][]byte,config *helpers.SERVER) helpers.Response{
	   if len(args)>0{
			   if helpers.CompareBytes(args[0],[]byte("replication")){
					     res:=fmt.Sprintf("# Replication\r\nrole: %s\r\nmaster_replid: %s\r\nmaster_repl_offset: %d\r\n",config.Role,config.MASTERREPLID,config.MASTERREPLOFFSET)
					     return helpers.Response{
									Body: []byte(res),
									Type: helpers.BULK_STRING,
						  }
				}
		}
       
		return helpers.Response{}
}