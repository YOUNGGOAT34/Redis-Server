package streams

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

	 lp.data[6]=EOF
	 return lp
}


