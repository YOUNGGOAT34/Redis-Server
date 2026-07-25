package server

import (
	"CacheDB/app/AOF"
	"CacheDB/app/RDB"
	"CacheDB/app/RESP"
	"CacheDB/app/storage"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func replayAOF(replConfig *RESP.SERVER, rdbConfig *rdb.RDB,aofFileConfig *aof.AOF) error{


	      aofDir:=filepath.Join(aofFileConfig.Dir,aofFileConfig.AppendDirName)
            
			  manifestPath:=filepath.Join(aofDir,aof.BuildManifestFileName(aofFileConfig.AppendFilename))

			  aoffilename,err:=aof.ReadManifest(manifestPath)

			  if err!=nil{
				   fmt.Fprintf(os.Stderr,"%s\r\n",err.Error())
			  }

			  aofPath:=filepath.Join(aofDir,aoffilename)

			  file,err:=os.Open(aofPath)

			  if err!=nil{
				  return err
			  }
     
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

						 dispatchCommands(&storage.Client{},parsedRequest,replConfig,rdbConfig,aofFileConfig)

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