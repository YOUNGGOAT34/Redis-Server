package rdb

import (
	"errors"
	"fmt"
	"io"
)



type LengthResult struct {
	Value   uint64
	Special bool
}


type readErr struct {
	Name string
	Err  error
}

type TYPE int

const (
	STRING TYPE = iota
	LIST
	STREAM
)

type Data struct {
	Key   []byte
	Value any
	Type  TYPE
	HasExpiry bool
	ExpiresAt uint64
}

func (err *readErr) Error() string {
	return fmt.Sprintf("%s:%s\r\n", err.Name, err.Err.Error())
}





func readString(data []byte ,pos *int,isKey bool) ([]byte,error){

	length, err := readLengthOrEncoding(data, pos)
	        
	if err != nil {
		return nil,err
	}

	var value []byte

	if length.Special {
		if isKey{
			 return nil,errors.New("Special encoding for keys is not allowed")
		}
		value = fmt.Appendf(nil, "%d", length.Value)
	} else {

		if length.Value > uint64(len(data)-*pos) {
			return nil,io.ErrUnexpectedEOF
		}

		value = data[*pos : *pos+int(length.Value)]

		*pos += int(length.Value)
	}

	return value,err
	  
}

func readByte(data []byte, pos *int) (byte, error) {
	if *pos >= len(data) {
		return 0, io.ErrUnexpectedEOF
	}

	value := data[*pos]

	(*pos)++

	return value, nil
}

func readNBytes(data []byte, pos *int, n int) ([]byte, error) {
	if *pos+n > len(data) {
		return nil, io.ErrUnexpectedEOF
	}

	result := data[*pos : *pos+n]
	*pos += n
	return result, nil
}

func readHeader(data []byte, pos *int) ([]byte, error) {

	if *pos+9 > len(data) {
		return nil, io.ErrUnexpectedEOF
	}

	header := data[*pos : *pos+9]

	*pos += 9

	return header, nil
}


func EOF() ([]byte, []byte, error) {
	return nil, nil, io.ErrUnexpectedEOF
}
