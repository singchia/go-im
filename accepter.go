package main

import (
	"bufio"
	"net"
	"os"
	"time"

	"github.com/singchia/go-hammer/doublinker"
)

const (
	UNPARSED = iota
	PASSTHROUGH
	CLOSED
)

type accepter struct {
	linker *doublinker.Doublinker
}

func newAccepter() *accepter {
	return &accepter{linker: doublinker.NewDoublinker()}
}

func (a *accepter) serve(addr string) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		os.Exit(1)
	}
	defer l.Close()

	a.dispatch()

	for {
		conn, err := l.Accept()
		if err != nil {
			os.Exit(1)
		}

		ch := make(chan string, 1024)
		chid := a.linker.Add(ch)
		go a.handle(conn, ch, chid)
	}
}

func (a *accepter) dispatch() {
	for i := 0; i < 100; i++ {
		go func() {
			for {
				select {
				case message := <-getQueue().pullDown():
					a.linker.Retrieve(message.chid).(chan string) <- message.data
				}
			}
		}()
	}
}

func (a *accepter) handle(conn net.Conn, ch <-chan string, chid doublinker.DoubID) {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	for {
		select {
		case data := <-ch:
			_, err := writer.WriteString(data)
			if err != nil {
				a.linker.Delete(chid)
				getQueue().pushUp(&message{mtype: CLOSED, chid: chid})
				return
			}
			writer.Flush()
		default:
			//in case of blocking the loop
			conn.SetReadDeadline(time.Now().Add(time.Millisecond * 200))
			//str, _, err := reader.ReadLine()
			str, err := reader.ReadString('\n')
			E, ok := err.(net.Error)
			if ok && E.Timeout() {
				continue
			}
			if err != nil {
				a.linker.Delete(chid)
				getQueue().pushUp(&message{mtype: CLOSED, chid: chid})
				conn.Close()
				return
			}
			getQueue().pushUp(&message{mtype: UNPARSED, chid: chid, data: str})
		}
	}
}
