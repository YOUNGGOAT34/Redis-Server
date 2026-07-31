package RESP

import "fmt"



func EncodeResponse(res Response) []byte {

	switch res.Type {

	case ERROR:
		return encodeError(res.Body)
	case SIMPLE_STRING:
		return encodeSimpleString(res.Body)

	case NIL:
		return encodeNil()
	case BULK_STRING:
		return encodeBulkString(res.Body)
	case INTEGER:
		return encodeInteger(res.Body)
	case ARRAY:
		return encodeArray(res.Array)
	case RDBFILE:
		return encodeRDB(res.Body)
	case LIST,STREAM,TRANSACTION:
		//a lists,streams,transaction returns is already encoded as a resp array
		return res.Body
	default:

		panic("Unknown Response type")
	}

}

func encodeArray(values []Response) []byte {
	var respArray []byte
	respArray = fmt.Appendf(respArray, "*%d\r\n", len(values))

	for _, value := range values {
		respArray =append(respArray, EncodeResponse(value)...)
	}

	return respArray
}


func encodeBulkString(body []byte) []byte{
	   return fmt.Appendf(nil, "$%d\r\n%s\r\n", len(body), body)
}

func encodeInteger(body []byte) []byte{
	  return fmt.Appendf(nil, ":%s\r\n", body)
}

func encodeNil() []byte{
	  return fmt.Appendf(nil, "$-1\r\n")
}

func encodeRDB(body []byte) []byte{
	  return fmt.Appendf(nil,"$%d\r\n%s",len(body),body)
}


func encodeSimpleString(body []byte) []byte{
	  return fmt.Appendf(nil, "+%s\r\n", body)
}

func encodeError(body []byte) []byte{
    	return fmt.Appendf(nil, "-%s\r\n", body)
  }