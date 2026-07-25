package server

import (
	"CacheDB/app/AOF"
	"CacheDB/app/RDB"
	"CacheDB/app/RESP"
	"CacheDB/app/storage"
	"errors"
	"io"
	"os"
)

func replayAOF(file *os.File,replConfig *RESP.SERVER, rdbConfig *rdb.RDB,aofConfig *aof.AOF) error{
     
	     request:=make([]byte,0,1024)
		  temp:=make([]byte,1024)


		  for{

			    bytesRead,err:=file.Read(temp)

				 if err!=nil && err!=io.EOF{
						return err
				 }

              if bytesRead>0{

					  request = append(request, temp[:bytesRead]...)
				  }

				

				 for len(request)>0{
      
					    parsedRequest,bytesConsumed,err:=RESP.ParseRequest(request)

						 if err!=nil{
							    if errors.Is(err,RESP.ErrIncomplete){
									  break
								 }

								 return err
						 }

						 dispatchCommands(&storage.Client{},parsedRequest,replConfig,rdbConfig,aofConfig)

						 request=request[bytesConsumed:]
					     
				 }


				 if err==io.EOF{

					   if len(request)>0{
							  return errors.New("Incomplete command at the end of AOF")
						}
					   break
				 }


		  }


		  return nil

}