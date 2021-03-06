package main

import (
	"sync"

	"github.com/singchia/go-hammer/doublinker"
)

var singleASI *authStatesIndex
var mutexASI sync.Mutex

type authStatesIndex struct {
	as    map[doublinker.DoubID]*authStates //chid and authStates
	mutex *sync.RWMutex
}

func getAuthStatesIndex() *authStatesIndex {
	if singleASI == nil {
		mutexASI.Lock()
		if singleASI == nil {
			singleASI = &authStatesIndex{as: make(map[doublinker.DoubID]*authStates), mutex: new(sync.RWMutex)}
		}
		mutexASI.Unlock()
	}
	return singleASI
}

func (a *authStatesIndex) delete(chid doublinker.DoubID) {
	a.mutex.Lock()
	delete(a.as, chid)
	a.mutex.Unlock()
}

func (a *authStatesIndex) handle(chid doublinker.DoubID, cmd string, suffix string) {
	if cmd == SIGNUP || cmd == SIGNIN { //cover the existing states
		if !validate(suffix) {
			getQueue().pushDown(&message{chid: chid, data: "[from system] invalid user name.\n"})
			getSessionStatesIndex().restoreSession(chid)
			return
		}
		if getUsers().isExists(suffix) && cmd == SIGNUP {
			getQueue().pushDown(&message{chid: chid, data: "[from system] user already exists.\n"})
			getSessionStatesIndex().restoreSession(chid)
			return
		} else if !getUsers().isExists(suffix) && cmd == SIGNIN {
			getQueue().pushDown(&message{chid: chid, data: "[from system] user does not exist.\n"})
			getSessionStatesIndex().restoreSession(chid)
			return
		}
		a.mutex.Lock()
		var as *authStates
		if cmd == SIGNUP {
			as = &authStates{uid: suffix, flag: 1}
		} else {
			as = &authStates{uid: suffix, flag: 3}
		}
		a.as[chid] = as
		a.mutex.Unlock()
		getQueue().pushDown(&message{chid: chid, data: "[from system] enter passward:"})
		return
	}

	a.mutex.RLock()
	defer a.mutex.RUnlock()
	if as, ok := a.as[chid]; ok {
		if as.flag == 1 {
			as.flag = 2
			as.passward = suffix
			getQueue().pushDown(&message{chid: chid, data: "[from system] re-enter passward:"})
			return
		} else if as.flag == 2 {
			if as.passward == suffix {
				getUsers().addUser(as.uid, as.passward)
				a.authSucceed(chid, as.uid)
				return
			}
			getQueue().pushDown(&message{chid: chid, data: "[from system] re-enter passward:"})
			return
		} else if as.flag == 3 {
			if getUsers().lookupPwd(as.uid) == suffix {
				a.authSucceed(chid, as.uid)
				return
			}
			getQueue().pushDown(&message{chid: chid, data: "[from system] re-enter passward:"})
		}
	}
}

func (a *authStatesIndex) authSucceed(chid doublinker.DoubID, uid string) {
	delete(a.as, chid) //clear the authStates
	getQueue().pushDown(&message{chid: chid, data: "[from system] auth succeed.\n"})
	us := &userStates{chid: chid, uid: uid}
	getUserStatesIndex().addIndex(chid, uid, us)
	getSessionStatesIndex().changeSession(chid, AUTHORIZED, true)
}

func (a *authStatesIndex) clear(chid doublinker.DoubID) {

}

type authStates struct {
	uid      string
	passward string
	flag     int //1: signup  2: first-passward  3: singin
}
