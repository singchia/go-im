package main

import (
	"sync"

	"github.com/singchia/go-hammer/doublinker"
)

const (
	CLOSED = iota
	DATA
)

type message struct {
	messageType int
	chID        doublinker.DoubID
	data        []byte
}

type queue struct {
	channel chan *message
	done    chan bool
}

func (q *queue) Push(m *message) {
	q.channel <- m
}

func (q *queue) Pull() *message {
	m := <-q.channel
	return m
}

//sigleton
var single *queue
var mutex sync.Mutex

func GetQueueInstance() *queue {
	if single == nil {
		mutex.Lock()
		if single == nil {
			single = &queue{channel: make(chan *message, 1024*1024), done: make(chan bool, 1)}
		}
		mutex.Unlock()
	}
	return single
}
