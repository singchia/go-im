package main

import (
	"sync"

	"github.com/singchia/go-hammer/doublinker"
)

const (
	UNPARSED = iota
	AUTH
	LOGGED
	CHATTING
	NOFICATION
	DATA
	CLOSED
)

var singleSSI *sessionStatesIndex
var mutexSSI sync.Mutex

type handler interface {
	handle(chid doublinker.DoubID, cmd string, suffix string)
}

type sessionStatesIndex struct {
	ss       map[doublinker.DoubID]*sessionStates //chid and sessionStates
	mutex    *sync.RWMutex
	handlers map[int]handler
}

func getSessionStatesIndex() *sessionStatesIndex {
	if singleSSI == nil {
		mutexSSI.Lock()
		if singleSSI == nil {
			singleSSI = &sessionStatesIndex{ss: make(map[doublinker.DoubID]*sessionStates), mutex: new(sync.RWMutex), handlers: make(map[int]handler)}
			singleSSI.handlers[AUTH] = getAuthStatesIndex()
			singleSSI.handlers[LOGGED] = singleSSI
		}
		mutexSSI.Unlock()
	}
	return singleSSI
}

func (s *sessionStatesIndex) handle(chid doublinker.DoubID, cmd, suffix string) {
	getQueueInstance().pushDown(&message{messageType: DATA, chid: chid, data: "unsupported command"})
}

func (s *sessionStatesIndex) dispatch(chid doublinker.DoubID, cmd, suffix string) {
	s.mutex.RLock()
	states, ok := s.ss[chid] //since chid is unique, value is unique too
	s.mutex.RUnlock()

	if !ok {
		if s.mapping(cmd) != AUTH {
			getQueueInstance().pushDown(&message{chid: chid, data: "using [signup: foo] or [signin: foo]"})
			return
		}
		ss := &sessionStates{states: make([]int, 0, 1), count: 0}
		ss.push(AUTH)
		s.mutex.Lock()
		s.ss[chid] = ss
		s.mutex.Unlock()
		getAuthStatesIndex().handle(chid, cmd, suffix)
		return
	}
	if cmd == NONCOMMAND {
		s.handlers[states.top()].handle(chid, cmd, suffix)
		return
	}
	s.changeSession(chid, s.mapping(cmd), true)
	s.handlers[s.mapping(cmd)].handle(chid, cmd, suffix)
}

func (s *sessionStatesIndex) mapping(cmd string) int {
	if cmd == SIGNUP || cmd == SIGNIN || cmd == SIGNOUT {
		return AUTH
	}
	if cmd == TOUSER || cmd == TOGROUP {
		return CHATTING
	}
	return DATA
}

func (s *sessionStatesIndex) changeSession(chid doublinker.DoubID, state int, cover bool) {
	s.mutex.RLock()
	states, ok := s.ss[chid]
	if ok {
		if cover {
			states.push(state)
			return
		}
		states.pop()
		states.push(state)
		return
	}
	defer s.mutex.RUnlock()
}

//an easy stack
type sessionStates struct {
	states []int
	count  int
}

func (s *sessionStates) pop() int {
	if s.count == 0 {
		return -1
	}
	s.count--
	return s.states[s.count]
}

func (s *sessionStates) top() int {
	if s.count == 0 {
		return -1
	}
	return s.states[s.count]
}

func (s *sessionStates) push(e int) {
	s.states = append(s.states[:s.count], e)
	s.count++
}
