package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/singchia/go-hammer/doublinker"
)

const (
	UNHANDLEDJOIN = iota
	UNHANDLEDINVITE
	AGREED
	REJECTED
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

func (g *groups) appendUser(gid string, uid string) []string {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	gp, _ := g.gs[gid]
	gp.members = append(gp.members, uid)
	return gp.members
}

func (g *groups) addGroup(gid string, uid string) *group {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	_, ok := g.gs[gid]
	if ok {
		return nil
	}
	g.gs[gid] = &group{owner: uid}
	return g.gs[gid]
}

type group struct {
	owner   string
	members []string //seperate the persistent data and the runtime states
}

var singleGSI *groupStatesIndex
var mutexGSI sync.Mutex

type groupStatesIndex struct {
}

func getGroupStatesIndex() *groupStatesIndex {
	if singleGSI == nil {
		mutexGSI.Lock()
		if singleGSI == nil {
			singleGSI = &groupStatesIndex{}
		}
		mutexGSI.Unlock()
	}
	return singleGSI
}

func (g *groupStatesIndex) handle(chid doublinker.DoubID, cmd, suffix string) {
	uid := getUserStatesIndex().lookupUid(chid)
	if cmd == CREATEGROUP {
		if !validate(suffix) {
			getQueue().pushDown(&message{chid: chid, data: "[from system] invalid group name.\n"})
			return
		}
		g := getGroups().addGroup(suffix, uid)
		if g == nil {
			getQueue().pushDown(&message{chid: chid, data: "[from system] group already exists.\n"})
			return
		}
		getUsersIndex().appendGroup(uid, suffix)
		getGroups().appendUser(suffix, uid)
		getQueue().pushDown(&message{chid: chid, data: "[from system] group create succeed.\n"})
		return

	} else if cmd == JOINGROUP {
		if !validate(suffix) {
			getQueue().pushDown(&message{chid: chid, data: "[from system] invalid group name.\n"})
			return
		}
		g := getGroups().getGroup(suffix)
		if g == nil {
			getQueue().pushDown(&message{chid: chid, data: "[from system] group does not exist.\n"})
			return
		}
		owner := getGroups().getGroup(suffix).owner
		ownerChid := getUserStatesIndex().lookupChid(owner)
		if ownerChid == nil {
			getQueue().pushDown(&message{chid: chid, data: "[from system] group owner offline.\n"})
			return
		}
		getQueue().pushDown(&message{chid: ownerChid, data: fmt.Sprintf("[from %s] apply for joining group %s, Y/N/I? ", uid, suffix)})
		getSessionStatesIndex().changeSession(ownerChid, OPERATING, true)
		getUsersIndex().changeGroupState(suffix, owner, uid, UNHANDLEDJOIN)

	} else if cmd == INVITEGROUP {
		elems := strings.Split(suffix, " ")
		if len(elems) != 2 || !validate(elems[0]) || !validate(elems[1]) {
			getQueue().pushDown(&message{chid: chid, data: "[from system] invalid field.\n"})
			return
		}
		g := getGroups().getGroup(elems[0])
		if g == nil || g.owner != uid {
			getQueue().pushDown(&message{chid: chid, data: "[from system] group does not exist or not group owner.\n"})
			return
		}
		if g.owner == elems[1] {
			getQueue().pushDown(&message{chid: chid, data: "[from system] already in group.\n"})
			return
		}
		peerChid := getUserStatesIndex().lookupChid(elems[1])
		if peerChid == nil {
			getQueue().pushDown(&message{chid: chid, data: "[from system] object offline.\n"})
			return
		}
		getQueue().pushDown(&message{chid: peerChid, data: fmt.Sprintf("[from %s] invite you to join group %s, Y/N/I? ", uid, elems[0])})
		getSessionStatesIndex().changeSession(peerChid, OPERATING, true)
		getUsersIndex().changeGroupState(elems[0], elems[1], uid, UNHANDLEDINVITE)

	} else if cmd == RESTORENOTES {
		if !validate(suffix) {
			getQueue().pushDown(&message{chid: chid, data: "[from system] invalid group name.\n"})
			getSessionStatesIndex().restoreSession(chid)
			return
		}
		gs := getUsersIndex().getGroupState(uid, suffix)
		if gs == nil {
			getQueue().pushDown(&message{chid: chid, data: "[from system] no note with this group.\n"})
			getSessionStatesIndex().restoreSession(chid)
			return
		}
		if gs.state == UNHANDLEDJOIN {
			getQueue().pushDown(&message{chid: chid, data: fmt.Sprintf("[from %s] apply for joining group %s, Y/N/I? ", gs.srcUid, suffix)})
			getSessionStatesIndex().changeSession(chid, OPERATING, true)
			getUsersIndex().changeCurGid(uid, suffix)
			return
		} else if gs.state == UNHANDLEDINVITE {
			getQueue().pushDown(&message{chid: chid, data: fmt.Sprintf("[from %s] invite you to join group %s, Y/N/I? ", gs.srcUid, suffix)})
			getSessionStatesIndex().changeSession(chid, OPERATING, true)
			getUsersIndex().changeCurGid(uid, suffix)
			return
		}
		return

	} else {
		g.interactive(chid, uid, suffix)
	}
}

func (g *groupStatesIndex) interactive(chid doublinker.DoubID, uid, suffix string) {
	gid, srcUid, state := getUsersIndex().restoreGroupState(uid)
	if suffix != "Y" && suffix != "N" && suffix != "I" {
		getQueue().pushDown(&message{chid: chid, data: fmt.Sprintf("[from %s] Y/N/I? ", srcUid)})
		return
	}
	if suffix == "Y" {
		if state == UNHANDLEDJOIN {
			members := getGroups().appendUser(gid, srcUid)
			getUsersIndex().appendGroup(srcUid, gid)
			getUsersIndex().deleteGroupState(uid)
			getSessionStatesIndex().restoreSession(chid)
			for _, member := range members {
				peerChid := getUserStatesIndex().lookupChid(member)
				getQueue().pushDown(&message{chid: peerChid, data: fmt.Sprintf("[from system] %s join in group %s.\n", srcUid, gid)})
			}
			return
		} else if state == UNHANDLEDINVITE {
			members := getGroups().appendUser(gid, uid)
			getUsersIndex().appendGroup(uid, gid)
			getUsersIndex().deleteGroupState(uid)
			getSessionStatesIndex().restoreSession(chid)
			for _, member := range members {
				peerChid := getUserStatesIndex().lookupChid(member)
				getQueue().pushDown(&message{chid: peerChid, data: fmt.Sprintf("[from system] %s join in group %s.\n", uid, gid)})
			}
			return
		}
		return
	} else if suffix == "N" {
		if state == UNHANDLEDJOIN || state == UNHANDLEDINVITE {
			getUsersIndex().deleteGroupState(uid)
			getSessionStatesIndex().restoreSession(chid)
			peerChid := getUserStatesIndex().lookupChid(srcUid)
			getQueue().pushDown(&message{chid: peerChid, data: fmt.Sprintf("[from system] join in group %s rejected.\n", gid)})
		}
	} else if suffix == "I" {
		getSessionStatesIndex().restoreSession(chid)
	}

}
