package server

import (
	"CacheDB/app/RDB"
	"CacheDB/app/RESP"
)

func handleSave(args [][]byte,rdbConfig *rdb.RDB) RESP.Response {
    if len(args) != 1 {
        return RESP.WrongNumberOfArguments("save")
    }

    err := rdb.SaveRDB(rdbConfig.Dir+"/"+rdbConfig.DbFileName)

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