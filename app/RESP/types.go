package RESP




type ResponseType int


const (
	ERROR ResponseType = iota
	SIMPLE_STRING
	BULK_STRING
	NIL
	INTEGER
	ARRAY
	LIST
	RDBFILE
	STREAM
	TRANSACTION
)

type Response struct {
	Body []byte
	Array []Response
	Type ResponseType
}
