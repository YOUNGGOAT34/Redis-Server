package commands

import (
	"CacheDB/app/RESP"
	"CacheDB/app/ZSET"
	"CacheDB/app/storage"
	"fmt"
	"strconv"
)

func ZaddCommand(args [][]byte) RESP.Response {
	if len(args) < 3 || (len(args)-1)%2 != 0 {
		return RESP.WrongNumberOfArguments("ZADD")
	}

	key := string(args[0])

	var zs *zset.ZSet

   storage.DatabaseMutex.Lock()
   data, exists := storage.Database[key]; 
	
	if exists {

		if data.Type != storage.ZSET {
			storage.DatabaseMutex.Unlock()
			return RESP.WrongType()
		}

		zs= data.Value.(*zset.ZSet)
	} else {

		zs= &zset.ZSet{
			Dict: make(map[string]*zset.SkipNode),
			List: zset.NewSkipList(),
		}

		storage.Database[key] = storage.Data{
			Value: zs,
			Type:  storage.ZSET,
		}
	}

	storage.DatabaseMutex.Unlock()
    
	zs.ZSMutex.Lock()
	defer zs.ZSMutex.Unlock()
	count, err := add(args[1:], zs)
	if err != nil {
		return RESP.Response{
			Body: []byte(err.Error()),
			Type: RESP.ERROR,
		}
	}

	return RESP.Response{
		Body: fmt.Appendf([]byte{}, "%d",count),
		Type: RESP.INTEGER,
	}

}

func add(args [][]byte, zs *zset.ZSet) (int, error) {

	var count int = 0
	for i := 0; i < len(args); i += 2 {

		score, err := strconv.ParseFloat(string(args[i]), 64)

		if err != nil {
			return 0, err
		}

		node := &zset.SkipNode{
			Member: string(args[i+1]),
			Score:  score,
		}

		//if deleted is true it means an existing member was updated
		deleted := zs.Add(node)
		if !deleted {
			count++
		}
	}

	return count, nil
}
