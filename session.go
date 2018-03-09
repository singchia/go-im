package main

import (
	"sync"

	"github.com/singchia/go-hammer/doublinker"
)

const (
	AUTHORIZING = iota + 10
	AUTHORIZED
	CHATTING
	OPERATING
	INTERACTING
	UNKNOW
)

var singleSSI *sessionStatesIndex
var mutexSSI sync.Mutex

type handler interface {
	handle(chid doublinker.DoubID, cmd string, suffix string)
	delete(chid doublinker.DoubID)
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
			singleSSI.handlers[AUTHORIZING] = getAuthStatesIndex()
			singleSSI.handlers[AUTHORIZED] = singleSSI
			singleSSI.handlers[CHATTING] = getChatStatesIndex()
			singleSSI.handlers[OPERATING] = getGroups()
			singleSSI.handlers[INTERACTING] = getGroups()
			singleSSI.handlers[CLOSED] = singleSSI
		}
		mutexSSI.Unlock()
	}
	return singleSSI
}

func (s *sessionStatesIndex) handle(chid doublinker.DoubID, cmd, suffix string) {
	getQueue().pushDown(&message{mtype: PASSTHROUGH, chid: chid, data: "[from system] unsupported command\n"})
	s.mutex.RLock()
	states, ok := s.ss[chid]
	if !ok && cmd == SIGNOUT {
		getQueue().pushDown(&message{mtype: PASSTHROUGH, chid: chid, data: "[from system] user unauthed yet.\n"})
		s.mutex.RUnlock()
		return
	}
	s.mutex.RUnlock()

	s.mutex.Lock()
	for ret := states.pop(); ret != -1; ret = states.pop() {
		s.handlers[ret].delete(chid)
	}
	s.mutex.Unlock()
	s.delete(chid)
	return
}

func (s *sessionStatesIndex) delete(chid doublinker.DoubID) {
	delete(s.ss, chid)
}

func (s *sessionStatesIndex) dispatch(chid doublinker.DoubID, cmd, suffix string) {
	s.mutex.RLock()
	states, ok := s.ss[chid] //since chid is unique, value is unique too
	s.mutex.RUnlock()

	if !ok {
		if s.mapping(cmd) != AUTHORIZING {
			getQueue().pushDown(&message{chid: chid, data: "using [signup: foo] or [signin: foo]\n"})
			return
		}
		ss := &sessionStates{}
		ss.push(AUTHORIZING)
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
	return
}

func (s *sessionStatesIndex) mapping(cmd string) int {
	if cmd == SIGNUP || cmd == SIGNIN || cmd == SIGNOUT {
		return AUTHORIZING
	}
	if cmd == TOUSER || cmd == TOGROUP {
		return CHATTING
	}
	if cmd == CREATEGROUP || cmd == JOINGROUP || cmd == INVITEGROUP || cmd == RESTORENOTES {
		return OPERATING
	}
	if cmd == CLOSE {
		return CLOSED
	}
	return UNKNOW
}

func (s *sessionStatesIndex) lookupSessionState(chid doublinker.DoubID) int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.ss[chid].top()
}

func (s *sessionStatesIndex) changeSession(chid doublinker.DoubID, state int, over bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	states, ok := s.ss[chid]
	if ok {
		if over {
			states.push(state)
			return
		}
		states.pop()
		states.push(state)
		return
	}
	return
}

func (s *sessionStatesIndex) restoreSession(chid doublinker.DoubID) {
	s.mutex.RLock()
	states, ok := s.ss[chid]
	s.mutex.RUnlock()
	if ok {
		states.pop()
		if states.count == 0 {
			s.mutex.Lock()
			delete(s.ss, chid)
			s.mutex.Unlock()
		}
		return
	}
	return
}

type sessionStates struct {
	states []int
	count  int
}

func (s *sessionStates) top() int {
	if s.count == 0 {
		return -1
	}
	return s.states[s.count-1]
}

func (s *sessionStates) push(n int) {
	s.states = append(s.states[:s.count], n)
	s.count++
}

// Pop removes and returns a node from the stack in last to first order.
func (s *sessionStates) pop() int {
	if s.count == 0 {
		return -1
	}
	s.count--
	return s.states[s.count]
}
