package sentinel

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)



func (node *Node) HandleRequestVote(req RequestVoteRequest) RequestVoteResponse{
	   
	   node.nodeMutex.Lock()
		defer node.nodeMutex.Unlock()

		resp:=RequestVoteResponse{
			  Term: node.currentTerm,
			  VoteGranted: false,
		}

		if req.Term<node.currentTerm{
			  log.Printf("[%s] Denying vote to %s, stale term %d<%d",node.ID,req.CandidateId,req.Term,node.currentTerm)
			  return resp
		}

		if req.Term>node.currentTerm{
			  node.currentTerm=req.Term
			  node.state=Follower
			  node.votedFor=""
			  resp.Term=node.currentTerm
		}


		voted:=node.votedFor!="" && node.votedFor!=req.CandidateId

		if voted{
			  log.Printf("[%s] Denying vote to--%s, already voted for",node.ID,req.CandidateId,node.votedFor)
			  return resp
		}

		if !node.raftLog.IsUpToDate(req.LastLogTerm,req.LastLogIndex){
			   log.Printf("[%s] Denying vote to %s---log not up to date",node.ID,req.CandidateId)
				return resp
		}

		node.votedFor=req.CandidateId
		resp.VoteGranted=true

		select{
		   case node.resetElectionTimer<-struct{}{}:
			default:
		}

		return resp
}


func (node *Node)sendVoteRequest(peer string,req RequestVoteRequest) (RequestVoteResponse,error){
	     conn,err:=net.DialTimeout("tcp",peer,2*time.Second)

		  if err!=nil{
			     return RequestVoteResponse{},fmt.Errorf("failed to connect to %s:%w",peer,err)
		  }

		  defer conn.Close()

		  conn.SetDeadline(time.Now().Add(2*time.Second))

		  message:=map[string]interface{}{
			   "type":"RequestVote",
				 "data":req,
		  }

		  encoder:=json.NewEncoder(conn)

		  if err:=encoder.Encode(message);err!=nil{
			     return RequestVoteResponse{},fmt.Errorf("failed to encode RequestVote: %w",err)
		  }

		  var resp RequestVoteResponse
        decoder:=json.NewDecoder(conn)

		  if err=decoder.Decode(&resp);err!=nil{
			   return RequestVoteResponse{},fmt.Errorf("failed to decode response: %w",err)
		  }

		  return resp,nil
}


