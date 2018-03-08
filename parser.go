package main

import "strings"

const (
	NONCOMMAND   = "non-command"
	SIGNUP       = "signup:"
	SIGNIN       = "signin:"
	SIGNOUT      = "signout"
	TOUSER       = "to user:"
	TOGROUP      = "to group:"
	CREATEGROUP  = "create group:"
	JOINGROUP    = "join group:"
	INVITEGROUP  = "invite group:"
	RESTORENOTES = "restore:"
	CLOSE        = "close" //system replace
	NULL         = ""
)

var cmds = [...]string{SIGNUP, SIGNIN, SIGNOUT, TOUSER, TOGROUP, CREATEGROUP, JOINGROUP, INVITEGROUP, RESTORENOTES, CLOSE}

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
			cmdMap[len(cmd)] = append(arr, cmd)
			continue
		}
		arr = make([]string, 0, len(cmds))
		arr = append(arr, cmd)
		cmdMap[len(cmd)] = arr
	}
	return &parser{cmdMap: cmdMap, minLen: minLen}
}

func (p *parser) parse() {
	for i := 0; i < 100; i++ {
		go func() {
			for {
				select {
				case message := <-getQueue().pullUp():
					if message.mtype == CLOSED {
						getSessionStatesIndex().dispatch(message.chid, CLOSE, NULL)
						continue
					}
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
				return cmd, strings.TrimSpace(str[k:len(str)])
			}
		}
	}
	return NONCOMMAND, strings.TrimSpace(str)
}
