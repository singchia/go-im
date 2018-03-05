package main

import "strings"

const (
	NONCOMMAND = "non command"
	SIGNUP     = "signup:"
	SIGNIN     = "signin:"
	SIGNOUT    = "signout"
	TOUSER     = "to user"
	TOGROUP    = "to group"
)

var cmds = [...]string{SIGNUP, SIGNIN, SIGNOUT}

type parser struct {
	cmdMap map[int][]string //length and command
	minLen int
}

func newParser() *parser {
	cmdMap := make(map[int][]string)
	minLen := 0
	for _, cmd := range cmds {
		if len(cmd) < minLen {
			minLen = len(cmd)
		}

		arr, ok := cmdMap[len(cmd)]
		if ok {
			append(arr, cmd)
			continue
		}
		arr = make([]string)
		append(arr, cmd)
		cmdMap[len(cmd)] = arr
	}
	return &parser{cmdMap: cmdMap, minLen: minLen}
}

func (p *parser) parse() {
	for i := 0; i < 100; i++ {
		go func() {
			for {
				select {
				case message := <-GetQueueInstance().pullUp():
					cmd, suffix := p.split(message.data)
					getSessionStatesIndex().dispatch(message.chid, cmd, suffix)
				}
			}
		}()
	}
}

func (p *parser) split(str string) (string, string) {
	if len(str) < p.minLen {
		return NONCOMMAND, str
	}
	for k, v := range p.cmdMap {
		if len(str) < k {
			continue
		}
		for _, cmd := range v {
			if str[0:k] == cmd {
				return cmd, strings.TrimSpace(str[k+1 : len(str)])
			}
		}
	}
	return NONCOMMAND, str
}
