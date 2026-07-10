package RESP

type ResponseType int

type SERVER struct {
	Role             string
	MasterHost       string
	MasterPort       int
	PORT             int
	MASTERREPLID     string
	MASTERREPLOFFSET int
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
