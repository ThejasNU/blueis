package command

import "net"

type Command struct{
	args []string
	conn net.Conn
}