package main

import (
	"log"
	"net"

	"github.com/ThejasNU/blueis/parser"
)

func main(){
	listener,err:=net.Listen("tcp", ":6380")
	if err!=nil{
		log.Fatal(err)
	}
	
	log.Println("Listening on tcp://0.0.0.0:6380")

	for{
		connection,err:=listener.Accept()
		log.Println("New connection",connection)

		if err!=nil{
			log.Fatal(err)
		}

		go startSession(connection)
	}
}

func startSession(connection net.Conn){
	defer func(){
		log.Panicln("Closing connection",connection)
		connection.Close()
	}()

	defer func(){
		if err:=recover();err!=nil{
			log.Println("Recovering from error",err)
		}
	}()

	prsr:=parser.NewParser(connection)

	for{
		cmd,err:=prsr.GetCommand()

		if err!=nil{
			connection.Write([]byte("-ERR " + err.Error() + "\r\n"))
			break
		}

		if !cmd.Handle(){
			break
		}
	}
}
