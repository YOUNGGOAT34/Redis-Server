package zset

import 	"math/rand"

//skip list methods
const MaxLevel=32

func RandomLevel() int{
	  level:=1

	  for rand.Float64() <0.25 && level<MaxLevel{
		 level++
	  }

	  return level
}

//constructor

func NewSkipList() *SkipList{
	  head:=&SkipNode{
		    Forward: make([]*SkipNode,MaxLevel),
			 Span: make([]int,MaxLevel),
	  }

	  return &SkipList{
		    Head: head,
			 Level: 1,
	  }
}


//search 
func (sl *SkipList) Search(node *SkipNode) (*SkipNode,int){
	    
	   _,rank,target:=sl.searchPath(node)
		return target,rank[0]
}

//insert
func (sl *SkipList) Insert(node *SkipNode){
	    
	    update,rank,_:=sl.searchPath(node)
		 //choose height
		 level:=RandomLevel()
		 
      
		 if level>sl.Level{
			  for i:=sl.Level;i<level;i++{
				  
				  update[i]=sl.Head
				  
				  sl.Head.Span[i]=sl.Length
			  }

			  sl.Level=level
		 }

		 //create the new node
		 node.Forward=make([]*SkipNode,level)
		 node.Span=make([]int,level)
		 //reconnect the nodes
		 for i:=0;i<level;i++{
			  oldSpan:=update[i].Span[i]
			  node.Span[i]=oldSpan-(rank[0]-rank[i])
			  update[i].Span[i]=(rank[0]-rank[i])+1
			  node.Forward[i]=update[i].Forward[i]
			  update[i].Forward[i]=node
		 }

		 //update levels above the new node
		 for i:=level;i<sl.Level;i++{
			   update[i].Span[i]++
		 }

		 sl.Length++
}

//update :for deletion and insertion

func (sl *SkipList) searchPath(node *SkipNode) ([]*SkipNode,[]int,*SkipNode){
	     //This will store the predecessor at each level
	    update:=make([]*SkipNode,MaxLevel)
       rank:=make([]int,MaxLevel)
		 current:=sl.Head
		 //search the insertion/deletion position and mark the predecessors
		 for i:=sl.Level-1;i>=0;i--{
              
			   //if it is the top most level ,the rank is 0,otherwise inherit the upper rank

				if i==sl.Level-1{
					 rank[i]=0
				}else{
					 rank[i]=rank[i+1]
				}

			   //isLess function will compare two nodes in terms of both the score and lexicographically
			    for current.Forward[i]!=nil && isLess(current.Forward[i],node){
					   rank[i]+=current.Span[i]
					   current=current.Forward[i]
				 }

				 update[i]=current
		 }

		 return update,rank,current.Forward[0]
}

//Delete

func (sl *SkipList) Delete(node *SkipNode) bool{
	 update,_,target:=sl.searchPath(node)
    
	 if target==nil || target.Score!=node.Score || target.Member!=node.Member{
		 return false
	 }

	 for i:=sl.Level-1;i>=0;i--{
		   if update[i].Forward[i]!=target{
				 update[i].Span[i]--
				 continue
			}
			update[i].Span[i]+=target.Span[i]-1
		   update[i].Forward[i]=target.Forward[i]
	 }

	 for sl.Level>1 && sl.Head.Forward[sl.Level-1]==nil{
		   sl.Level--
	 }

	 sl.Length--
	
	 return true
}

//comparison
func isLess(node1 *SkipNode,node2 *SkipNode) bool{

	if node1.Score<node2.Score{
		   return true
	}
	  
	if node1.Score>node2.Score{
		 return false
	}

	return node1.Member<node2.Member
}
