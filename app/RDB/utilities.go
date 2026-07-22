package rdb

import (
	"errors"
	"fmt"
	"io"
)

var RDBHeader = []byte{}

var EmptyRDB = []byte{
	0x52, 0x45, 0x44, 0x49, 0x53, 0x30, 0x30, 0x31,
	0x31, 0xfa, 0x09, 0x72, 0x65, 0x64, 0x69, 0x73,
	0x2d, 0x76, 0x65, 0x72, 0x05, 0x37, 0x2e, 0x32,
	0x2e, 0x30, 0xfa, 0x0a, 0x72, 0x65, 0x64, 0x69,
	0x73, 0x2d, 0x62, 0x69, 0x74, 0x73, 0xc0, 0x40,
	0xfa, 0x05, 0x63, 0x74, 0x69, 0x6d, 0x65, 0xc2,
	0x6d, 0x08, 0xbc, 0x65, 0xfa, 0x08, 0x75, 0x73,
	0x65, 0x64, 0x2d, 0x6d, 0x65, 0x6d, 0xc2, 0xb0,
	0xc4, 0x10, 0x00, 0xfa, 0x08, 0x61, 0x6f, 0x66,
	0x2d, 0x62, 0x61, 0x73, 0x65, 0xc0, 0x00,

	// SELECTDB
	0xfe,
	0x00,

	// RESIZEDB
	0xfb,
	0x03, // 3 keys
	0x00, // zero expiring keys

	// String object
	0x00,

	// key: "name"
	0x04,
	0x6e, 0x61, 0x6d, 0x65,

	// value: "goat"
	0x04,
	0x67, 0x6f, 0x61, 0x74,

	// --- KEY 2: List object (0x01) ---
	0x01,                                     // Value type: LIST
	0x06, 0x66, 0x72, 0x75, 0x69, 0x74, 0x73, // key: "fruits" (len 6)
	0x02,                               // list size: 2 items
	0x05, 0x61, 0x70, 0x70, 0x6c, 0x65, // item 1: "apple" (len 5)
	0x06, 0x62, 0x61, 0x6e, 0x61, 0x6e, 0x61, // item 2: "banana" (len 6)

	// --- KEY 3: List object (0x01) ---
	0x01,                                     // Value type: LIST
	0x06, 0x63, 0x6f, 0x6c, 0x6f, 0x72, 0x73, // key: "colors" (len 6)
	0x03,                   // list size: 3 items
	0x03, 0x72, 0x65, 0x64, // item 1: "red" (len 3)
	0x05, 0x67, 0x72, 0x65, 0x65, 0x6e, // item 2: "green" (len 5)
	0x04, 0x62, 0x6c, 0x75, 0x65, // item 3: "blue" (len 4)

	// EOF
	0xff,

	// Checksum (CRC64 calculated for this exact payload)
	0x77, 0xd0, 0x7c, 0xd6, 0x6f, 0x24, 0x19, 0xd1,
}

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
	Key       []byte
	Value     any
	Type      TYPE
	HasExpiry bool
	ExpiresAt uint64
}

func (err *readErr) Error() string {
	return fmt.Sprintf("%s:%s\r\n", err.Name, err.Err.Error())
}

func readString(data []byte, pos *int, isKey bool) ([]byte, error) {

	length, err := decodeLength(data, pos)

	if err != nil {
		return nil, err
	}

	var value []byte

	if length.Special {
		if isKey {
			return nil, errors.New("Special encoding for keys is not allowed")
		}
		value = fmt.Appendf(nil, "%d", length.Value)
	} else {

		if length.Value > uint64(len(data)-*pos) {
			return nil, io.ErrUnexpectedEOF
		}

		value = data[*pos : *pos+int(length.Value)]

		*pos += int(length.Value)
	}

	return value, err

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
