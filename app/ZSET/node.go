package zset

import "sync"

//skip list

type SkipNode struct{
	Member string
	Score float64
	Forward []*SkipNode
	Span    []int
}

type SkipList struct{
	Head *SkipNode
	Level int
	Length int
}

type ZSet struct{
	Dict map[string]*SkipNode
	List *SkipList
	ZSMutex sync.RWMutex
}
