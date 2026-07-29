package server

import (
	"CacheDB/app/RESP"
	"CacheDB/app/storage"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

func acl(client *storage.Client,args [][]byte) RESP.Response {

	if len(args)<1{
		  return RESP.WrongNumberOfArguments("ACL")
	}

	switch strings.ToUpper(string(args[0])){
			case "WHOAMI":
				return whoami(client,args[1:])
			case "GETUSER":
				return getUser(args[1:])
			case "SETUSER":
				return setUser(args[1:])
	}

	return RESP.Response{}
	
}

func getUser(args [][]byte) RESP.Response {
	   if len(args)!=1{
			    return RESP.WrongNumberOfArguments("ACL GETUSER")
		}


		storage.UserMutex.Lock()
		user,exists:=storage.Users[string(args[0])]
		if !exists{
			  return invalid()
		}
		passwordHashes:=make([][32]byte,len(user.Passwords))
		copy(passwordHashes,user.Passwords)
		userFlags:=user.Flags
		storage.UserMutex.Unlock()

		//build flags array

		flags:=make([]RESP.Response,0)
		if userFlags.NoPass{
			  flags = append(flags, RESP.Response{
				         Body: []byte("nopass"),
							Type: RESP.BULK_STRING,
			  })
		}


		if userFlags.Enabled{
			  flags = append(flags, RESP.Response{
				         Body: []byte("on"),
							Type: RESP.BULK_STRING,
			  })
		}

	
	 //build an array of hex passwords

	 passwords:=make([]RESP.Response,0,len(passwordHashes))
	 for _,hash:=range passwordHashes{
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
						   Type: RESP.ARRAY,
							Array:flags,
					  },

					    {
						  Type: RESP.BULK_STRING,
						  Body: []byte("passwords"),
					  },

					  {
						   Type: RESP.ARRAY,
							Array:passwords,
					  },
			},
		}
}

func whoami(client *storage.Client,args [][]byte) RESP.Response {
	if len(args)!=0{
		 return RESP.WrongNumberOfArguments("ACL WHOAMI")
	}

	return RESP.Response{
		Body: []byte(client.User.Name),
		Type: RESP.BULK_STRING,
	}

}


func setUser(args [][]byte) RESP.Response{
	   if len(args)<2{
			  return RESP.WrongNumberOfArguments("ACL SETUSER")
		}

		for _,arg:=range args[1:]{

			rule:=string(arg)
        
			switch{
					case strings.HasPrefix(rule,">"):
						return addPassword(string(args[0]),arg[1:])
					case strings.HasPrefix(rule,"<"):
						//remove password
					case rule=="nopass":
						//enable nopass
					case rule=="on":
						return disableOrEnableUser(string(args[0]),true)
					case rule=="off":
						return disableOrEnableUser(string(args[0]),false)
					default:
						//syntax error
			}
		}

	return RESP.Response{}

}

func disableOrEnableUser(username string,on bool) RESP.Response {
	    storage.UserMutex.Lock()
		 defer storage.UserMutex.Unlock()

		 user,exists:=storage.Users[username]

		 if exists{
			  user.Flags.Enabled=on
					return RESP.Response{
								Body: []byte("OK"),
								Type: RESP.SIMPLE_STRING,
						}
		 }


		return RESP.Response{
        Type: RESP.ERROR,
        Body: fmt.Appendf(nil, "ERR User '%s' does not exist", username),
    }
}

func  addPassword(username string,password []byte) RESP.Response{
	    hash:=sha256.Sum256(password)

		 storage.UserMutex.Lock()
		 defer storage.UserMutex.Unlock()

		 if user,exists:=storage.Users[username];exists{
			     user.Passwords = append(user.Passwords, hash)
				  user.Flags.NoPass=false
				  return RESP.Response{
					   Body: []byte("OK"),
						Type: RESP.SIMPLE_STRING,
				  }
		 }


		 return invalid()

}

func auth(client *storage.Client,args [][]byte) RESP.Response{
	      if len(args)!=2{
				 return RESP.WrongNumberOfArguments("AUTH")
			}


			storage.UserMutex.RLock()
			defer storage.UserMutex.RUnlock()

			if user,exists:=storage.Users[string(args[0])];exists{
				      if !user.Flags.Enabled{
							  return invalid()
						}
				      for _,password:=range user.Passwords{
							   if password==sha256.Sum256(args[1]){
									  client.User=user
									   return RESP.Response{
											  Body: []byte("OK"),
											  Type: RESP.SIMPLE_STRING,
										}
								}
						}

						return invalid()
			}

			
       return invalid()
			
}



func invalid() RESP.Response{
	  return RESP.Response{
							    Body: []byte(" WRONGPASS invalid username-password pair or user is disabled"),
								 Type: RESP.ERROR,
						}
}

