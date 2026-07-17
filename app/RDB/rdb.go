package rdb

import (
	"CacheDB/app/RESP"
	"fmt"
	"io"
	"os"
)

var EmptyRDB = []byte{
    0x52, 0x45, 0x44, 0x49, 0x53, 0x30, 0x30, 0x31,
    0x31, 0xfa, 0x09, 0x72, 0x65, 0x64, 0x69, 0x73,
    0x2d, 0x76, 0x65, 0x72, 0x05, 0x37, 0x2e, 0x32,
    0x2e, 0x30, 0xfa, 0x0a, 0x72, 0x65, 0x64, 0x69,
    0x73, 0x2d, 0x62, 0x69, 0x74, 0x73, 0xc0, 0x40,
    0xfa, 0x05, 0x63, 0x74, 0x69, 0x6d, 0x65, 0xc2,
    0x6d, 0x08, 0xbc, 0x65, 0xfa, 0x08, 0x75, 0x73,
    0x65, 0x64, 0x2d, 0x6d, 0x65, 0x6d, 0xc2, 0xb0,
    0xc4, 0x10, 0x00, 0xfa, 0x08, 0x61, 0x6f, 0x66,
    0x2d, 0x62, 0x61, 0x73, 0x65, 0xc0, 0x00, 0xff,
    0xf0, 0x6e, 0x3b, 0xfe, 0xc0, 0xff, 0x5a, 0xa2,
}



func readByte(data []byte,pos *int) (byte,error){
      if *pos>=len(data){
           return 0,io.ErrUnexpectedEOF
      }

      value:=data[*pos]

      (*pos)++

      return value,nil
}


func readHeader(data []byte,pos *int) ([]byte,error){

       if *pos+9>len(data){
           return nil,io.ErrUnexpectedEOF
       }
       
     
       header:=data[*pos:*pos+9]

       *pos+=9
      
      return header,nil
}


func readRdbFile(rdbConfig RDB){

     //cursor position
     pos:=0

      data,err:=os.ReadFile(rdbConfig.Dir+"/"+rdbConfig.DbFileName)

      header,err:=readHeader(data,&pos)
      
      if err!=nil{
           
      }

      if   !RESP.CompareBytes(header[:5],[]byte("REDIS")){
           fmt.Printf("Not an rdb file\r\n");
      }


     

      loop:
      for{
          
           opcode,err:=readByte(data,&pos)
           
           if err!=nil{
              //handle error
           }

           /*
               0xFA-->auxilary field
               0xFB-->database size
               0XFE--->database selector
               0xFF--->end of file
           
           */

           switch opcode{
                 case 0xFA:
                    auxiliaryKey,auxiliaryValue,err:=parseAuxilarySection(data,&pos)
                    if err!=nil{
                          //handle error
                    }


                    fmt.Printf("key=%s,value=%s\r\n",auxiliaryKey,auxiliaryValue)

                 case 0xFB:
                 case 0xFE:
                 case 0xFF:
                    break loop
                  

           }
      }


}

func parseAuxilarySection(data []byte, pos *int) ([]byte, []byte, error) {
	  if *pos>=len(data){
           return EOF()
      }

      keyLength:=data[*pos]

      (*pos)++

      if *pos+int(keyLength)>=len(data){
          return EOF()
      }

      key:=data[*pos:*pos+int(keyLength)]

      *pos+=int(keyLength)

      if *pos>len(data){
          return EOF()
      }

      valueLength:=data[*pos]
      (*pos)++

      if *pos+int(valueLength)>len(data){
          return EOF()
      }

      value:=data[*pos:*pos+int(valueLength)]

      *pos+=int(valueLength)

      return key,value,nil

}


func EOF() ([]byte,[]byte,error){
      return nil,nil,io.ErrUnexpectedEOF
}


