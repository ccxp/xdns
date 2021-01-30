package main

import (
	"container/list"
)

type stringEle struct {
	k    string
	v    interface{}
	size int
}

type stringLru struct {
	M     map[string]*list.Element
	l     *list.List
	total int
}

// 添加一个k/v，并放在最前面，注意要自行保证原本无这个key
func (lru *stringLru) Push(k string, v interface{}, size int) {
	if lru.l == nil {
		lru.l = list.New()
		lru.M = make(map[string]*list.Element)
	}
	ele := lru.l.PushFront(stringEle{k: k, v: v, size: size})
	lru.M[k] = ele
	lru.total += size
}

// 读取一个k，如果存在，则移到前面
func (lru *stringLru) Get(k string) (interface{}, *list.Element) {
	ele := lru.M[k]
	if ele != nil {
		lru.l.MoveToFront(ele)
		return ele.Value.(stringEle).v, ele
	}
	return nil, nil
}

// 从队列移出最后一个数据
func (lru *stringLru) RemoveBack() {
	ele := lru.l.Back()
	if ele != nil {
		dat := ele.Value.(stringEle)
		delete(lru.M, dat.k)
		lru.l.Remove(ele)
		lru.total -= dat.size
	}
}

func (lru *stringLru) Remove(ele *list.Element) {
	dat := ele.Value.(stringEle)
	delete(lru.M, dat.k)
	lru.l.Remove(ele)
	lru.total -= dat.size
}

func (lru *stringLru) Count() int {
	return len(lru.M)
}

func (lru *stringLru) Size() int {
	return lru.total
}
