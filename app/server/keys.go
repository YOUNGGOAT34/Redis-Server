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

		 exists,index:=hasWildCard(args[0],'*')

		 if exists{
                
			       prefix:=string(args[0][:index])
			     
					matchingKeys:=collectMatchingKeys(func (key string) bool{
						    return startsWith(key,prefix)
					})

					 return RESP.Response{
							Body: RESP.EncodeArray(matchingKeys),
							Type: RESP.ARRAY,
		           }
		 }


	exists,index=hasWildCard(args[0],'?')

	if exists{
		       
		         
			      prefix:=string(args[0][:index])
			     
					matchingKeys:=collectMatchingKeys(func (key string) bool{
						    return startsWith(key,prefix) && len(prefix)+1==len(key)
					})
				 
					
					 return RESP.Response{
							Body: RESP.EncodeArray(matchingKeys),
							Type: RESP.ARRAY,
		           }
	}


	 
	return RESP.Response{}
		
}

func startsWith(key string, pattern string) bool {
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



func collectMatchingKeys(matches func(string) bool) [][]byte{
	  databaseMutex.RLock()
		         
			count:=0
			for key:=range database{
					if matches(key){
								count++
							
					}
			}
			matchingKeys:=make([][]byte,0,count)
			
			
				for key:=range database{
					
						if matches(key){
								matchingKeys = append(matchingKeys, []byte(key))
						}
				}
			

			databaseMutex.RUnlock()

			return  matchingKeys
}