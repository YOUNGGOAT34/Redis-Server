package server

import "CacheDB/app/storage"

// this lets a client know that a key in a transaction was modified
func markDirty(key string, writer *storage.Client) {

	storage.WatchedKeysMutex.Lock()
	defer storage.WatchedKeysMutex.Unlock()

	if set, exists := storage.WatchedKeys[key]; exists {

		for client := range set {
			if client != writer {
				client.Dirty = true
			}
		}
	}
}

func hasWildCard(userInput []byte, card byte) (bool, int) {
	for i, char := range userInput {
		if char == card {
			return true, i
		}
	}
	return false, -1
}
