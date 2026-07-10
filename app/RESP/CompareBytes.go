package RESP

import "bytes"

func CompareBytes(a, b []byte) bool {
	if bytes.EqualFold(a, b) {
		return true
	}

	return false
}
