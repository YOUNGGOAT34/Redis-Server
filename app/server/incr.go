package server

import (
	"CacheDB/app/helpers"
	"math"
	"strconv"
)

func incrCommand(arguments [][]byte,client *Client) helpers.Response {
	
	   if len(arguments)!=1{
			  return wrongNumberOfArguments("INCR")
		}


		key:=string(arguments[0])

		databaseMutex.Lock()
		defer databaseMutex.Unlock()

		var intValue int64
		var err error

		data,exists:=database[key];
		if exists{
			   if data.Type!=STRING{
					   return wrongType()
				}

				value:=data.Value.(string)

				intValue,err=strconv.ParseInt(value,10,64)

				if err!=nil{
					  return helpers.Response{
						   Body: []byte("ERR value is not an integer or out of range"),
							Type: helpers.ERROR,
					  }
				}else{

					  if intValue==math.MaxInt64{
						   return helpers.Response{
						   Body: []byte("ERR increment or decrement would overflow"),
							Type: helpers.ERROR,
					  }
					  }
					  
					  intValue++
				}


		}else{

			 intValue=1
			  
			
		}

		strValue:=strconv.FormatInt(intValue,10)

		database[key]=Data{
			   Type: STRING,
				Value:strValue,
		}

		markDirty(key,client)

		return helpers.Response{
			   Body:[]byte(strValue),
				Type:helpers.INTEGER,
		  }
}


