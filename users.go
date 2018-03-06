package main

import (
	"sync"

	"github.com/singchia/go-hammer/doublinker"
)

var singleUS *users
var mutexUS sync.Mutex

type users struct {
	us    map[string]string //uid and passward
	mutex *sync.RWMutex
}

func getUsersInstance() *users {
	if singleUS == nil {
		mutexUS.Lock()
		if singleUS == nil {
			singleUS = &users{us: make(map[string]string), mutex: new(sync.RWMutex)}
		}
		mutexUS.Unlock()
	}
	return singleUS
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
	u.us[uid] = passward
}

func (u *users) lookupPwd(uid string) string {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	return u.us[uid]
}

var singleUSI *userStatesIndex
var mutexUSI sync.Mutex

type userStatesIndex struct {
	uids  map[string]*userStates            //uid and userStates
	chids map[doublinker.DoubID]*userStates //chid and userStates
	mutex *sync.RWMutex
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
