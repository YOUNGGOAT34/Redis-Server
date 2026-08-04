package zset

import "fmt"

//ZADD
func (zs *ZSet) Add(node *SkipNode) bool{
      deleted:=false
	   if existing,exists:=zs.Dict[node.Member];exists{
			  if existing.Score==node.Score{
				   return false
			  }
			  deleted=zs.List.Delete(existing)
		}
	   
	   zs.List.Insert(node)
		zs.Dict[node.Member]=node
		return deleted
}
//ZScore

func (zs *ZSet) ZScore(member string) *SkipNode{
	   if node,exists:=zs.Dict[member];exists{
			   return node
		}

		return nil
}

func (zs *ZSet) ZRem(member string) bool{
	  node,exists:=zs.Dict[member]
      
	  deleted:=false
     
	  if exists{
		  deleted=zs.List.Delete(node)
		  delete(zs.Dict,member)
		  
	  }
	  return deleted
}


func (sl *SkipList) Print() {
    for level := sl.Level - 1; level >= 0; level-- {
        fmt.Printf("Level %d: HEAD", level)

        current := sl.Head
        for current.Forward[level] != nil {
            fmt.Printf(" --(%d)--> %s",
                current.Span[level],
                current.Forward[level].Member,
            )
            current = current.Forward[level]
        }

        fmt.Println()
    }
}