package RESP

import "fmt"



func EncodeResponse(res Response) []byte {

	body := res.Body

	switch res.Type {

	case ERROR:
		return fmt.Appendf(nil, "-%s\r\n", body)
	case SIMPLE_STRING:
		return fmt.Appendf(nil, "+%s\r\n", body)

	case NIL:

		return fmt.Appendf(nil, "$-1\r\n")

	case BULK_STRING:

		return fmt.Appendf(nil, "$%d\r\n%s\r\n", len(body), body)
	case INTEGER:
		return fmt.Appendf(nil, ":%s\r\n", body)
	case ARRAY:
		//a resp array is already encoded from the parser
		return res.Body
	case RDBFILE:
		return fmt.Appendf(nil,"$%d\r\n%s",len(body),body)

	default:

		panic("Unknown Response type")
	}

}