package rdb

import (
	"CacheDB/app/RESP"
	"errors"
	"fmt"
	"io"
	"os"
)

type LengthResult struct {
	Value   uint64
	Special bool
}

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
	0x01, // one key
	0x00, // zero expiring keys

	// String object
	0x00,

	// key: "name"
	0x04,
	0x6e, 0x61, 0x6d, 0x65,

	// value: "goat"
	0x04,
	0x67, 0x6f, 0x61, 0x74,

	// EOF
	0xff,

	// checksum
	0xf0, 0x6e, 0x3b, 0xfe, 0xc0, 0xff, 0x5a, 0xa2,
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
	Value []byte
	Type  TYPE
}

func (err *readErr) Error() string {
	return fmt.Sprintf("%s:%s\r\n", err.Name, err.Err.Error())
}

//helpers

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

func specialEncoding(data []byte, encoding byte, pos *int) (uint64, error) {

	specialEncodingType := encoding & 0x3F

	/*
	   0-->read the next byte
	   1--->read the next two bytes
	   2--->read the next 4 bytes
	   3--->LZF compressed string
	*/

	switch specialEncodingType {
	case 0:
		if *pos >= len(data) {
			return 0, io.ErrUnexpectedEOF
		}

		value := uint64(data[*pos])
		(*pos)++

		return value, nil

	case 1:
		if *pos+2 > len(data) {
			return 0, io.ErrUnexpectedEOF
		}

		firstByte := data[*pos]
		(*pos)++
		secondByte := data[*pos]
		(*pos)++

		value := (uint64(firstByte)<<8 | uint64(secondByte))

		return value, nil

	case 2:

		if *pos+4 > len(data) {
			return 0, io.ErrUnexpectedEOF
		}

		firstByte := data[*pos]
		(*pos)++
		secondByte := data[*pos]
		(*pos)++
		thirdByte := data[*pos]
		(*pos)++
		fourthByte := data[*pos]
		(*pos)++

		value := (uint64(firstByte)<<24 | uint64(secondByte)<<16 | uint64(thirdByte)<<8 | uint64(fourthByte))
		return value, nil

	case 3:
		panic("LZF compressed string is not yet implemented")
	}

	return 0, nil

}

func readEntry(data []byte, pos *int) (*Data, error) {

	DATA := &Data{}

	if *pos >= len(data) {
		WrappedError := &readErr{
			Name: "No entry",
			Err:  io.ErrUnexpectedEOF,
		}

		fmt.Fprintln(os.Stderr,WrappedError.Error())
		return nil, io.ErrUnexpectedEOF
	}

	opcode := data[*pos]
	(*pos)++

	switch opcode {
	case 0xFD:

		_, err := readNBytes(data, pos, 4)
		if err != nil {
			//handle error
			WrappedError := &readErr{
				Name: "0xFD reading Expiry In seconds",
				Err:  err,
			}

			fmt.Fprintln(os.Stderr, WrappedError.Error())
			return nil, err
		}

		opcode, err = readByte(data, pos)
		if err != nil {
			WrappedError := &readErr{
				Name: "0xFD reading entry",
				Err:  err,
			}

			fmt.Fprintln(os.Stderr,WrappedError.Error())

			return nil, err
		}

	case 0xFC:

		_, err := readNBytes(data, pos, 8)

		if err != nil {

			WrappedError := &readErr{
				Name: "0xFC expiry in milliseconds",
				Err:  err,
			}

			fmt.Fprintln(os.Stderr, WrappedError.Error())
			return nil, err
		}

		opcode, err = readByte(data, pos)

		if err != nil {

			WrappedError:= &readErr{
				Name: "Reading entry's Opcode(key type)",
				Err:  err,
			}

			fmt.Fprintln(os.Stderr,WrappedError.Error())

			return nil, err
		}

	}

	switch opcode {

	case 0x00:
		//read key value length
		key, value, err := readKeyValuePair(data, pos)

		if err != nil {
			WrappedError := &readErr{
				Name: "Key of string type",
				Err:  err,
			}

			fmt.Fprintln(os.Stderr,WrappedError.Error())
			return nil, err
		}

		DATA.Key = key
		DATA.Value = value
		DATA.Type = STRING
		return DATA, nil
	case 0x01:
		fmt.Printf("List\r\n")
	case 0x02:
		fmt.Printf("set\r\n")
	}

	return &Data{}, nil

}

