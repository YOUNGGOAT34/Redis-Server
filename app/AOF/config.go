package aof

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type AOF struct{
	   Dir string
		AppendOnly string
		AppendDirName string
		AppendFilename string
		AppendFsync string
		Sequence int
		File *os.File
}


func (aofConfig *AOF) CreateAOFDir() error{

	    if aofConfig.AppendOnly=="no"{
			 return nil
		 }
	    
	    aofDir:=filepath.Join(aofConfig.Dir,aofConfig.AppendDirName)

		 if err:=os.MkdirAll(aofDir,0755);err!=nil{
			  return err
		 }

		 manifestPath:=filepath.Join(aofDir,BuildManifestFileName(aofConfig.AppendFilename))

		 if _,err:=os.Stat(manifestPath);err==nil{
			 filename,err:=ReadManifest(manifestPath)

			 if err!=nil{
				 return err
			 }
			  aofPath:=filepath.Join(aofDir,filename)

			file,err:=os.OpenFile(aofPath,os.O_CREATE|os.O_WRONLY|os.O_APPEND,0644)

			if err!=nil{
				 return err
			}

			 aofConfig.File=file
			 return nil

		 }

		 aofConfig.Sequence=1

		 aofFileName:=aofConfig.buildAOFFileName(aofConfig.AppendFilename)


		 aofPath:=filepath.Join(aofDir,aofFileName)

		 aofFile,err:=os.OpenFile(aofPath,os.O_CREATE|os.O_WRONLY|os.O_APPEND,0644)

		 if err!=nil{
			  return err
		 }

		 aofConfig.File=aofFile

		 return os.WriteFile(manifestPath,fmt.Appendf(nil, "file %s type i sequence %d\n",aofFileName,aofConfig.Sequence),0644)
}


func (aofConfig *AOF) buildAOFFileName(baseName string) string{
	    return fmt.Sprintf("%s.%d.incr.aof",baseName,aofConfig.Sequence)
}

func BuildManifestFileName(aofFilename string) string{
	   return fmt.Sprintf("%s.manifest",aofFilename)
}


func ReadManifest(manifestPath string) (string,error){
	  data,err:=os.ReadFile(manifestPath)
			 if err!=nil{
				 return "",err
			 }

		parts:=strings.Fields(string(data))

		if len(parts)<2{
			  return "",errors.New("invalid manifest")
		}

		if parts[0]!="file"{
			  return "",errors.New("invalid manifest format")
		}

		return parts[1],nil
}




