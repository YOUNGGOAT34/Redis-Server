package replication

import (
	"CacheDB/app/RESP"
	"fmt"
)

func PropagateCommands(parsedRequest []byte,config *RESP.SERVER){
	
	  config.ReplicasMutex.Lock()
	  defer config.ReplicasMutex.Unlock()

	  for i:=0;i<len(config.REPLICAS);{
		     fmt.Printf("%q\n",parsedRequest)
		      _,err:=config.REPLICAS[i].Write(parsedRequest)

				if err!=nil{
					  //if the write fails remove the replica
					  config.REPLICAS[i].Close()
					  config.REPLICAS = append(config.REPLICAS[:i],config.REPLICAS[i+1:]...)
					  continue
				}

				i++
	  }

}