func readLengthOrEncoding(data []byte, pos *int) (LengthResult, error) {

	if *pos >= len(data) {
		return LengthResult{
			Value:   0,
			Special: false,
		}, io.ErrUnexpectedEOF
	}

	/*
	   encodings:

	   first 2 bits of the first byte

	   if they are:

	   00-->the length is the remaining 6 bits
	   01-->the length value is the next 14 bits (next byte+ 6 bits of the current byte)
	   10-->the length value is 4 bytes
	   11-->special encoded value (will be implemented later)

	*/

	encoding := data[*pos]

	(*pos)++

	encodingType := encoding >> 6

	switch encodingType {
	case 0:
		length := encoding & 0x3F
		return LengthResult{
			Value:   uint64(length),
			Special: false,
		}, nil

	case 1:
		if *pos >= len(data) {
			return LengthResult{
				Value:   0,
				Special: false,
			}, io.ErrUnexpectedEOF
		}
		//get the remaining 6 bits
		low6Bits := encoding & 0x3F
		//get the second byte
		secondByte := data[*pos]
		(*pos)++
		/*
		   The length should be
		   [the low 6 bits][8 bits from the second byte]

		   therefore:
		    we shit the low 6 bits left by 8
		    then perform and or operation with the 8 bits from the second byte
		*/

		length := (uint32(low6Bits)<<8 | uint32(secondByte))
		return LengthResult{
			Value:   uint64(length),
			Special: false,
		}, nil
	case 2:
		if *pos+4 > len(data) {
			return LengthResult{
				Value:   0,
				Special: false,
			}, io.ErrUnexpectedEOF
		}

		firstByte := data[*pos]
		(*pos)++
		secondByte := data[*pos]
		(*pos)++
		thirdByte := data[*pos]
		(*pos)++

		fourthByte := data[*pos]
		(*pos)++

		length := (uint32(firstByte)<<24 |
			uint32(secondByte)<<16 |
			uint32(thirdByte)<<8 |
			uint32(fourthByte))
		return LengthResult{
			Value:   uint64(length),
			Special: false,
		}, nil

	case 3:
		value, err := specialEncoding(data, encoding, pos)
		if err != nil {
			return LengthResult{}, err
		}

		return LengthResult{
			Value:   value,
			Special: true,
		}, nil
	}

	return LengthResult{}, nil

}

func ReadRDBFile(rdbConfig *RDB) ([]*Data, error) {

	var database []*Data

	//cursor position
	pos := 0

	path := rdbConfig.Dir + "/" + rdbConfig.DbFileName

	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		err = os.WriteFile(path, EmptyRDB, 0644)

		if err != nil {
			WrappedError := &readErr{
				Name: "Creating new file",
				Err:  err,
			}

			fmt.Fprintln(os.Stderr,WrappedError.Error())

			return nil, err
		}

	}

	if err != nil {
		WrappedError := &readErr{
			Name: "getting file information",
			Err:  err,
		}

		fmt.Fprintln(os.Stderr,WrappedError.Error())

		return nil, err
	}

	data, err := os.ReadFile(rdbConfig.Dir + "/" + rdbConfig.DbFileName)

	if err != nil {
     
		fmt.Printf("directory: %s ,file name :%s\r\n",rdbConfig.DbFileName,rdbConfig.Dir)

		WrappedError := &readErr{
			Name: "reading rdb file",
			Err:  err,
		}

		fmt.Fprintln(os.Stderr,WrappedError.Error())

		return nil, err
	}

	header, err := readHeader(data, &pos)

	if err != nil {
		WrappedError := &readErr{
			Name: "empty header",
			Err:  io.ErrUnexpectedEOF,
		}

		fmt.Fprintln(os.Stderr,WrappedError.Error())
		return []*Data{}, err
	}

	if !RESP.CompareBytes(header[:5], []byte("REDIS")) {
		WrappedError := &readErr{
			Name: "The file is not an rdb file",
			Err:  io.ErrUnexpectedEOF,
		}

		fmt.Fprintln(os.Stderr,WrappedError.Error())

		return []*Data{}, errors.New("The given file is not an rdb file")
	}

