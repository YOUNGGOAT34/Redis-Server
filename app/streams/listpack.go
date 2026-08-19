package streams

import (
	"encoding/binary"
	"errors"
)

const(
	  EOF byte=0xFF
	  HeaderSize int =6
)

type ListPack struct{
	  data []byte
}

func NewListPack() *ListPack{
	 lp:=&ListPack{
		  data:make([]byte,HeaderSize+1),
	 }

	 //total size
	 binary.LittleEndian.PutUint32(lp.data[0:4],uint32(len(lp.data)))
	 //number of elements-->at this point there are 0 elemenents
	 binary.LittleEndian.PutUint16(lp.data[4:6],0)
	 //end of file
	 lp.data[6]=EOF
	 return lp
}

/*
   Read the header:
	    1.first 4 bytes-->total bytes
		 2.next 2 bytes--->number of elements
*/

func(lp *ListPack) TotalBytes() uint32{
	  return binary.LittleEndian.Uint32(lp.data[0:4])
}

func (lp *ListPack) Length()uint16{
	  return binary.LittleEndian.Uint16(lp.data[4:6])
}


//encodings

func encode12BitString(length int) ([]byte,error){
		if length<0 || length>4095{
			  return nil,errors.New("value cannot fit into 12 bits")
		}
		/*
			The first 4 bits identify this as a 12-bit string:

				1110xxxx xxxxxxxx

			The remaining 12 bits store the string length.
		*/
		encoded:=uint16(0xE000) | uint16(length)

		return []byte{byte(encoded>>8),byte(encoded&0xFF)},nil
}


func encode32BitString(length int) ([]byte,error){
	  if length<0 || length>4294967295{
			 return nil,errors.New("value cannot fit into 32 bits")
	  }

	  /*
			0xF0 identifies this entry as a 32-bit string
			The next 4 bytes store the string length in big-endian
			order, from the most significant byte to the least
			significant byte.
	  */

	  prefix:=byte(0xF0)
	  return []byte{prefix,byte(length>>24),byte(length>>16),byte(length>>8),byte(length & 0xFF)},nil
}
func encode7BitInteger(value int) (byte,error){
	  if value<0 || value>127{
		  return 0,errors.New("Value does not fit in 7 bits")
	  }
	return byte(value),nil
}

func encode13BitInteger(value int)([]byte,error){
	 if value < -4096 || value > 4095 {
		  return nil,errors.New("Value does not fit in 13 bits")
	 }

	 /*
	      A 13-bit integer only has 13 bits available for its value.
			This is important for negative numbers because Go's
			int uses two's-complement representation-->meaning a negative
			value has 1s extending into the higher bits
			Masking with 0x1FFF  discards those higher bits and
			keeps only the 13 bits that belong to the Listpack encoding
	 */

	 value&=0x1FFF

	 encoded:=uint16(0xC000) | uint16(value)

	 return []byte{byte(encoded>>8),byte(encoded & 0xFF)},nil
}

func encode16BitInteger(value int) ([]byte,error){
	   if value < -32768 || value>32767{
           return nil,errors.New("value cannot fit into 16 bits")
		}

	/*
	   0xF1 identifies this entry as a 16-bit signed integer.
	   Keep only the lower 16 bits so that negative values are
	   represented using their 16-bit two's-complement form.
	*/

		prefix:=byte(0xF1)

		value&=0xFFFF
	
		return []byte{prefix,byte(value>>8),byte(value & 0xFF)},nil
}

func encode24BitInteger(value int) ([]byte,error){
	   if  value < -8388608 || value> 8388607{
			 return nil,errors.New("value cannot fit into 24 bits")
		}

		/*
			0xF2 identifies this entry as a 24-bit signed integer.
			A 24-bit integer uses the remaining 3 bytes to store the value.
			Masking with 0xFFFFFF keeps only those 24 bits, which gives
			negative values their 24-bit two's-complement representation.
			The 24 bits are then split into three bytes in big-endian order.
		*/

		prefix:=byte(0xF2)
		
		value &=0xFFFFFF
		return []byte{prefix,byte(value>>16),byte(value>>8),byte(value & 0xFF)},nil
}

