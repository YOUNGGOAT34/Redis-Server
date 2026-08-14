package streams

import "encoding/binary"

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


