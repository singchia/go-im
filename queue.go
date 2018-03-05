package main

import (
	"sync"

	"github.com/singchia/go-hammer/doublinker"
)

type message struct {
	messageType int
	chid        doublinker.DoubID
	data        string
}

type queue struct {
	upChannel   chan *message
	downChannel chan *message
	done        chan bool
}

func (q *queue) pushUp(m *message) {
	q.upChannel <- m
}

func (q *queue) pullUp() <-chan *message {
	return q.upChannel
}

func (q *queue) pushDown(m *message) {
	q.downChannel <- m
}

func (q *queue) pullDown() <-chan *message {
	return q.downChannel
}

var singleQ *queue
var mutexQ sync.Mutex

func getQueueInstance() *queue {
	if singleQ == nil {
		mutexQ.Lock()
		if singleQ == nil {
			singleQ = &queue{upChannel: make(chan *message, 1024*1024), downChannel: make(chan *message, 1024*1024), done: make(chan bool, 1)}
		}
		mutexQ.Unlock()
	}
	return singleQ
}
