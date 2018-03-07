package main

import (
	"fmt"
	"sync"

	"github.com/singchia/go-hammer/doublinker"
)

var singleCSI *chatStatesIndex
var mutexCSI sync.Mutex

type chatStatesIndex struct {
	uc    map[string]*chatStates
	mutex *sync.RWMutex
}

func getChatStatesIndex() *chatStatesIndex {
	if singleCSI == nil {
		mutexCSI.Lock()
		if singleCSI == nil {
			singleCSI = &chatStatesIndex{uc: make(map[string]*chatStates), mutex: new(sync.RWMutex)}
		}
		mutexCSI.Unlock()
	}
	return singleCSI
}

func (c *chatStatesIndex) handle(chid doublinker.DoubID, cmd, suffix string) {
	uid := getUserStatesIndex().lookupUid(chid)
	if cmd == TOUSER || cmd == TOGROUP {
		if !validate(suffix) {
			getQueue().pushDown(&message{chid: chid, data: "[from system] invalid object name.\n"})
			return
		}

		c.mutex.Lock()
		c.uc[uid] = &chatStates{object: suffix, otype: cmd}
		c.mutex.Unlock()
		return
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()
	cs, _ := c.uc[uid]

	if cs.otype == TOUSER {
		peerChid := getUserStatesIndex().lookupChid(cs.object)
		if peerChid == nil {
			getQueue().pushDown(&message{chid: chid, data: "[from system] object offline.\n"})
			return
		}
		getQueue().pushDown(&message{chid: peerChid, data: fmt.Sprintf("[from user %s] %s", uid, suffix)})
		return
	}

	group := getGroups().getGroup(cs.object)
	if group == nil {
		getQueue().pushDown(&message{chid: chid, data: "[from system] group does not exist\n"})
		return
	}

	var isMemeber bool = false
	for _, v := range group.members {
		if v == uid {
			isMemeber = true
			break
		}
	}
	if isMemeber == false {
		getQueue().pushDown(&message{chid: chid, data: "[from system] not a member of this group\n"})
		return
	}

	for _, v := range group.members {
		peerChid := getUserStatesIndex().lookupChid(v)
		if peerChid == nil {
			continue
		}
		getQueue().pushDown(&message{chid: peerChid, data: fmt.Sprintf("[from user %s in group %s] %s", uid, cs.object, suffix)})
	}
}

type chatStates struct {
	object string
	otype  string
}
