package server

// Regression test for the data race between createDefaultUser() (invoked for
// every new client connection) and ACL SETUSER's field mutations on the same
// *storage.User (nopass/addPassword/etc.).
//
// storage.UsersMutex protects the Users MAP (insertion/lookup by name).
// storage.User.UserMutex protects an individual User struct's own mutable
// fields (Flags, Passwords, CommandPermissions). Every other reader/writer of
// those fields (see commands/auth.go: getUser, setUser, whoami) correctly
// takes UserMutex around the field access. createDefaultUser() only held
// UsersMutex and read defaultUser.Flags.NoPass with no lock on the User
// itself - a real, reproducible data race with `go test -race`, not just a
// theoretical one, since every accepted client connection calls this.

import (
	"sync"
	"testing"

	"CacheDB/app/storage"
)

func TestCreateDefaultUser_DoesNotRaceWithConcurrentACLMutation(t *testing.T) {
	defaultUser := &storage.User{
		Name:      "default",
		Passwords: make([][32]byte, 0),
		Flags: storage.UserFlags{
			NoPass:  true,
			Enabled: true,
		},
	}

	storage.UsersMutex.Lock()
	if storage.Users == nil {
		storage.Users = make(map[string]*storage.User)
	}
	storage.Users["default"] = defaultUser
	storage.UsersMutex.Unlock()

	var wg sync.WaitGroup

	// Simulate many clients connecting concurrently (each calls
	// createDefaultUser once, exactly as handleClient does).
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &storage.Client{}
			createDefaultUser(client)
		}()
	}

	// Simulate concurrent ACL mutation of the very same User struct's fields,
	// the way `ACL SETUSER default >somepassword` or `ACL SETUSER default
	// nopass` would from another client connection.
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defaultUser.UserMutex.Lock()
			defaultUser.Flags.NoPass = false
			defaultUser.Passwords = append(defaultUser.Passwords, [32]byte{1})
			defaultUser.Flags.NoPass = true
			defaultUser.Passwords = defaultUser.Passwords[:0]
			defaultUser.UserMutex.Unlock()
		}()
	}

	wg.Wait()
	// Correctness here is "go test -race reports nothing" - there is no
	// meaningful final-value assertion since both goroutine groups are
	// racing by design; the point is proving the access pattern is safe.
}
