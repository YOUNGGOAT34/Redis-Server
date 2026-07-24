package aof

import (
	"fmt"
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


var Sequence int


func (aofConfig *AOF) CreateAOFDir() error{

	    if aofConfig.AppendOnly=="no"{
			 return nil
		 }
	    
	    aofDir:=filepath.Join(aofConfig.Dir,aofConfig.AppendDirName)

		 if err:=os.MkdirAll(aofDir,0755);err!=nil{
			  return err
		 }

		 Sequence=1

		 aofFileName:=buildAOFFileName(aofConfig.AppendFilename)


		 aofPath:=filepath.Join(aofDir,aofFileName)

		 aofFile,err:=os.OpenFile(aofPath,os.O_CREATE|os.O_WRONLY|os.O_APPEND,0644)

		 if err!=nil{
			  return err
		 }

		 aofFile.Close()


		 manifestPath:=filepath.Join(aofDir,buildManifestFileName(aofConfig.AppendFilename))

		 return os.WriteFile(manifestPath,[]byte(fmt.Sprintf("file %s sequence %d type i",aofFileName,Sequence)),0644)
}


func buildAOFFileName(baseName string) string{
	    return fmt.Sprintf("%s.%d.incr.aof",baseName,Sequence)
}

func buildManifestFileName(aofFilename string) string{
	   return fmt.Sprintf("%s.manifest",aofFilename)
}

