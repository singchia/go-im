package main

import (
	"github.com/singchia/go-hammer/doublinker"
)

type users struct {
	us map[string]user
	cs map[doublinker.DoubID]user
}

type user struct {
	chID     doublinker.DoubID
	uid      string
	passward string
	online   bool
	sid      string //session object id
	stype    int    //session object type
}
