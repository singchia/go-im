package main

func main() {
	parser := newParser()
	go parser.parse()

	accepter := newAccepter()
	accepter.serve("127.0.0.1:1202")
}
