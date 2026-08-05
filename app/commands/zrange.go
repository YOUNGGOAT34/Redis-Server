package commands

import (
	"CacheDB/app/RESP"
	zset "CacheDB/app/ZSET"
	"CacheDB/app/storage"
	"strconv"
)

func Zrange(args [][]byte) RESP.Response {

	if len(args) != 3 {
		return RESP.WrongNumberOfArguments("ZRANGE")
	}

	storage.DatabaseMutex.RLock()
	data, exists := storage.Database[string(args[0])]
	storage.DatabaseMutex.RUnlock()

	if !exists {
		return RESP.Response{
			Body: []byte{},
			Type: RESP.ARRAY,
		}
	}

	if data.Type !=storage.ZSET {
		return RESP.WrongType()
	}

	startIndex, err := strconv.Atoi(string(args[1]))
	if err != nil {
		return stringToIntError(err)
	}

	stopIndex, err := strconv.Atoi(string(args[2]))

	if err != nil {
		return stringToIntError(err)
	}

	zs := data.Value.(*zset.ZSet)
   zs.ZSMutex.RLock()
	defer zs.ZSMutex.RUnlock()
	if startIndex < 0 {
		startIndex = zs.List.Length + startIndex
	}

	if stopIndex < 0 {
		stopIndex = zs.List.Length + stopIndex
	}

	if startIndex < 0 {
		startIndex = 0
	}

	if startIndex >= zs.List.Length || startIndex > stopIndex {
		return RESP.Response{
			Body: []byte{},
			Type: RESP.ARRAY,
		}
	}

	if stopIndex >= zs.List.Length {
		stopIndex = zs.List.Length - 1
	}

	res := getElementsInRange(zs, startIndex, stopIndex)

	return RESP.Response{
		Array: res,
		Type:  RESP.ARRAY,
	}

}

func getElementsInRange(zs *zset.ZSet, startIndex, stopIndex int) []RESP.Response {

	currentIndex := 0
	current := zs.List.Head.Forward[0]

	//find the starting node

	for current != nil && currentIndex != startIndex {
		current = current.Forward[0]
		currentIndex++
	}

	res := make([]RESP.Response, 0, stopIndex-startIndex+1)

	for current != nil && currentIndex <= stopIndex {
		res = append(res, RESP.Response{
			Body: []byte(current.Member),
			Type: RESP.BULK_STRING,
		})

		current = current.Forward[0]
		currentIndex++
	}

	return res
}

func stringToIntError(err error) RESP.Response {
	return RESP.Response{
		Body: []byte(err.Error()),
		Type: RESP.ERROR,
	}
}
