package replication

import "CacheDB/app/RESP"

func PropagateCommands(parsedRequest []byte,config *RESP.SERVER){
	
	  for i,replica :=range config.REPLICAS{
		      _,err:=replica.Write(parsedRequest)

				if err!=nil{
					  //if the write fails remove the replica
					  config.REPLICAS = append(config.REPLICAS[:i],config.REPLICAS[i+1:]...)
					  continue
				}
	  }

}