package commands

import (
	"CacheDB/app/RESP"
	"CacheDB/app/config"
)

func Keys(args [][]byte,replconfig *config.SERVER) RESP.Response {
	if len(args) != 1 {
		return RESP.WrongNumberOfArguments("KEYS")
	}

	if RESP.CompareBytes(args[0], []byte("*")) {
		responses := make([]RESP.Response, 0, len(replconfig.Database))
		replconfig.DatabaseMutex.RLock()
		for key := range replconfig.Database {
			responses = append(responses, RESP.Response{
				Body: []byte(key),
				Type: RESP.BULK_STRING,
			})
		}

		replconfig.DatabaseMutex.RUnlock()
		return RESP.Response{
			Array: responses,
			Type:  RESP.ARRAY,
		}
	}

	exists, index := hasWildCard(args[0], '*')

	if exists {

		prefix := string(args[0][:index])

		matchingKeys := collectMatchingKeys(func(key string) bool {
			return startsWith(key, prefix)
		},replconfig)

		return RESP.Response{
			Array: matchingKeys,
			Type:  RESP.ARRAY,
		}
	}

	exists, index = hasWildCard(args[0], '?')

	if exists {

		prefix := string(args[0][:index])

		matchingKeys := collectMatchingKeys(func(key string) bool {
			return startsWith(key, prefix) && len(prefix)+1 == len(key)
		},replconfig)

		return RESP.Response{
			Array: matchingKeys,
			Type:  RESP.ARRAY,
		}
	}

	return RESP.Response{}

}

func startsWith(key string, pattern string) bool {
	if len(key) < len(pattern) {
		return false
	}

	for i := 0; i < len(pattern); i++ {
		if pattern[i] != key[i] {
			return false
		}

	}

	return true
}

func collectMatchingKeys(matches func(string) bool,replconfig *config.SERVER) []RESP.Response {
	replconfig.DatabaseMutex.RLock()

	count := 0
	for key := range replconfig.Database {
		if matches(key) {
			count++

		}
	}
	matchingKeys := make([]RESP.Response, 0, count)

	for key := range replconfig.Database {

		if matches(key) {
			matchingKeys = append(matchingKeys, RESP.Response{
				Body: []byte(key),
				Type: RESP.BULK_STRING,
			})
		}
	}

	replconfig.DatabaseMutex.RUnlock()

	return matchingKeys
}
