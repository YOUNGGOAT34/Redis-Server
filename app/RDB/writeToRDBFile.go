package rdb

import (
	"errors"
	"io"
)


func writeRDBHeader(w io.Writer) error{
	    _,err:=w.Write([]byte("REDIS0011"))
		 if err!=nil{
			   return err
		 }

		 return nil
}

func writeAuxFileds(w io.Writer) error{
	    
}

func writeAuxFiled(w io.Writer,key string,value string) error{
	     /*
		      AUX opcode
				length of key
				key
				length of value
				value
		  */
	    if  _,err:=w.Write([]byte{0xFA});err !=nil{
			   return err
		 }

		 if _,err:=w.Write([]byte{byte(len(key))});err!=nil{
			  return err
		 }


		 if _,err:=w.Write([]byte(key));err!=nil{
			  return err
		 }


		if _,err:=w.Write([]byte{byte(len(value))});err!=nil{
			  return err
		 }


		 if _,err:=w.Write([]byte(value));err!=nil{
			  return err
		 }
      
		 

		 return nil
}


func encodeLength(w io.Writer,length int) error{

	   if length<0{
			  return errors.New("length cannot be negative")
		}
	   if length<64{
			   buffer:=[1]byte{byte(length)}
			  _,err:=w.Write(buffer[:])
			     
				  return err
			  

		}else if length<16384{
			  
				buffer:=[2]byte{
					  byte((length>>8) | 0x40),
					  byte(length & 0xFF),
				}

				_,err:=w.Write(buffer[:])

			
					  return err
				
		}else{
			  
			   buffer:=[5]byte{
					     0x80,
						  byte(length>>24),
						  byte(length>>16),
						  byte(length>>8),
						  byte(length & 0xFF),

				}
         
			  _,err:=w.Write(buffer[:])
			 
				 return err
			  
		}

}