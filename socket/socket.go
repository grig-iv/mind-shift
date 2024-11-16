package socket

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"
	"strings"
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

	split := strings.Split(message, " ")

	parser, ok := parsers[split[0]]
	if !ok {
		return UnknownCmd{message}
	}

	return parser(split[1:])
}
