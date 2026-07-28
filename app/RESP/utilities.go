package RESP

import (
	"bytes"
	"fmt"
)

func CompareBytes(a, b []byte) bool {
	if bytes.EqualFold(a, b) {
		return true
	}

	return false
}

func WrongType() Response {
	return Response{
		Body: []byte("WRONGTYPE Operation against a key holding the wrong kind of value"),
		Type: ERROR,
	}
}



func WrongNumberOfArguments(command string) Response {
	errMessage := fmt.Sprintf("Wrong number of arguments for '%s' command", command)
	return Response{
		Body: []byte(errMessage),
		Type: ERROR,
	}
}


