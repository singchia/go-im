package main

import (
	"net"
	"os"
	"time"

	"github.com/singchia/go-hammer/doublinker"
)

type accepter struct {
	linker *doublinker.Doublinker
}

func (a *accepter) Serve() {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			os.Exit(1)
		}

		ch := make(chan []byte, 1024)
		chID := a.linker.Add(ch)
		go a.handle(conn, ch, chID)
	}
}

func (a *accepter) handle(conn net.Conn, ch <-chan []byte, chID doublinker.DoubID) {
	//in case of blocking the loop
	conn.SetReadDeadline(time.Now().Add(time.Microsecond * 500))
	buf := make([]byte, 1024)
	for {
		select {
		case data <- ch:
			conn.Write(data)
		default:
			len, err := conn.Read(buf)
			if err != nil {
				//delete node in linker
				a.linker.Delete(chID)
				//notify the user system
				GetQueueInstance().Push(&message{messageType: CLOSED, chID: chID, data: nil})
				conn.Close()
				return
			}
		}
	}
}
