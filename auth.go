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

func (a *authStatesIndex) validate(uid string) bool {
	if uid == "" || uid[0] < 'a' || uid[0] > 'z' {
		return false
	}
	for _, r := range uid[1:len(uid)] {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '1' || r > '9') {
			return false
		}
	}
	return true
}

func (a *authStatesIndex) handle(chid doublinker.DoubID, cmd string, suffix string) {
	if cmd == SIGNUP || cmd == SIGNIN { //cover the existing states
		if !a.validate(suffix) {
			getQueueInstance().pushDown(&message{chid: chid, data: "invalid username\n"})
			return
		}
		if getUsersInstance().isExists(suffix) && cmd == SIGNUP {
			getQueueInstance().pushDown(&message{chid: chid, data: "user already exists\n"})
			return
		} else if !getUsersInstance().isExists(suffix) && cmd == SIGNIN {
			getQueueInstance().pushDown(&message{chid: chid, data: "user does not exist\n"})
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
		getQueueInstance().pushDown(&message{chid: chid, data: "enter passward\n"})
		return
	} else if cmd == SIGNOUT {

	}

	a.mutex.Lock()
	defer a.mutex.Unlock()
	if as, ok := a.as[chid]; ok {
		if as.flag == 1 {
			as.flag = 2
			as.passward = suffix
			getQueueInstance().pushDown(&message{chid: chid, data: "re-enter passward\n"})
			return
		} else if as.flag == 2 {
			if as.passward == suffix {
				getUsersInstance().addUser(as.uid, as.passward)
				a.authSucceed(chid, as.uid)
				return
			}
			getQueueInstance().pushDown(&message{chid: chid, data: "re-enter passward\n"})
			return
		} else if as.flag == 3 {
			if getUsersInstance().lookupPwd(as.uid) == suffix {
				a.authSucceed(chid, as.uid)
				return
			}
			getQueueInstance().pushDown(&message{chid: chid, data: "re-enter passward\n"})
		}
	}
}

func (a *authStatesIndex) authSucceed(chid doublinker.DoubID, uid string) {
	delete(a.as, chid) //clear the authStates
	getQueueInstance().pushDown(&message{chid: chid, data: "auth succeed\n"})
	us := &userStates{chid: chid, uid: uid, online: true}
	getUserStatesIndex().addIndex(chid, uid, us)
	getSessionStatesIndex().changeSession(chid, LOGGED, true)
}

func (a *authStatesIndex) clear(chid doublinker.DoubID) {

}

type authStates struct {
	uid      string
	passward string
	flag     int //1: signup  2: first-passward  3: singin
}
