package replication

import (

	"CacheDB/app/config"
)

func PropagateCommands(parsedRequest []byte,serverConfig *config.SERVER){
	
	  serverConfig.ReplicasMutex.RLock()
	  replicas:=append([]*config.REPLICA(nil),serverConfig.REPLICAS...)
	  serverConfig.ReplicasMutex.RUnlock()


	  for _,replica:=range replicas{
		     
		      _,err:=replica.Conn.Write(parsedRequest)

				if err!=nil{
					  //if the write fails remove the replica
					  serverConfig.ReplicasMutex.Lock()
					  for i,r:=range serverConfig.REPLICAS{
						    if r==replica{

								 
								 serverConfig.REPLICAS[i].Conn.Close()
								 serverConfig.REPLICAS = append(serverConfig.REPLICAS[:i],serverConfig.REPLICAS[i+1:]...)
								 break
							 }
					  }

					  serverConfig.ReplicasMutex.Unlock()
					 
				}

		
	  }

}