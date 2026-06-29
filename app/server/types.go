package server

/*
   
*/
type Data struct{

	    Type string
		 Value any
}

type ResponseType int

const (
	  ERROR ResponseType=iota
	  SIMPLE_STRING
	  BULK_STRING
	  NIL
	  INTEGER
)

type Response struct{
	   Body []byte
		Type ResponseType
}
