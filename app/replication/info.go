package replication

import (
	"CacheDB/app/helpers"
	"fmt"
)


func InfoCommand(args [][]byte,config *helpers.SERVER) helpers.Response{
	   if len(args)>0{
			   if helpers.CompareBytes(args[0],[]byte("replication")){
					     res:=fmt.Sprintf("#Replication\r\nrole:%s\r\n",config.Role)
					     return helpers.Response{
									Body: []byte(res),
									Type: helpers.BULK_STRING,
						  }
				}
		}
       
		return helpers.Response{}
}