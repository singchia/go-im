package main

import (
	"sync"

	"github.com/singchia/go-hammer/doublinker"
)

var singleUS *users
var mutexUS sync.Mutex

type users struct {
	us    map[string]*user //uid and passward
	mutex *sync.RWMutex
}

type user struct {
	passward string
	groups   []string
	gs       map[string]*groupState //gid and state //TODO bug here
	curGid   string
}

type groupState struct {
	state  int
	srcUid string
}

func getUsers() *users {
	if singleUS == nil {
		mutexUS.Lock()
		if singleUS == nil {
			singleUS = &users{us: make(map[string]*user), mutex: new(sync.RWMutex)}
		}
		mutexUS.Unlock()
	}
	return singleUS
}

func (u *users) changeGroupState(gid string, dst string, src string, state int) {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	ur := u.us[dst]
	if ur.gs == nil {
		ur.gs = make(map[string]*groupState)
	}
	gs := &groupState{state: state, srcUid: src}
	ur.gs[gid] = gs
	ur.curGid = gid
}

func (u *users) restoreGroupState(uid string) (gid string, srcUid string, state int) {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	gid = u.us[uid].curGid
	return gid, u.us[uid].gs[gid].srcUid, u.us[uid].gs[gid].state
}

func (u *users) deleteGroupState(uid string) {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	delete(u.us[uid].gs, u.us[uid].curGid)
}

func (u *users) isExists(uid string) bool {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	_, ok := u.us[uid]
	return ok
}

func (u *users) addUser(uid string, passward string) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	user := &user{passward: passward}
	u.us[uid] = user
}

func (u *users) appendGroup(uid string, gid string) {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	u.us[uid].groups = append(u.us[uid].groups, gid)
	return
}

func (u *users) lookupPwd(uid string) string {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	return u.us[uid].passward
}

var singleUSI *userStatesIndex
var mutexUSI sync.Mutex

type userStatesIndex struct {
	uids  map[string]*userStates            //uid and userStates
	chids map[doublinker.DoubID]*userStates //chid and userStates
	mutex *sync.RWMutex
}

func (u *userStatesIndex) delete(chid doublinker.DoubID) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	us, ok := u.chids[chid]
	if !ok {
		return
	}
	delete(u.chids, us.chid)
	delete(u.uids, us.uid)
}

func (u *userStatesIndex) addIndex(chid doublinker.DoubID, uid string, states *userStates) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	u.chids[chid] = states
	u.uids[uid] = states
}

func (u *userStatesIndex) lookupUid(chid doublinker.DoubID) string {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	us, ok := u.chids[chid]
	if ok {
		return us.uid
	}
	return NULL
}

func (u *userStatesIndex) lookupChid(uid string) doublinker.DoubID {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	us, ok := u.uids[uid]
	if ok {
		return us.chid
	}
	return nil
}

func getUserStatesIndex() *userStatesIndex {
	if singleUSI == nil {
		mutexUSI.Lock()
		if singleUSI == nil {
			singleUSI = &userStatesIndex{uids: make(map[string]*userStates), chids: make(map[doublinker.DoubID]*userStates), mutex: new(sync.RWMutex)}
		}
		mutexUSI.Unlock()
	}
	return singleUSI
}

type userStates struct {
	chid doublinker.DoubID
	uid  string
}
