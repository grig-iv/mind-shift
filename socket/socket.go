package socket

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"

	"github.com/grig-iv/mind-shift/domain"
)

var CommandCh = make(chan Cmd)

func Listen() {
	ln, err := net.Listen("tcp", ":10005")
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
		}

		go handle(conn)
	}
}

func handle(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) == false {
				log.Println(err)
			}
			return
		}

		CommandCh <- parseCmd(message[:len(message)-1])
	}
}

func parseCmd(message string) Cmd {
	log.Println("Recive message:", message)

	switch message {
	case "go-to-tag -n":
		return GoToTagCmd{Dir: domain.Next}
	case "go-to-tag -p":
		return GoToTagCmd{Dir: domain.Prev}
	case "move-to-tag -n":
		return MoveToTagCmd{Dir: domain.Next}
	case "move-to-tag -p":
		return MoveToTagCmd{Dir: domain.Prev}
	case "kill-client":
		return KillClientCmd{}
	case "quit":
		return QuitCmd{}
	}

	return message
}