func encode32BitInteger(value int) ([]byte,error){
	  if value < -2147483648 || value>2147483647{
		   return nil,errors.New("value cannot fit into 32 bits")
	  }
     
	  /*
			0xF3 identifies this entry as a 32-bit signed integer.

			A 32-bit integer uses the remaining 4 bytes to store the value.
			Masking with 0xFFFFFFFF keeps only those 32 bits, giving negative
			values their 32-bit two's-complement representation.

			The 32 bits are then split into four bytes in big-endian order.
		*/
	  prefix:=byte(0xF3)
	  value&=0xFFFFFFFF

	  return []byte{prefix,byte(value>>24),byte(value>>16),byte(value>>8),byte(value&0xFF)},nil
}

func encode64BitInteger(value int64) ([]byte,error){
	 /*
			0xF4 identifies this entry as a 64-bit signed integer.
			The integer occupies the remaining 8 bytes and is stored in
			big-endian order, from the most significant byte to the least
			significant byte.
			Since value is already an int64, it already uses exactly 64 bits.
			Negative values are represented using two's-complement, so no
			additional masking is required.
		*/
	  prefix:=byte(0xF4)
	  return []byte{prefix,byte(value>>56),byte(value>>48),byte(value>>40),byte(value>>32),byte(value>>24),byte(value>>16),byte(value>>8),byte(value & 0xFF)},nil
}

/*
   Backlen encoding:
	   length <= 127          → 1 byte
		length <= 16,383       → 2 bytes
		length <= 2,097,151    → 3 bytes
		length <= 268,435,455  → 4 bytes
		otherwise              → 5 bytes
*/

func backlenSize(length int) int{
	 
	  if length<=127{
		  return 1
	  }else if length<=16383{
		 return 2
	  }else if length<=2097151{
		 return 3
	  }else if length<=268435455{
		 return 4
	  }

	  return 5
}

func encodeBacklen(length int) ([]byte,error){
	   if length<0{
			 return nil,errors.New("Invalid length")
		}

		size:=backlenSize(length)
      
		switch size{
		case 0x01:
			 return []byte{byte(length&0x7F)},nil
		case 0x02:
			  return []byte{byte((length>>7)&0x7F),byte((length&0x7F) | 0x80)},nil
		case 0x03:
			  return []byte{byte(((length>>14)&0x7F)),byte(((length>>7)&0x7F)| 0x80),byte((length&0x7F) | 0x80)},nil
		case 0x04:
			  return []byte{byte(((length>>21)&0x7F)),byte(((length>>14)&0x7F) | 0x80),byte(((length>>7)&0x7F) | 0x80),byte((length&0x7F) | 0x80)},nil
		default:
			  return []byte{byte((length>>28)&0x7F),byte(((length>>21)&0x7F) | 0x80),byte(((length>>14)&0x7F) | 0x80),byte(((length>>7)&0x7F) | 0x80),byte((length&0x7F) | 0x80)},nil
		}
}

func entryTotalSize(contentSize int) int{
	   backlenBytes:=1

		for{
			  totalSize:=contentSize+backlenBytes
			  requiredBytes:=backlenSize(contentSize)

			  if backlenBytes==requiredBytes{
				 return totalSize
			  }

			  backlenBytes=requiredBytes
		}
}

func encodeStringEntry(value string) ([]byte,error){
	  data:=[]byte(value)
	  
	  length:=len(data)

	  var encoding []byte
	  var err error
	
	  if length<=4095{
		    encoding,err=encode12BitString(length)
	  }else{
		    encoding,err=encode32BitString(length)
	  }

	  if err!=nil{
		  return nil,err
	  }
    
	  content:=append(encoding,data...)

	  contentSize:=len(content)
	  totalSize:=entryTotalSize(contentSize)

	  backlenEncoding,err:=encodeBacklen(totalSize)

	  if err!=nil{
		 return nil,err
	  }

	  content = append(content, backlenEncoding...)

	  return content,nil

}