package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	help       bool
	addr       string
	VERSION    string
	BUILD_TIME string
	GO_VERSION string
)

func main() {
	flag.BoolVar(&help, "h", false, "help")
	flag.StringVar(&addr, "addr", ":1202", "listen addr")
	flag.Usage = usage
	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	parser := newParser()
	go parser.parse()

	accepter := newAccepter()
	accepter.serve(addr)
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n -addr=127.0.0.1:1202\n\nInfo:\n VERSION: %s\n BUILD_TIME: %s\n GO_VERSION: %s\n", VERSION, BUILD_TIME, GO_VERSION)
}
