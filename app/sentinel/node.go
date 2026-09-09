package sentinel

import (
	"log"
	"math/rand"
	"sync"
	"time"
)


const(
	  HeartbeatInterval=50*time.Millisecond
	  ElectionTimeoutMin=50*time.Millisecond
     ElectionTimeoutMax=50*time.Millisecond

)

type Node struct{
	   /*
		    Identity:
			 	 Id
				 Address
				 Peers--their addresses
		*/

		ID string
		Address string
		Peers []string

		/*
		    Raft state-->updated automatically:
			 		state-->follower,candidate,leader
					current term-->latest term this node has seen
					voted for-->candidate that this node voted for,,empty string means the node has not voted for anyone yet
		*/
		nodeMutex sync.RWMutex

		state NodeState
		currentTerm Term
		votedFor string

		raftLog *RaftLog


		//volatile leader state
		nextIndex map[string]int
		matchIndex map[string]int

		/*
		    event channels and timers for internal communication
			  heartbeat channel-->receives heartbeat/appendentries
			  vote channel-->receives a vote request
			  commit channel-->new entries to commit
			  reset election timer-->resets the election countdown timer
			  apply log entry channel-->delivers commited log entries to the application layer
			  stop chan-->for graceful shutdown
			  wg-->Tracks background goroutines

		*/

		heartbeatCh chan AppendEntriesRequest
		voteCh chan RequestVoteRequest
		commitCh chan int

		resetElectionTimer chan struct{}
		applyCh chan LogEntry
		stopCh chan struct{}
		wg sync.WaitGroup

		//cacheDB cluster being monitored
		clusterMutex sync.RWMutex
		master *CacheDBNode
		replicas []*CacheDBNode
}



func NewNode(id,address string,peers []string) *Node{
	 return &Node{
		   ID: id,
			Address: address,
			state: Follower,
			Peers: peers,
			currentTerm: 0,
			votedFor: "",
			raftLog: NewRaftLog(),
			nextIndex: make(map[string]int),
			matchIndex: make(map[string]int),
			heartbeatCh: make(chan AppendEntriesRequest,100),
			voteCh: make(chan RequestVoteRequest,100),
			commitCh: make(chan int,100),
			resetElectionTimer: make(chan struct{},1),
			stopCh: make(chan struct{}),
			applyCh: make(chan LogEntry,100),
	 }
}


func (node *Node) Start(){
	  log.Printf("[%s] Starting as %s in term %d",node.ID,node.state.NodeStateToString(),node.currentTerm)
	  node.wg.Add(1)
	  go node.run()
}

func (node *Node) Stop(){
	  close(node.stopCh)
	  node.wg.Wait()
}

func (node *Node) run(){
	   defer node.wg.Done()

		for{
			  select{
				case<-node.stopCh:
					return
				 default:
			  }

			  node.nodeMutex.RLock()
			  state:=node.state
			  node.nodeMutex.RUnlock()

			  switch state{
			  case Follower:
			  case Candidate:
			  case Leader:
			  }
		}
}


func randomElectionTimeout() time.Duration{
	  diff:=ElectionTimeoutMax-ElectionTimeoutMin
     return ElectionTimeoutMin+time.Duration(rand.Int63n(int64(diff)))
	  
}

func (node *Node)becomeFollower(term Term){
       node.nodeMutex.Lock()
		 defer node.nodeMutex.Unlock()
	    log.Printf("[%s] Becoming follower in term %d (was %s in term %d)",node.ID,term,node.state.NodeStateToString(),node.currentTerm)
		 node.state=Follower
		 node.currentTerm=term
		 node.votedFor=""

}

func (node *Node) becomeCandidate(){
	  node.nodeMutex.Lock()
	  defer node.nodeMutex.Unlock()

	  node.state=Candidate
	  node.currentTerm++
	  node.votedFor=node.ID

	  log.Printf("[%s] Becoming candidate in term %d",node.ID,node.currentTerm)
}

func (node *Node) becomeLeader(){
	  node.nodeMutex.Lock()
	  defer node.nodeMutex.Unlock()
     log.Printf("[%s] Becoming leader at term %d",node.ID,node.currentTerm)
	  node.state=Leader

	  lastIndex:=node.raftLog.GetLastIndex()

	  for _,peer:=range node.Peers{
		     node.nextIndex[peer]=lastIndex+1
			  node.matchIndex[peer]=0
	  }

}

func (node *Node) runFollower(){
	  timeout:=randomElectionTimeout()

	  timer:=time.NewTimer(timeout)
	  defer timer.Stop()

	  log.Printf("[%s] Follower waiting with timeout %v",node.ID,timeout)

	  for{
		  select{
			case <-node.stopCh:
				return
			case <-timer.C:
				log.Printf("[%s] Election timeout --- starting an election",node.ID)
				node.becomeCandidate()
				return
			case  req:=<-node.heartbeatCh:
				node.nodeMutex.RLock()

				if req.Term<node.currentTerm{
					 node.nodeMutex.RUnlock()
					 continue
				}

				if req.Term>node.currentTerm{
					 node.becomeFollower(req.Term)
				}

				node.nodeMutex.RUnlock()

				if !timer.Stop(){
					  select{
					  case <-timer.C:
					  default:
					  }
				}

				timer.Reset(randomElectionTimeout())
			case req:=<-node.voteCh:
				node.nodeMutex.Lock()
				if req.Term<node.currentTerm{
					 node.nodeMutex.Unlock()
					 continue
				}

				if req.Term>node.currentTerm{
					  node.becomeFollower(req.Term)
				}

				canVote:=node.votedFor=="" || node.votedFor==req.CandidateId
				logIsUptoDate:=node.raftLog.IsUpToDate(req.LastLogTerm,req.LastLogIndex)

				voteGranted:=false
				if canVote && logIsUptoDate{
					 node.votedFor=req.CandidateId
					 voteGranted=true
				
					}
				node.nodeMutex.Unlock()

					if voteGranted{

						if !timer.Stop(){
							 select{
							 case <-timer.C:
							 default:
							 }
						  }
						  timer.Reset(randomElectionTimeout())
					}
		  }
	  }

}

func (node *Node) runCandidate(){
	  node.nodeMutex.Lock()
	  currentTerm:=node.currentTerm
	  node.nodeMutex.Unlock()

	  votesReceived:=1
	  votesNeeded:=(len(node.Peers)+1)/2+1
	  votesResults:=make(chan bool,len(node.Peers))

	  for _,peer:= range node.Peers{
		     go func(peer string){
				   req:=RequestVoteRequest{
						  CandidateId: node.ID,
						  Term: currentTerm,
						  LastLogIndex: node.raftLog.GetLastIndex(),
						  LastLogTerm: node.raftLog.GetLastTerm(),
					}

					resp,err:=node.sendRequestVote(peer,req)

					if err!=nil{
						  votesResults<-false
						  return
					}

					
			  }(peer)
	  }

}



// func (n *Node) applyEntry(entry LogEntry) {
//     n.raftLog.SetLastApplied(entry.Index)

//     // send to application layer for processing
//     select {
//     case n.applyCh <- entry:
//     default:
//         log.Printf("[%s] applyCh full — dropping entry %d", n.ID, entry.Index)
//     }
// }


func (node *Node) applyEntry(entry LogEntry){
	   node.raftLog.SetLastAppliedIndex(entry.index)

		select{
		case node.applyCh<-entry:
		default:
			 log.Printf("[%s] applyCh full — dropping entry %d", node.ID, entry.index)
		}
}
