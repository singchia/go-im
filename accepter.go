package main

import (
	"bufio"
	"net"
	"os"
	"time"

	"github.com/singchia/go-hammer/doublinker"
)

type accepter struct {
	linker *doublinker.Doublinker
}

func (a *accepter) serve() {
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
		chid := a.linker.Add(ch)
		go a.handle(conn, ch, chid)
	}
}

func (a *accepter) dispatch() {
	for i := 0; i < 100; i++ {
		go func() {
			for {
				select {
				case message := <-GetQueueInstance().pullDown():
					a.linker.Retrieve(message.chid).(chan []byte) <- message.data
				}
			}
		}()
	}
}

func (a *accepter) handle(conn net.Conn, ch <-chan []byte, chid doublinker.DoubID) {
	//in case of blocking the loop
	conn.SetReadDeadline(time.Now().Add(time.Microsecond * 50))
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	buf := make([]byte, 1024)
	for {
		select {
		case data <- ch:
			_, err := writer.WriteString(data)
			if err != nil {
				a.linker.Delete(chid)
				GetQueueInstance().pushUp(&message{messageType: CLOSED, chid: chid})
				return
			}
		default:
			str, err := reader.ReadString('\n')
			E, ok := err.(net.Error)
			if ok && E.Timeout() {
				continue
			}
			if err != nil {
				a.linker.Delete(chid)
				GetQueueInstance().pushUp(&message{messageType: CLOSED, chid: chid})
				conn.Close()
				return
			}
			GetQueueInstance().pullUp(&message{messageType: DATA, chid: chid, data: str})
		}
	}
}
