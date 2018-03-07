package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/singchia/go-hammer/doublinker"
)

const (
	UNHANDLEDJOIN = iota
	IGNOREDJOIN
	UNHANDLEDINVITE
	IGNOREDINVITE
)

var singleGS *groups
var mutexGS sync.Mutex

type groups struct {
	gs    map[string]*group
	mutex *sync.RWMutex
}

func getGroups() *groups {
	if singleGS == nil {
		mutexGS.Lock()
		if singleGS == nil {
			singleGS = &groups{gs: make(map[string]*group), mutex: new(sync.RWMutex)}
		}
		mutexGS.Unlock()
	}
	return singleGS
}

func (g *groups) getGroup(gid string) *group {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	gp, _ := g.gs[gid]
	return gp
}

func (g *groups) addGroup(gid string, uid string) *group {
	g.mutex.Lock()
	defer g.mutex.UnLock()
	_, ok := g.gs[gid]
	if ok {
		return nil
	}
	g.gs[gid] = &group{gid: gid, owner: uid}
	return g.gs[gid]
}

type group struct {
	owner   string
	members []string //seperate the persistent data and the runtime states
}

var singleGSI *groupStatesIndex
var mutexGSI sync.Mutex

type groupStatesIndex struct {
	states map[string]*groupStates //gid and userStates
	mutex  *sync.RWMutex
}

func getGroupStatesIndex() *groupStatesIndex {
	if singleGSI == nil {
		mutexGSI.Lock()
		if singleGSI == nil {
			singleGSI = &groupStatesIndex{states: make(map[string]*groupStates), mutex: new(sync.RWMutex)}
		}
		mutexGSI.Unlock()
	}
	return singleGSI
}

func (g *getGroupStatesIndex) handle(chid doublinker.DoubID, cmd, suffix string) {
	uid := getUserStatesIndex().lookupUid(chid)
	if cmd == CREATEGROUP {
		if !validate(suffix) {
			getQueue().pushDown(&message{chid: chid, data: "[from system] invalid group name.\n"})
			return
		}
		g := getGroups().addGroup(suffix)
		if g == nil {
			getQueue().pushDown(&message{chid: chid, data: "[from system] group already exists.\n"})
			return
		}
		getUsersIndex().appendGroup(uid, suffix)
		getQueue().pushDown(&message{chid: chid, data: "[from system] group create succeed.\n"})
		return

	} else if cmd == JOINGROUP {
		if !validate(suffix) {
			getQueue().pushDown(&message{chid: chid, data: "[from system] invalid group name.\n"})
			return
		}
		owner := getGroups().getGroup(suffix).owner
		ownerChid := getUserStatesIndex().lookupChid(owner)
		if ownerChid == nil {
			getQueue().pushDown(&message{chid: chid, data: "[from system] group owner offline.\n"})
			return
		}
		getQueue().pushDown(&message{chid: ownerChid, data: fmt.Sprintf("[from %s] apply for joining group %s, Y/N/I?.\n", uid, suffix)})
		getSessionStatesIndex().changeSession(ownerChid, OPERATING, true)
		g.changeState(suffix, suffix, UNHANDLEDINVITE)

	} else if cmd == INVITEGROUP {
		elems := strings.Split(suffix)
		if len(elems) != 2 || !validate(elems[0]) || !validate(elems[1]) {
			getQueue().pushDown(&message{chid: chid, data: "[from system] invalid field.\n"})
			return
		}
		g := getGroups().getGroup(suffix[0])
		if g == nil || g.owner != uid {
			getQueue().pushDown(&message{chid: chid, data: "[from system] group does not exist or not group owner.\n"})
			return
		}
		if owner == suffix[1] {
			getQueue().pushDown(&message{chid: chid, data: "[from system] already in group.\n"})
			return
		}
		peerChid := getUserStatesIndex().lookupChid(suffix[1])
		if peerChid == nil {
			getQueue().pushDown(&message{chid: chid, data: "[from system] object offline.\n"})
			return
		}
		getQueue().pushDown(&message{chid: peerChid, data: fmt.Sprintf("[from %s] invite you to join group %s, Y/N/I?.\n", uid, suffix[0])})
		getSessionStatesIndex().changeSession(peerChid, OPERATING, true)
		g.changeState(suffix, suffix[1], UNHANDLEDINVITE)

	} else {
		if suffix != "Y" && suffix != "N" && suffix != "I" {
			getQueue().pushDown(&message{chid: peerChid, data: fmt.Sprintf("[from system] re-enter Y/N/I?.\n", uid, suffix[0])})
			return
		}
		if suffix == Y {

		}
	}
}

func (g *groupStatesIndex) changeState(gid string, uid string, state int) {
	g.mutex.RLock()
	ss, ok := g.states[gid]
	g.mutex.RUnlock()
	if !ok {
		g.mutex.Lock()
		gs := &groupStates{state: make(map[string]int)}
		gs.state[uid] = state
		g.states[gid] = gs
		g.mutex.Unlock()
	}
	ss.states[uid] = state
}

type groupStates struct {
	state map[string]int
}
