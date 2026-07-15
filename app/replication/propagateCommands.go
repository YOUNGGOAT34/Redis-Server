package replication

import (
	"CacheDB/app/RESP"
	
)

func PropagateCommands(parsedRequest []byte,config *RESP.SERVER){
	
	  config.ReplicasMutex.RLock()
	  replicas:=append([]*RESP.REPLICA(nil),config.REPLICAS...)
	  config.ReplicasMutex.RUnlock()


	  for _,replica:=range replicas{
		     
		      _,err:=replica.Conn.Write(parsedRequest)

				if err!=nil{
					  //if the write fails remove the replica
					  config.ReplicasMutex.Lock()
					  for i,r:=range config.REPLICAS{
						    if r==replica{

								 config.ReplicasMutex.Unlock()
								 config.REPLICAS[i].Conn.Close()
								 config.REPLICAS = append(config.REPLICAS[:i],config.REPLICAS[i+1:]...)
								 break
							 }
					  }
					 
				}

		
	  }

}