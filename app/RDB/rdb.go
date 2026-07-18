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


//helpers

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



func specialEncoding(data []byte, encoding byte,pos *int) (uint64, error) {

	  specialEncodingType:=encoding & 0x3F

      /*
          0-->read the next byte
          1--->read the next two bytes
          2--->read the next 4 bytes
          3--->LZF compressed string
      */

      switch specialEncodingType{
            case 0:
                if *pos>=len(data){
                      return 0,io.ErrUnexpectedEOF
                }

                value:=uint64(data[*pos])
                (*pos)++

                return value,nil

            case 1:
                if *pos+2>len(data){
                     return 0,io.ErrUnexpectedEOF
                }

                firstByte:=data[*pos]
                (*pos)++
                secondByte:=data[*pos]
                (*pos)++

                value:=(uint64(firstByte)<<8 | uint64(secondByte))

                return value,nil

            case 2:

                if *pos+4>len(data){
                     return 0,io.ErrUnexpectedEOF
                }


                firstByte:=data[*pos]
                (*pos)++
                secondByte:=data[*pos]
                (*pos)++
                thirdByte:=data[*pos]
                (*pos)++
                fourthByte:=data[*pos]
                (*pos)++
            
                value:=(uint64(firstByte)<<24 | uint64(secondByte)<<16 | uint64(thirdByte)<<8 | uint64(fourthByte))
                return value,nil

            case 3:
                panic("LZF compressed string is not yet implemented")
      }


      return 0,nil

}


func readLength(data []byte,pos *int)(uint64,error){

     if *pos>=len(data){
           return 0,io.ErrUnexpectedEOF
      }

     /* 
        encodings:

        first 2 bits of the first byte

        if they are:

        00-->the length is the remaining 6 bits
        01-->the length value is the next 14 bits (next byte+ 6 bits of the current byte)
        10-->the length value is 4 bytes
        11-->special encoded value (will be implemented later)
     
     */

     
      encoding:=data[*pos]

      (*pos)++

      encodingType:=encoding>>6
      
      

      switch encodingType{
      case 0:
        length:=encoding & 0x3F
        return uint64(length),nil

      case 1:
          if *pos>=len(data){
              return 0,io.ErrUnexpectedEOF
          }
         //get the remaining 6 bits
         low6Bits:=encoding & 0x3F
         //get the second byte
         secondByte:=data[*pos]
         (*pos)++
         /* 
              The length should be 
              [the low 6 bits][8 bits from the second byte]

              therefore:
               we shit the low 6 bits left by 8
               then perform and or operation with the 8 bits from the second byte
         */

         length:= (uint32(low6Bits)<<8 | uint32(secondByte))
         return uint64(length),nil
      case 2:
           if *pos+4>len(data){
              return 0,io.ErrUnexpectedEOF
           }

          firstByte:=data[*pos]
          (*pos)++
          secondByte:=data[*pos]
          (*pos)++
          thirdByte:=data[*pos]
          (*pos)++

          fourthByte:=data[*pos]
          (*pos)++

          length:=(uint32(firstByte)<<24 |
                   uint32(secondByte)<<16 |
                   uint32(thirdByte)<<8|
                   uint32(fourthByte))
          return uint64(length),nil

      case 3:
           return specialEncoding(data,encoding,pos)
      }



 return 0,nil

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
               0xFB-->database size(RESIZEDB)
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
                       dbHashTableSize,err:=readLength(data,&pos)
                       if err!=nil{
                            //handle error
                       }

                       expiryHashTableSize,err:=readLength(data,&pos)
                       if err!=nil{
                          //handle error
                       }
                       fmt.Printf("hash table size=%d, expiry hash table size=%d\r\n",dbHashTableSize,expiryHashTableSize)


                 case 0xFE:
                    dbNumber,err:=selectDatabase(data,&pos)
                    if err!=nil{
                        //handle error
                    }
                      fmt.Printf("database number=%d\r\n",dbNumber)
                 case 0xFF:
                    break loop
                  

           }
      }


}

func selectDatabase(data []byte, pos *int) (uint64,error){
	   databaseNumber,err:=readLength(data,pos)

       if err!=nil{
           return 0,err
       }

       return databaseNumber,nil
}

func parseAuxilarySection(data []byte, pos *int) ([]byte, []byte, error) {
      
	 

      keyLength,err:=readLength(data,pos)

      if err!=nil{
          return EOF()
      }

      if keyLength> uint64(len(data)-*pos){
          return EOF()
      }

      key:=data[*pos:*pos+int(keyLength)]

      *pos+=int(keyLength)

      valueLength,err:=readLength(data,pos)
      
   
      if err!=nil{
          return EOF()
      }

      if valueLength>uint64(len(data)-*pos){
          return EOF()
      }

      value:=data[*pos:*pos+int(valueLength)]

      *pos+=int(valueLength)

      return key,value,nil

}


func EOF() ([]byte,[]byte,error){
      return nil,nil,io.ErrUnexpectedEOF
}


