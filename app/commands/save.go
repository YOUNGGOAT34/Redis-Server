package commands

import (
	rdb "CacheDB/app/RDB"
	"CacheDB/app/RESP"
	"CacheDB/app/config"
)

func HandleSave(args [][]byte, rdbConfig *rdb.RDB,replconfig *config.SERVER) RESP.Response {
	if len(args) != 1 {
		return RESP.WrongNumberOfArguments("save")
	}

	err := rdb.SaveRDB(rdbConfig.Dir + "/" + rdbConfig.DbFileName,replconfig)

	if err != nil {
		return RESP.Response{
			Body: []byte(err.Error()),
			Type: RESP.ERROR,
		}
	}

	return RESP.Response{
		Body: []byte("OK"),
		Type: RESP.SIMPLE_STRING,
	}
}
