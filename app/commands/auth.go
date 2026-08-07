package commands

import (
	"CacheDB/app/RESP"
	"CacheDB/app/storage"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func Acl(client *storage.Client, args [][]byte) RESP.Response {

	if len(args) < 1 {
		return RESP.WrongNumberOfArguments("ACL")
	}

	switch strings.ToUpper(string(args[0])) {
	case "WHOAMI":
		return whoami(client, args[1:])
	case "GETUSER":
		return getUser(args[1:])
	case "SETUSER":
		return setUser(args[1:])
	default:
		return RESP.Response{
			Body: []byte("unknown acl subcommand"),
			Type: RESP.ERROR,
		}
	}

}

func getUser(args [][]byte) RESP.Response {
	if len(args) != 1 {
		return RESP.WrongNumberOfArguments("ACL GETUSER")
	}

	storage.UsersMutex.RLock()
	user, exists := storage.Users[string(args[0])]
	storage.UsersMutex.RUnlock()

	if !exists {
		return Invalid()
	}

	passwordHashes := make([][32]byte, len(user.Passwords))
	user.UserMutex.RLock()
	copy(passwordHashes, user.Passwords)
	userFlags := user.Flags
	user.UserMutex.RUnlock()

	//build flags array

	flags := make([]RESP.Response, 0)
	if userFlags.NoPass {
		flags = append(flags, RESP.Response{
			Body: []byte("nopass"),
			Type: RESP.BULK_STRING,
		})
	}

	if userFlags.Enabled {
		flags = append(flags, RESP.Response{
			Body: []byte("on"),
			Type: RESP.BULK_STRING,
		})
	}

	//build an array of hex passwords

	passwords := make([]RESP.Response, 0, len(passwordHashes))
	for _, hash := range passwordHashes {
		passwords = append(passwords,
			RESP.Response{
				Body: []byte(hex.EncodeToString(hash[:])),
				Type: RESP.BULK_STRING,
			})
	}

	return RESP.Response{
		Type: RESP.ARRAY,
		Array: []RESP.Response{
			{
				Type: RESP.BULK_STRING,
				Body: []byte("flags"),
			},

			{
				Type:  RESP.ARRAY,
				Array: flags,
			},

			{
				Type: RESP.BULK_STRING,
				Body: []byte("passwords"),
			},

			{
				Type:  RESP.ARRAY,
				Array: passwords,
			},
		},
	}
}

func whoami(client *storage.Client, args [][]byte) RESP.Response {
	if len(args) != 0 {
		return RESP.WrongNumberOfArguments("ACL WHOAMI")
	}

	return RESP.Response{
		Body: []byte(client.User.Name),
		Type: RESP.BULK_STRING,
	}

}

func setUser(args [][]byte) RESP.Response {
	if len(args) < 2 {
		return RESP.WrongNumberOfArguments("ACL SETUSER")
	}

	user:=createUser(string(args[0]))

	user.UserMutex.Lock()
	defer user.UserMutex.Unlock()
	for _, arg := range args[1:] {

		rule := string(arg)

		switch {
		case strings.HasPrefix(rule, ">"):
			res := addPassword(user, arg[1:])
			if res.Type == RESP.ERROR {
				return res
			}
		case strings.HasPrefix(rule, "<"):
			res := removePassword(user, arg[1:])
			if res.Type == RESP.ERROR {
				return res
			}
		case strings.HasPrefix(rule, "+"):
			res := grantOrRevokePermission(user, arg[1:],true)
			if res.Type == RESP.ERROR {
				return res
			}
		case strings.HasPrefix(rule, "-"):
			res := grantOrRevokePermission(user, arg[1:],false)
			if res.Type == RESP.ERROR {
				return res
			}
		case rule == "nopass":
			res := nopass(user)
			if res.Type == RESP.ERROR {
				return res
			}
		case rule == "on":
			res := disableOrEnableUser(user, true)
			if res.Type == RESP.ERROR {
				return res
			}
		case rule == "off":
			res := disableOrEnableUser(user, false)
			if res.Type == RESP.ERROR {
				return res
			}
		case rule == "resetpass":
			res := reset(user, false)
			if res.Type == RESP.ERROR {
				return res
			}
		case rule == "reset":
			res := reset(user, true)
			if res.Type == RESP.ERROR {
				return res
			}
		case rule == "nocommands":
			res := revokeOrGrantAllPermissions(user, false)
			if res.Type == RESP.ERROR {
				return res
			}
		case rule == "allcommands":
			res := revokeOrGrantAllPermissions(user, true)
			if res.Type == RESP.ERROR {
				return res
			}

		default:
			return RESP.Response{
				Body: []byte("syntax error"),
				Type: RESP.ERROR,
			}
		}
	}

	return RESP.Response{
		Body: []byte("OK"),
		Type: RESP.SIMPLE_STRING,
	}
}

func createUser(username string) *storage.User{
	storage.UsersMutex.Lock()
	defer storage.UsersMutex.Unlock()
	if user, exists := storage.Users[username]; exists{
		  return user
	}

	user:=&storage.User{
		Name: username, 
		Passwords: make([][32]byte, 0),
		Flags: storage.UserFlags{
			NoPass:  false,
			Enabled: false,
		},
	}

    storage.Users[username]=user
	 return user
}

/*
	   resetUser will help in determining whether user flags need to be disabled (if true)
		if the call is from reset-->this flag will be true
		if the call is from resetpass -->this flag will be false
*/
func reset(user *storage.User, resetUser bool) RESP.Response {
	
		user.Passwords=user.Passwords[:0]
		if resetUser {
			user.Flags.Enabled = false
			user.Flags.NoPass = false
		}

		return RESP.Response{
			Body: []byte("OK"),
			Type: RESP.SIMPLE_STRING,
		}

}

func nopass(user *storage.User) RESP.Response {
		user.Flags.NoPass = true
		user.Passwords=user.Passwords[:0]

	return RESP.Response{
		Body: []byte("OK"),
		Type: RESP.SIMPLE_STRING,
	}
}

func disableOrEnableUser(user *storage.User, on bool) RESP.Response {

		user.Flags.Enabled = on

		return RESP.Response{
			Body: []byte("OK"),
			Type: RESP.SIMPLE_STRING,
		}
}

func addPassword(user *storage.User, password []byte) RESP.Response {

		hash := sha256.Sum256(password)
		//don't store the same hash twice
		for _, passwordHash := range user.Passwords {
			if passwordHash == hash {
				return RESP.Response{
					Body: []byte("OK"),
					Type: RESP.SIMPLE_STRING,
				}
			}
		}
		user.Passwords = append(user.Passwords, hash)
		user.Flags.NoPass = false

		return RESP.Response{
			Body: []byte("OK"),
			Type: RESP.SIMPLE_STRING,
		}
}

func removePassword(user *storage.User, password []byte) RESP.Response {
	hash := sha256.Sum256(password)


	for index, password := range user.Passwords {
		if password == hash {
			user.Passwords = append(user.Passwords[:index], user.Passwords[index+1:]...)
			break
		}
	}

	return RESP.Response{
		Body: []byte("OK"),
		Type: RESP.SIMPLE_STRING,
	}
	
}

func Auth(client *storage.Client, args [][]byte) RESP.Response {
	if len(args) != 1 && len(args) != 2 {
		return RESP.WrongNumberOfArguments("AUTH")
	}

	

	var username string
	var givenPassword []byte

	if len(args) == 1 {
		givenPassword = args[0]
		username = "default"
	} else {
		givenPassword = args[1]
		username = string(args[0])
	}

	storage.UsersMutex.RLock()
	user, exists := storage.Users[username]
	storage.UsersMutex.RUnlock()

	if exists {
		user.UserMutex.RLock()
		defer user.UserMutex.RUnlock()
		if !user.Flags.Enabled {
			return Invalid()
		}
		hash := sha256.Sum256(givenPassword)
		for _, password := range user.Passwords {
			if password == hash {
				client.User = user
				return RESP.Response{
					Body: []byte("OK"),
					Type: RESP.SIMPLE_STRING,
				}
			}
		}

		return Invalid()
	}

	return Invalid()
}

func Invalid() RESP.Response {
	return RESP.Response{
		Body: []byte("WRONGPASS invalid username-password pair or user is disabled"),
		Type: RESP.ERROR,
	}
}

func grantOrRevokePermission(user *storage.User, command []byte,grant bool) RESP.Response {
	
   flag:=strings.ToUpper(string(command))

	switch flag{
	       case "@ALL":
				 revokeOrGrantAllPermissions(user,grant)
			 default:
				CMD,exists:= storage.CommandToPermission[flag]
				if !exists{
					  return RESP.Response{
							 Body:[]byte("ERR trying to set an unknown command"),
							 Type: RESP.ERROR,
					  }
				}
			
				if grant{
					user.CommandPermissions |= CMD
				}else{
					user.CommandPermissions &^= CMD
				}
	}

	return RESP.Response{
		Body: []byte("OK"),
		Type: RESP.SIMPLE_STRING,
	}
}

func revokeOrGrantAllPermissions(user *storage.User, grant bool) RESP.Response {

	if grant {
		user.CommandPermissions = storage.AllCommands
	} else {
		user.CommandPermissions = 0
	}

	return RESP.Response{
		Body: []byte("OK"),
		Type: RESP.SIMPLE_STRING,
	}
	
}
