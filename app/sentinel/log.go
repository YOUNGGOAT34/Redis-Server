package sentinel

import "sync"

/*
   RaftLog manages the log of entries
	it contains:
		an array of log entries -->where index 0 is a dummy entry, real logging starts at index 1
		highest committed entry index
      highest applied entry index
*/



type RaftLog struct{
	  logMutex sync.RWMutex
	  entries []LogEntry
	  commitIndex int
	  lastAppliedIndex int
}

//NewRaftLog method creates a new log with a dummy entry at index 0

func NewRaftLog() *RaftLog{
	  return &RaftLog{
		     entries: []LogEntry{
				   {Term: 0,index: 0,Command: ""},
			  },
	  }
}


func (log *RaftLog) GetLastIndex() int{
	    log.logMutex.RLock()
		 defer log.logMutex.RUnlock()

		 return len(log.entries)-1
}

func (log *RaftLog) GetLastTerm() Term{

	  log.logMutex.RLock()
	  defer log.logMutex.RUnlock()

     return log.entries[len(log.entries)-1].Term
}

func (log *RaftLog) Append(entries ...LogEntry){
	  log.logMutex.Lock()
	  defer log.logMutex.Unlock()
     
	  log.entries = append(log.entries, entries...)
}

func (log *RaftLog) GetEntry(index int) (LogEntry,bool){
	  log.logMutex.RLock()
	  defer log.logMutex.RUnlock()

	  if index<0 || index>=len(log.entries){
		 return LogEntry{},false
	  }

	  return log.entries[index],true
}

func (log *RaftLog) GetFrom(index int) []LogEntry{

	   log.logMutex.RLock()
		defer log.logMutex.RUnlock()
     
		if index>=len(log.entries){
			 return nil
		}

		logs:=make([]LogEntry,len(log.entries)-index)
		copy(logs,log.entries[index:])

      return logs
}

func (log *RaftLog) TruncateFrom(index int){
	  log.logMutex.Lock()
	  defer log.logMutex.Unlock()

	  if index<len(log.entries){
		   log.entries=log.entries[:index]
	  }
}


func (log *RaftLog) GetCommitIndex() int{
	  log.logMutex.RLock()
	  defer log.logMutex.RUnlock()

	  return log.commitIndex
}

func (log *RaftLog) SetCommitIndex(index int){
	 log.logMutex.Lock()
	 defer log.logMutex.RUnlock()

	 if index>log.commitIndex{
		  log.commitIndex=index
	 }
}

func (log *RaftLog) GetLastAppliedIndex() int{
	  log.logMutex.RLock()
	  defer log.logMutex.RUnlock()

	  return log.lastAppliedIndex
}

func (log *RaftLog) SetLastAppliedIndex(index int){
	   log.logMutex.Lock()
		defer log.logMutex.RUnlock()

		if index>log.lastAppliedIndex{
			log.lastAppliedIndex=index
		}
}

func (log *RaftLog) IsUpToDate(lastTerm Term,lastIndex int) bool{
	  log.logMutex.RLock()
	  defer log.logMutex.RUnlock()
     
	  myLastTerm:=log.entries[len(log.entries)-1].Term
	  myLastIndex:=len(log.entries)-1

	  if  myLastTerm!=lastTerm{
		     return lastTerm> myLastTerm
	  }

	  return lastIndex>=myLastIndex
}

