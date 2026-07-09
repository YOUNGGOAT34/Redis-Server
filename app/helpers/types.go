package helpers

type ResponseType int


type SERVER struct{
	   Role string
		MasterHost string
		MasterPort int
		PORT int
}


const (
	ERROR ResponseType = iota
	SIMPLE_STRING
	BULK_STRING
	NIL
	INTEGER
	ARRAY
)

type Response struct {
	Body []byte
	Type ResponseType
}