loop:
	for {

		opcode, err := readByte(data, &pos)

		if err != nil {
			return []*Data{}, err
		}

		/*
		   0xFA-->auxilary field
		   0xFB-->database size(RESIZEDB)
		   0XFE--->database selector
		   0xFF--->end of file

		*/

		switch opcode {
		case 0xFA:
			_ , _ , err := readKeyValuePair(data, &pos)
			if err != nil {
				WrappedError := &readErr{
					Name: "Auxilary values",
					Err:  err,
				}

				fmt.Fprintln(os.Stderr,WrappedError.Error())
				return []*Data{}, err
			}

			// fmt.Printf("key=%s,value=%s\r\n", auxiliaryKey, auxiliaryValue)

		case 0xFB:
			dbHashTableSize, err := readLengthOrEncoding(data, &pos)
			if err != nil {
				WrappedError := &readErr{
					Name: "Database HashTable size",
					Err:  io.ErrUnexpectedEOF,
				}

				fmt.Fprintln(os.Stderr,WrappedError.Error())
			}

			_, err = readLengthOrEncoding(data, &pos)
			if err != nil {

				WrappedError := &readErr{
					Name: "Database expiry HashTable size",
					Err:  io.ErrUnexpectedEOF,
				}

				fmt.Fprintln(os.Stderr,WrappedError.Error())
			}
			// fmt.Printf("hash table size=%d, expiry hash table size=%d\r\n", dbHashTableSize.Value, expiryHashTableSize.Value)

			for i := uint64(0); i < dbHashTableSize.Value; i++ {
				dataEntry, err := readEntry(data, &pos)

				if err != nil {
					WrappedError := &readErr{
						Name: "Key value pair",
						Err:  io.ErrUnexpectedEOF,
					}

					fmt.Fprintln(os.Stderr,WrappedError.Error())

					return nil, err
				}

				database = append(database, dataEntry)

			}

		case 0xFE:
			_, err := selectDatabase(data, &pos)
			if err != nil {
				WrappedError := &readErr{
					Name: "Database number",
					Err:  err,
				}

				fmt.Fprintln(os.Stderr,WrappedError.Error())

				return nil, err
			}
			// fmt.Printf("database number=%d\r\n", dbNumber)
		case 0xFF:
			break loop

		}
	}

	return database, nil

}

func selectDatabase(data []byte, pos *int) (uint64, error) {
	length, err := readLengthOrEncoding(data, pos)

	if err != nil {
		return 0, err
	}
    
	return length.Value, err
}

func readKeyValuePair(data []byte, pos *int) ([]byte, []byte, error) {

	keyLength, err := readLengthOrEncoding(data, pos)

	if err != nil {
		return EOF()
	}

	var key []byte

	if keyLength.Special {
		key = fmt.Appendf(nil, "%d", keyLength.Value)
	} else {

		if keyLength.Value > uint64(len(data)-*pos) {
			return EOF()
		}

		key = data[*pos : *pos+int(keyLength.Value)]

		*pos += int(keyLength.Value)
	}

	valueLength, err := readLengthOrEncoding(data, pos)

	if err != nil {
		return EOF()
	}

	var value []byte

	if valueLength.Special {
		value = fmt.Appendf(nil, "%d", valueLength.Value)
	} else {

		if valueLength.Value > uint64(len(data)-*pos) {
			return EOF()
		}

		value = data[*pos : *pos+int(valueLength.Value)]

		*pos += int(valueLength.Value)
	}
    
	return key, value, nil

}

func EOF() ([]byte, []byte, error) {
	return nil, nil, io.ErrUnexpectedEOF
}
