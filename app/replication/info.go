package replication

import (
	"CacheDB/app/RESP"
	"fmt"
)

func InfoCommand(args [][]byte, config *RESP.SERVER) RESP.Response {
	if len(args) > 0 {
		if RESP.CompareBytes(args[0], []byte("replication")) {
			res := fmt.Sprintf("# Replication\r\nrole: %s\r\nmaster_replid: %s\r\nmaster_repl_offset: %d\r\n", config.Role, config.MASTERREPLID, config.MASTERREPLOFFSET.Load())
			return RESP.Response{
				Body: []byte(res),
				Type: RESP.BULK_STRING,
			}
		}
	}
    
	return RESP.Response{}
}
