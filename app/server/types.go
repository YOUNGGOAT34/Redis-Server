package server

/*
   
*/


type TYPE int

const(
	  STRING TYPE=iota
	  LIST
)

type Data struct{

	    Type TYPE
		 Value any
}




type ResponseType int

const (
	  ERROR ResponseType=iota
	  SIMPLE_STRING
	  BULK_STRING
	  NIL
	  INTEGER
	  ARRAY
)

type Response struct{
	   Body []byte
		Type ResponseType
}


type Node struct{
	   data []byte
		Prev *Node
		Next *Node
}


type List struct{
	    Head *Node
		 Tail *Node
		 len int
}


func (list *List) PushFront(value []byte){
	     
	     node:=&Node{
								  data:value,  
							}

							if list.Head!=nil{

								tmp:=list.Head
								node.Next=tmp
								tmp.Prev=node
								list.Head=node
							}else{
								  list.Head=node
								  list.Tail=node
				}
			list.len++
			
}




func (list *List) PushBack(value []byte){
	   node:=&Node{
			   data:value,
		}

		if list.Head!=nil{
			  tmp:=list.Tail
			  tmp.Next=node
			  list.Tail=node
			  node.Prev=tmp

		}else{
			  list.Head=node
			  list.Tail=node
		}

		list.len++
}


func (list *List) LPop() []byte{
  
	  if list==nil || list.len==0{
		    return nil
	  }

	  tmp:=list.Head
	  list.Head=tmp.Next
	  if list.Head!=nil{
		  list.Head.Prev=nil 
	  }else{
		   list.Tail=nil
	  }
	  list.len--

	  return tmp.data
}
