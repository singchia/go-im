package main

type groups struct {
	gs map[string]group
}

type group struct {
	gid     string
	owner   user
	members []user
}
