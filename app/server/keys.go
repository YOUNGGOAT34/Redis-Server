package server

import "CacheDB/app/RESP"

func keys(args [][]byte) RESP.Response{
	    if len(args)!=1{
			   return RESP.WrongNumberOfArguments("KEYS")
		 }


		 
		 if RESP.CompareBytes(args[0],[]byte("*")){
			 keys:=make([][]byte,0,len(database))
			 databaseMutex.RLock()
			 for key := range database{
					keys = append(keys, []byte(key))
			 }
	
			 databaseMutex.RUnlock()
			   return RESP.Response{
					    Body: RESP.EncodeArray(keys),
						 Type: RESP.ARRAY,
				}
		 }


		
       count:=0
		 exists,index:=hasWildCard(args[0])

		
		 if exists{
                
			      databaseMutex.RLock()
		         
			   
			      for key:=range database{
						   if startsWith([]byte(key),args[0][:index]){
								  count++
							}
					}
					matchingKeys:=make([][]byte,0,count)
				 
					if exists,index:=hasWildCard(args[0]);exists{
							  for strkey:=range database{
								      key:=[]byte(strkey)
									  if startsWith(key,args[0][:index]){
											 matchingKeys = append(matchingKeys, key)
									  }
							  }
					}

					databaseMutex.RUnlock()

					 return RESP.Response{
							Body: RESP.EncodeArray(matchingKeys),
							Type: RESP.ARRAY,
		           }
		 }




	 
		


	return RESP.Response{}
		
}

func startsWith(key []byte, pattern []byte) bool {
	if len(key)<len(pattern){
		   return false
	}

	for i:=0;i<len(pattern);i++{
		  if pattern[i]!=key[i]{
			  return false
		  }
		  
	}

	return true
}