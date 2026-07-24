package aof

import (
	"os"
	"path/filepath"
)

type AOF struct{
	   Dir string
		AppendOnly string
		AppendDirName string
		AppendFilename string
		AppendFsync string
}



func (aofConfig *AOF) CreateAOFDir() error{

	    if aofConfig.AppendOnly=="no"{
			 return nil
		 }
	    
	    aofDir:=filepath.Join(aofConfig.Dir,aofConfig.AppendDirName)

		 if err:=os.MkdirAll(aofDir,0755);err!=nil{
			  return err
		 }

		 aofPath:=filepath.Join(aofDir,aofConfig.AppendFilename)

		 file,err:=os.OpenFile(aofPath,os.O_CREATE|os.O_WRONLY|os.O_APPEND,0644)

		 if err!=nil{
			  return err
		 }

		 file.Close()
		

		 return nil
}

