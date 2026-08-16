package streams

import (
	"encoding/binary"
	"errors"

	"github.com/segmentio/encoding/thrift"
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

func encode7BitInteger(value int) (byte,error){
	  if value<0 || value>127{
		  return 0,errors.New("Value does not fit in 7 bits")
	  }
	return byte(value),nil
}

func encode13BitInteger(value int)([]byte,error){
	 if value < -4096 || value > 4096 {
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

	 highBits:=byte(encoded>>8)
	 lowBits:=byte(encoded & 0xFF)

	 return []byte{highBits,lowBits},nil
}

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

		highBits:=byte(encoded>>8)
		lowBits:=byte(encoded&0xFF)

		return []byte{highBits,lowBits},nil
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

	  firstByte:=byte(length>>24)
	  secondByte:=byte(length>>16)
	  thirdByte:=byte(length>>8)
	  fouthByte:=byte(length & 0xFF)

	  return []byte{prefix,firstByte,secondByte,thirdByte,fouthByte},nil
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

		value=value & 0xFFFF
		
		firstByte:=byte(value>>8)
		secondByte:=byte(value & 0xFF)

		return []byte{prefix,firstByte,secondByte},nil
}