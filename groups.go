package main

import "sync"

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

type group struct {
	gid     string
	owner   string
	members []string //seperate the persistent data and the runtime states
}

type groupStatesIndex struct {
	states map[string]groupStates //gid and userStates
}

type groupStates struct {
	//TODO
}
