package main

import (
	"bufio"
	"log"
	"net"
)

type socket struct {
	requestCh chan request
}

type request struct {
	command string
}

func newSocket() *socket {
	socket := socket{}
	socket.requestCh = make(chan request)
	return &socket
}

func (s *socket) listen() {
	ln, err := net.Listen("tcp", ":10005")
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
		}

		go s.handle(conn)
	}
}

func (s *socket) handle(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}

		s.requestCh <- request{message}
	}
}
