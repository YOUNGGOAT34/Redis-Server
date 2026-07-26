package RESP




type ResponseType int


const (
	ERROR ResponseType = iota
	SIMPLE_STRING
	BULK_STRING
	NIL
	INTEGER
	ARRAY
	RDBFILE
)

type Response struct {
	Body []byte
	Type ResponseType
}
