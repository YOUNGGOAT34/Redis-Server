package sentinel

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)


func (node *Node) HandleAppendEntries(req AppendEntriesRequest) AppendEntriesResponse{
	   node.nodeMutex.Lock()
		defer node.nodeMutex.Unlock()

		response:=AppendEntriesResponse{
			 Term: node.currentTerm,
			 Success: false,
		}

		if req.Term<node.currentTerm{
			  log.Printf("[%s] rejecting append entries from %s---stale term",node.ID,req.LeaderId)
			  return response
		}

		if req.Term>node.currentTerm{
			  node.currentTerm=req.Term
			  node.votedFor=""
			  response.Term=node.currentTerm
		}

		node.state=Follower

		select{
			case node.heartbeatCh<-req:
			default:
		}

		if req.PrevLogIndex>0{
			  prevEntry,ok:=node.raftLog.GetEntry(req.PrevLogIndex)

			  if !ok || prevEntry.Term!=req.PrevLogTerm{
				  log.Printf("[%s] log inconsistency at index %d --rejecting",node.ID,req.PrevLogIndex)
				  return response
			  }

		}


		for i,entry:=range req.Entries{
			  existingEntry,ok:=node.raftLog.GetEntry(entry.index)

			  if ok && existingEntry.Term != entry.Term{
				   log.Printf("[%s] Conflicting entry at index %d---truncating ",node.ID,entry.index)
					node.raftLog.TruncateFrom(entry.index)
			  }

			  if !ok && i>=node.raftLog.GetLastIndex(){
				   node.raftLog.Append(req.Entries[i:]...)
					break
			  }
		}
     
		if req.LeaderCommit>node.raftLog.GetCommitIndex(){
			  lastNewIndex:=node.raftLog.commitIndex
			  newCommitIndex:=req.LeaderCommit

			  if lastNewIndex<newCommitIndex{
				   newCommitIndex=lastNewIndex
			  }

			  node.raftLog.SetCommitIndex(newCommitIndex)

			  for i:=node.raftLog.GetLastAppliedIndex()+1;i<=newCommitIndex;i++{
				     entry,ok:=node.raftLog.GetEntry(i)

					  if ok{
						   node.applyEntry(entry)
					  }
			  }

		}

		response.Success=true
		return response
}


func (node *Node) sendHeartBeats(){
	  node.nodeMutex.RLock()
	  currentTerm:=node.currentTerm
	  node.nodeMutex.RUnlock()

	  for _,peer:=range node.Peers{
		  go func(peer string){
			    node.nodeMutex.RLock()
				 nextIdx:=node.nextIndex[peer]
				 node.nodeMutex.RUnlock()

				 prevLogIndex:=nextIdx-1
				 prevLogTerm:=Term(0)

				 if prevLogIndex>0{
					  if entry,ok:=node.raftLog.GetEntry(prevLogIndex);ok{
						   prevLogTerm=entry.Term
					  }
				 }

				 entries:=node.raftLog.GetFrom(nextIdx)

				 req:=AppendEntriesRequest{
					   Entries: entries,
						PrevLogIndex: prevLogIndex,
						PrevLogTerm: prevLogTerm,
						Term: currentTerm,
						LeaderId: node.ID,
						LeaderCommit: node.raftLog.commitIndex,
				 }

				 res,err:=node.sendAppendEntries(peer,req)

				 if err!=nil{
					    log.Printf("[%s] failed to reach peer %s:%w",node.ID,peer,err)
						 return
				 }

				 if res.Term>currentTerm{
					 node.becomeFollower(currentTerm)
					 return
				 }

				 if res.Success{
					  node.nodeMutex.Lock()
					  if len(entries)>0{
						  node.nextIndex[peer]=entries[len(entries)-1].index+1
						  node.matchIndex[peer]=entries[len(entries)-1].index
					  }
					  node.nodeMutex.Unlock()
				 }else{
					  node.nodeMutex.Lock()
					  if node.nextIndex[peer]>1{
						     node.nextIndex[peer]--
					  }
					  node.nodeMutex.Unlock()
				 }


		  }(peer)
	  }
}

func (node *Node) sendAppendEntries(peer string, req AppendEntriesRequest) (AppendEntriesResponse,error){
      
	conn,err:=net.DialTimeout("tcp",peer,2*time.Second)

	if err!=nil{
		   return AppendEntriesResponse{},fmt.Errorf("failed to connect to %s:%w",peer,err)
	}

	defer conn.Close()

	message:=map[string]interface{}{
		  "type":"AppendEntries",
		  "data":req,
	}

	encoder:=json.NewEncoder(conn)

	if err=encoder.Encode(message);err!=nil{
		  return AppendEntriesResponse{},fmt.Errorf("failed to encode AppendEntries: %w",err)
	}

	var resp AppendEntriesResponse

	decoder:=json.NewDecoder(conn)

	if err=decoder.Decode(&resp);err!=nil{
		   return AppendEntriesResponse{},fmt.Errorf("failed to decode AppendEntries: %w",err)
	}

	return resp,nil
}