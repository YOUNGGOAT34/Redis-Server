package rdb

import (
	"CacheDB/app/storage"
	"errors"
	"io"
	"os"
	"time"
)




func SaveRDB(path string) error {
    tempPath := path + ".tmp"

    file, err := os.Create(tempPath)
    if err != nil {
        return err
    }

    if err := writeRDBHeader(file); err != nil {
        file.Close()
        return err
    }

    if err := writeAuxFileds(file); err != nil {
        file.Close()
        return err
    }

    if err := writeselectdatabase(file, 0); err != nil {
        file.Close()
        return err
    }

    if err := WriteReSizeDB(file); err != nil {
        file.Close()
        return err
    }

    if err := writeKeyValueString(file); err != nil {
        file.Close()
        return err
    }

    if _, err := file.Write([]byte{0xFF}); err != nil {
        file.Close()
        return err
    }

    if err := file.Close(); err != nil {
        return err
    }

    return os.Rename(tempPath, path)
}


func writeRDBHeader(w io.Writer) error{
	    _,err:=w.Write([]byte("REDIS0011"))
		 if err!=nil{
			   return err
		 }

		 return nil
}


func writeAuxFileds(w io.Writer) error{
	if err := writeAuxString(w,"redis-ver","7.2.0",); err != nil {
        return err
    }

    if err := writeAuxInteger(w,"redis-bits",64,); err != nil {
        return err
    }

	 if err := writeAuxInteger(w, "ctime", uint32(time.Now().Unix())); err != nil {
        return err
    }
	 

	  if err := writeAuxInteger(w, "used-mem", 0); err != nil {
        return err
    }


	 if err := writeAuxInteger(w, "aof-base", 0); err != nil {
        return err
    }

	return nil
}




func encodeSpecialInteger(w io.Writer,value uint32) error{

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


func writeselectdatabase(w io.Writer,databaseNumber int) error{
	   _,err:=w.Write([]byte{0xFE})
		if err!=nil{
			  return err
		}

	   return encodeLength(w,databaseNumber)
}


func WriteReSizeDB(w io.Writer) error{
	     _,err:=w.Write([]byte{0xFB})

		  if err!=nil{
			 return err
		  }


	

		  err=encodeLength(w,len(storage.Database))
	
		  
		  if err!=nil{
			 return err
		  }
		 err=encodeLength(w,len(storage.Expiry))
		 if err!=nil{
			 return err
		 }

		 return nil

}

func writeAuxInteger(w io.Writer,key string,value uint32) error{

	   /*
					AUX opcode
					encoded length of key
					key
					special encoding
					integer bytes
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


		 err=encodeSpecialInteger(w,value)

		 if err!=nil{
			  return err
		 }
       
		 return nil
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


func writeKeyValueString(w io.Writer) error{
	   
	     for key,value:=range storage.Database{

			    err:=writeObjectType(w,value.Type)
				 if err!=nil{
					 return err
					}
					
				err=writeKey(w,key)
				 if err!=nil{
					 return err
				 }

				 err=writeValue(w,value)

				 if err!=nil{
					  return err
				 }
              
		  }

		  return nil
}

func writeKey(w io.Writer,key string) error{
	   err:=encodeLength(w,len(key))

		if err!=nil{
			 return err
		}

		_,err=w.Write([]byte(key))

    return err
}

func writeValue(w io.Writer,data storage.Data) error{
	    switch data.Type{
		 case storage.STRING:
			    value,ok:=data.Value.([]byte)

				 if !ok {
					  return  errors.New("Wrong data stored in a string type")
				 }

				 err:=encodeLength(w,len(value))

				 if err!=nil{
					 return err
				 }

				 _,err=w.Write(value)

				 return err

				default:
					return errors.New("unknown data type")

		 }
}


func writeObjectType(w io.Writer,objectType storage.TYPE) error{
	   switch objectType{

				case storage.STRING:
					_,err:=w.Write([]byte{0x00})

					return err
				default:
					return errors.New("Unkown object type")
		}
}


