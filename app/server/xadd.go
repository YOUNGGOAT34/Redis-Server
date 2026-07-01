package server

import (
	"errors"
	"strconv"
)

func createStreamID(id []byte) (StreamID,error){
	    //find the hyphen in the user's id

		 hyphenIndex:=-1
		 for index,char:=range id{
			    if char=='-'{
					   hyphenIndex=index
						break
				 }
		 }


		 if hyphenIndex==-1{
			   return StreamID{},errors.New("invalid stream Id")
		 }


		 milliseconds,err:=strconv.ParseUint(string(id[0:hyphenIndex]),10,64)
		 if err!=nil{
			   return StreamID{},err
		 }
		 sequence,err:=strconv.ParseUint(string(id[hyphenIndex+1:]),10,64)

		 if err!=nil{
			   return StreamID{},err
		 }

		 return StreamID{
			     Milliseconds:milliseconds,
				  Sequence:sequence,
		 },err
}


func xaddCommand(arguments [][]byte) Response {
	     if len(arguments) <4{

						return Response{
						Body:[]byte("Wrong number of arguments for 'XADD' command"),
						Type:ERROR,

				         }
		  }


		if len(arguments[2:])%2!=0{
			   return Response{
						Body:[]byte("Error wrong number of field-value arguments"),
						Type:ERROR,

				         }
		}


		var stream *Stream


		databaseMutex.Lock()
		defer databaseMutex.Unlock()

		data,exists:=database[string(arguments[0])]

		if exists{
			    if data.Type!=STREAM{
					    return Response{
							Body:[]byte("WRONGTYPE Operation against a key holding the wrong kind of value"),
							Type:ERROR,
						  }
				 }

				 stream=data.Value.(*Stream)


		}else{
			  stream=&Stream{}
			  database[string(arguments[0])]=Data{
			      Type: STREAM,
		  			Value: stream,
		     }
		}


		
		var Id StreamID

		if string(arguments[1])=="*"{
			     Id=stream.NextID()
		}else{
			    var err error
			    Id,err=createStreamID(arguments[1])

				 if err!=nil{
					   return Response{
							Body:[]byte("Invalid stream Id"),
							Type:ERROR,
						  }
				 }
		}

      fields:=make(map[string][]byte)

		for i:=2;i<len(arguments);i+=2{
			    fields[string(arguments[i])]=arguments[i+1]
		}

		entry:=&StreamEntry{
			     ID: Id,
				  Fields: fields,
		}


      stream.LastID=Id
		stream.Entries=append(stream.Entries, entry)
     stream.Len++
   
		return Response{
			   Body: []byte(Id.String()),
				Type: BULK_STRING,
		}

}