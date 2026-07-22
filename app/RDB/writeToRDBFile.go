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




func specialEncoding(w io.Writer,key string,value int) error{
	  if value<0{
		  return errors.New("negative integers are not supported")
	  }
	   if value <256{
			     buffer:=[2]byte{0xc0,byte(value&0xFF)}
				  _,err:=w.Write(buffer[:])

				  return err
		}else if value<65536{
			     buffer:=[3]byte{0xc1,byte(value>>8),byte(value&0xFF)}
              _,err:=w.Write(buffer[:])

				  return err

		}else{
			  buffer:=[5]byte{
				         0xc2,
							byte(value>>24),
							byte(value>>16),
							byte(value>>8),
							byte(value&0xFF),
			  }

			  _,err:=w.Write(buffer[:])

				  return err
		}


}

func writeAuxString(w io.Writer,key string,value string) error{
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

		err:=encodeLength(w,len(key))

		if err!=nil{
			  return err
		}


		 if _,err:=w.Write([]byte(key));err!=nil{
			  return err
		 }


		err=encodeLength(w,len(value))

		if err!=nil{
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