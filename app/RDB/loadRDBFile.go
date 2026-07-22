package rdb

import (
	"CacheDB/app/RESP"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

func decodeSpecialEncoding(data []byte, encoding byte, pos *int) (uint64, error) {

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

		fmt.Fprintln(os.Stderr, WrappedError.Error())
		return nil, io.ErrUnexpectedEOF
	}

	opcode := data[*pos]
	(*pos)++

	switch opcode {
	case 0xFD:

		expiryInSeconds, err := readNBytes(data, pos, 4)
		if err != nil {

			WrappedError := &readErr{
				Name: "0xFD reading Expiry In seconds",
				Err:  err,
			}

			fmt.Fprintln(os.Stderr, WrappedError.Error())
			return nil, err
		}

		seconds := binary.LittleEndian.Uint32(expiryInSeconds)

		DATA.HasExpiry = true
		DATA.ExpiresAt = uint64(seconds) * 1000

		opcode, err = readByte(data, pos)
		if err != nil {
			WrappedError := &readErr{
				Name: "0xFD reading entry",
				Err:  err,
			}

			fmt.Fprintln(os.Stderr, WrappedError.Error())

			return nil, err
		}

	case 0xFC:

		expiryInMilliseconds, err := readNBytes(data, pos, 8)

		if err != nil {

			WrappedError := &readErr{
				Name: "0xFC expiry in milliseconds",
				Err:  err,
			}

			fmt.Fprintln(os.Stderr, WrappedError.Error())
			return nil, err
		}

		milliseconds := binary.LittleEndian.Uint64(expiryInMilliseconds)

		DATA.HasExpiry = true
		DATA.ExpiresAt = milliseconds

		opcode, err = readByte(data, pos)

		if err != nil {

			WrappedError := &readErr{
				Name: "Reading entry's Opcode(key type)",
				Err:  err,
			}

			fmt.Fprintln(os.Stderr, WrappedError.Error())

			return nil, err
		}

	}

	//object type

	switch opcode {

	case 0x00:

		key, value, err := readStringKeyValuePair(data, pos)

		if err != nil {
			WrappedError := &readErr{
				Name: "Key of string type",
				Err:  err,
			}

			fmt.Fprintln(os.Stderr, WrappedError.Error())
			return nil, err
		}

		DATA.Key = key
		DATA.Value = value
		DATA.Type = STRING
		return DATA, nil
	case 0x01:

		key, list, err := readListKeyValuePair(data, pos)

		if err != nil {
			return nil, err
		}

		DATA.Key = key
		DATA.Value = list
		DATA.Type = LIST
		return DATA, nil

	case 0x02:
		return nil, errors.New("sets encoding is not implemented")
	case 0x0F:
		return nil, errors.New("streams encoding is not implemented")
	default:
		return nil, fmt.Errorf("unknown object type: 0x%02x", opcode)

	}

	// return &Data{}, nil

}

func decodeLength(data []byte, pos *int) (LengthResult, error) {

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
		value, err := decodeSpecialEncoding(data, encoding, pos)
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

			fmt.Fprintln(os.Stderr, WrappedError.Error())

			return nil, err
		}

	}

	if err != nil {
		WrappedError := &readErr{
			Name: "getting file information",
			Err:  err,
		}

		fmt.Fprintln(os.Stderr, WrappedError.Error())

		return nil, err
	}

	data, err := os.ReadFile(rdbConfig.Dir + "/" + rdbConfig.DbFileName)

	if err != nil {

		fmt.Printf("directory: %s ,file name :%s\r\n", rdbConfig.DbFileName, rdbConfig.Dir)

		WrappedError := &readErr{
			Name: "reading rdb file",
			Err:  err,
		}

		fmt.Fprintln(os.Stderr, WrappedError.Error())

		return nil, err
	}

	header, err := readHeader(data, &pos)

	if err != nil {
		WrappedError := &readErr{
			Name: "empty header",
			Err:  io.ErrUnexpectedEOF,
		}

		fmt.Fprintln(os.Stderr, WrappedError.Error())
		return []*Data{}, err
	}

	if !RESP.CompareBytes(header[:5], []byte("REDIS")) {
		WrappedError := &readErr{
			Name: "The file is not an rdb file",
			Err:  io.ErrUnexpectedEOF,
		}

		fmt.Fprintln(os.Stderr, WrappedError.Error())

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
			_, _, err := readStringKeyValuePair(data, &pos)
			if err != nil {
				WrappedError := &readErr{
					Name: "Auxilary values",
					Err:  err,
				}

				fmt.Fprintln(os.Stderr, WrappedError.Error())
				return []*Data{}, err
			}

			// fmt.Printf("key=%s,value=%s\r\n", auxiliaryKey, auxiliaryValue)

		case 0xFB:
			dbHashTableSize, err := decodeLength(data, &pos)
			if err != nil {
				WrappedError := &readErr{
					Name: "Database HashTable size",
					Err:  io.ErrUnexpectedEOF,
				}

				fmt.Fprintln(os.Stderr, WrappedError.Error())

				return nil, err
			}

			_, err = decodeLength(data, &pos)
			if err != nil {

				WrappedError := &readErr{
					Name: "Database expiry HashTable size",
					Err:  io.ErrUnexpectedEOF,
				}

				fmt.Fprintln(os.Stderr, WrappedError.Error())
			}
			// fmt.Printf("hash table size=%d, expiry hash table size=%d\r\n", dbHashTableSize.Value, expiryHashTableSize.Value)

			for i := uint64(0); i < dbHashTableSize.Value; i++ {
				dataEntry, err := readEntry(data, &pos)

				if err != nil {
					WrappedError := &readErr{
						Name: "Key value pair",
						Err:  io.ErrUnexpectedEOF,
					}

					fmt.Fprintln(os.Stderr, WrappedError.Error())

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

				fmt.Fprintln(os.Stderr, WrappedError.Error())

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
	length, err := decodeLength(data, pos)

	if err != nil {
		return 0, err
	}

	return length.Value, err
}

// strings
func readStringKeyValuePair(data []byte, pos *int) ([]byte, []byte, error) {

	key, err := readString(data, pos, true)

	if err != nil {
		return EOF()
	}

	value, err := readString(data, pos, false)

	if err != nil {
		return EOF()
	}

	return key, value, nil

}

// lists
func readListKeyValuePair(data []byte, pos *int) ([]byte, [][]byte, error) {
	key, err := readString(data, pos, true)

	if err != nil {
		return nil, nil, err
	}

	listLength, err := decodeLength(data, pos)

	if err != nil {
		return nil, nil, err
	}

	list := make([][]byte, 0, listLength.Value)

	for i := uint64(0); i < listLength.Value; i++ {
		element, err := readString(data, pos, false)

		if err != nil {
			return nil, nil, err
		}

		list = append(list, element)
	}

	return key, list, nil
}
