package RESP

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

/*
findCRLF returns the index of the first '\r' in the first CRLF sequence.
If no CRLF exists, the request is considered malformed.
*/
func findCRLF(request []byte) int {
	for i := 0; i < len(request); i++ {
		if request[i] == '\r' && i+1 < len(request) && request[i+1] == '\n' {
			return i
		}
	}
	return -1
}

/*
	Split the RESP array header (e.g. "*2") from the remaining payload.
	The returned body intentionally begins with "\r\n" so every bulk string
	can be parsed using the same logic.
*/

func getHeaderAndBody(request []byte) (header, body []byte) {
	headerEndsAt := findCRLF(request)

	if headerEndsAt == -1 {
		return nil, nil
	}

	

	return request[:headerEndsAt], request[headerEndsAt:]
}

/*
	Parse a RESP request into its command arguments.
	Each bulk string is extracted and stored in args for dispatch.
*/

func ParseRequest(request []byte) ([][]byte,int,error) {
	bytesConsumed:=0

	if len(request) < 1 {
		return  nil,bytesConsumed,errors.New("Empty request")
	}

	

	header, body := getHeaderAndBody(request)
  
	var args [][]byte

	if header == nil {
		return nil,bytesConsumed,errors.New("No header")
	}

	bytesConsumed+=len(header)+2

	/*Read the array length after '*'.
	  Supports multi-digit array sizes such as *12.

	*/

	index := 1

	for index < len(header) {
		if header[index] >= '0' && header[index] <= '9' {

			index++
		} else {
			break
		}
	}

	size, err := strconv.Atoi(string(header[1:index]))

	if err != nil {
		return nil,bytesConsumed,err
	}

	// Extract each RESP bulk string from the request body.

	for i := 0; i < size; i++ {

		if len(body) < 5 {

			fmt.Fprint(os.Stderr, "Malformed body\n")
			return nil,bytesConsumed,errors.New("Malformed body")
		}

		/*
			     Find the end of the bulk string length.
			     Example:
			     "\r\n$34\r\nhello..."

							 ^
							stop here

		*/

		index := 4

		for index < len(body) {

			if body[index] >= '0' && body[index] <= '9' {
				index++
			} else {
				break
			}
		}

		digits := body[3:index]

		// Convert the ASCII digits ("34") into an integer.
		elementSize, err := strconv.Atoi(string(digits))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error converting string to integer %s\n", err.Error())
			return nil,bytesConsumed,err
		}
		/*

			   Number of bytes before the payload.

				For example:

				"\r\n$4\r\n"   -> offset = 6
				"\r\n$34\r\n"  -> offset = 7

			    The offset grows as the length field gains more digits.

		*/

		offset := 5 + len(digits)

		if elementSize+offset <= len(body) {
			// Extract the payload and advance the body to the next bulk string.

			arg := body[offset : elementSize+offset]
			argCopy := make([]byte, len(arg))
			copy(argCopy, arg)

			args = append(args, argCopy)
			bytesConsumed+=offset+elementSize
			body = body[offset+elementSize:]

		} else {

			fmt.Fprintf(os.Stderr, "Malformed body\n")
			return  nil,bytesConsumed,errors.New("Malformed body")
		}

	}

	return args,bytesConsumed,nil
}
