package sentinel


type Term int

type NodeState int

const(
	   Follower NodeState=iota
		Candidate
		Leader
)

func (state NodeState) NodeStateToString() string{
	  
	  switch state{
		case Follower:
			return "Follower"
		case Candidate:
			 return "Candidate"
		case Leader:
			 return "Leader"
		default:
			 return "Unknown"
	  }

}

type LogEntry struct{
      Term Term
		index int
		Command string
}


/*
    sent by a candidate to all other nodes requesting
	 for their votes in an election
	 contains:
	 	candidate's term
		candidate's node id
		index of candidate's last log entry
		term of candidate's last log entry
*/

type RequestVoteRequest struct{
	   Term Term
		CandidateId string
		LastLogIndex int
		LastLogTerm Term
}


/*
  Reply to a RequestVote message
  Contains:
		Responder's Term-->if responder has a higher term ,the candidate updates itself
		did the node vote for the candidate??(boolean)-->true if the node voted for the candidate

*/

type RequestVoteResponse struct{
	   Term Term
		VoteGranted bool
}

/*
   sent by the leader to all its followers
	 servers two purposes:
	  			Heartbeat: empty Entries slice to prove leader is alive
				Log replication: Entries contains new commands to append
	 contains:
	 	 leader's current term
		 Leader's Id
		 previous Log index
		 previous log term
		 Entries slice--empty for heartbeat
		 Leader's commit index

*/

type AppendEntriesRequest struct{
	  Term Term
	  LeaderId string
	  PrevLogIndex int
	  PrevLogTerm Term
	  Entries []LogEntry
	  LeaderCommit int
}

/*
    A response to appendEntriesRequest

	 contains:
	 		responder's term
			did the follower accept the entries??(bool)

*/

type AppendEntriesResponse struct{
	  Term Term
	  Success bool
}

/*
    Cluster events:
			MasterUp-->master responded to ping
			MasterDown-->master missed too many pings
			ReplicaUp-->replica responded to ping
			ReplicaDown-->replica missed too many pings
			FailoverDone-->a replica has been promoted
		
*/


type ClusterEvent int

const (
	  MasterUp ClusterEvent=iota
	  MasterDown
	  ReplicaUp
	  ReplicaDown
	  FailoverDone
)

/*
    This represents a node in the cluster being monitored
	 contains:
	 	address-->i.e localhost:6370
		Is it the master?
		Replication offset-->higher means more up to date
		Failed consecutive pings
		Is the node currently reachable??
*/

type CacheDBNode struct{
	  Address string
	  IsMaster bool
	  ReplOffset int64
	  FailedPings int
	  IsAlive bool
}


