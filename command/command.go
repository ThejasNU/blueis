package command

import "net"

type Command struct{
	Args []string
	Connection net.Conn
}