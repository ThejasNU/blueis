package main

import (
	"log"
	"net"

	"github.com/ThejasNU/blueis/parser"
)

func main() {
	listener, err := net.Listen("tcp", ":6380")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Listening on tcp://0.0.0.0:6380")

	for {
		connection, err := listener.Accept()
		log.Println("New connection", connection)

		if err != nil {
			log.Fatal(err)
		}

		go startSession(connection)
	}
}

func startSession(connection net.Conn) {
	prsr := parser.NewParser(connection)

	for {
		cmd, err := prsr.GetCommand()

		if err != nil {
			connection.Write([]byte("-ERR " + err.Error() + "\r\n"))
			break
		}

		if !cmd.Handle() {
			log.Println("Closing connection", connection)
			connection.Close()
			return
		}
	}

}
