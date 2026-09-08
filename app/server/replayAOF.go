package server

import (
	aof "CacheDB/app/AOF"
	rdb "CacheDB/app/RDB"
	"CacheDB/app/RESP"
	"CacheDB/app/config"
	"CacheDB/app/storage"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// replayAOF replays previously-persisted commands from the append-only file
// back through the normal command-dispatch path. replayUser is the trusted
// identity (the server's own seeded default user) used for these replayed
// commands: replay is internal state reconstruction of writes that were
// already authorized once when originally executed, not new client input,
// so it must not be subject to a fresh NOAUTH/ACL rejection.
func replayAOF(replConfig *config.SERVER, rdbConfig *rdb.RDB, aofFileConfig *aof.AOF, replayUser *storage.User) error {

	aofDir := filepath.Join(aofFileConfig.Dir, aofFileConfig.AppendDirName)

	manifestPath := filepath.Join(aofDir, aof.BuildManifestFileName(aofFileConfig.AppendFilename))

	aoffilename, err := aof.ReadManifest(manifestPath)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\r\n", err.Error())
	}

	aofPath := filepath.Join(aofDir, aoffilename)

	file, err := os.Open(aofPath)

	if err != nil {
		return err
	}

	request := make([]byte, 0, 1024)
	temp := make([]byte, 1024)

	for {

		bytesRead, err := file.Read(temp)

		if err != nil && err != io.EOF {
			return err
		}

		if bytesRead > 0 {

			request = append(request, temp[:bytesRead]...)
		}

		for len(request) > 0 {

			parsedRequest, bytesConsumed, err := RESP.ParseRequest(request)

			if err != nil {
				if errors.Is(err, RESP.ErrIncomplete) {
					break
				}

				return err
			}

			dispatchCommands(&storage.Client{User: replayUser}, parsedRequest, replConfig, rdbConfig, aofFileConfig)

			request = request[bytesConsumed:]

		}

		if err == io.EOF {

			if len(request) > 0 {
				return errors.New("Incomplete command at the end of AOF")
			}
			break
		}

	}

	return nil

}
