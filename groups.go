package main

type groups struct {
	gs map[string]group
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
