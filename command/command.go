package command

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

var cache sync.Map

type Command struct {
	Args       []string
	Connection net.Conn
}

// bool here tells when the connection should be closed
func (cmd *Command) Handle() bool {
	switch strings.ToUpper(cmd.Args[0]) {
	case "GET":
		return cmd.Get()
	case "SET":
		return cmd.Set()
	case "DEL":
		return cmd.Del()
	case "QUIT":
		return cmd.Quit()
	default:
		log.Println("Command not supported", cmd.Args[0])
		cmd.Connection.Write([]byte("-ERR unknown command '" + cmd.Args[0] + "'\r\n"))
	}
	return true
}

func (cmd *Command) Get() bool {
	if len(cmd.Args) != 2 {
		cmd.Connection.Write([]byte("-ERR wrong number of arguments for '" + cmd.Args[0] + "' command\r\n"))
		return true
	}
	key := cmd.Args[1]
	value, _ := cache.Load(key)
	if value != nil {
		//use this kind of type assertion for interfaces
		output, _ := value.(string)
		if strings.HasPrefix(output, "\"") {
			output, _ = strconv.Unquote(output)
		}

		cmd.Connection.Write([]byte(fmt.Sprintf("$%d\r\n", len(output))))
		cmd.Connection.Write(append([]byte(output), []byte("\r\n")...))
	} else {
		cmd.Connection.Write([]byte("$-1\r\n"))
	}

	return true
}

func (cmd *Command) Set() bool {
	numArgs := len(cmd.Args)
	if numArgs < 3 || numArgs > 6 {
		cmd.Connection.Write([]byte("-ERR wrong number of arguments for '" + cmd.Args[0] + "' command\r\n"))
		return true
	}

	if numArgs > 3 {
		index := 3
		option := strings.ToUpper(cmd.Args[index])
		_, ok := cache.Load(cmd.Args[1])
		switch option {
		case "NX":
			//only set the key if it does not already exist
			if ok {
				cmd.Connection.Write([]byte("$-1\r\n"))
				return true
			}
			index++
		case "XX":
			//only set the key if it already exist
			if !ok {
				cmd.Connection.Write([]byte("$-1\r\n"))
				return true
			}
			index++
		}

		if index < numArgs {
			if err := cmd.setExpiration(index); err != nil {
				cmd.Connection.Write([]byte("-ERR " + err.Error() + "\r\n"))
				return true
			}
		}
	}
	cache.Store(cmd.Args[1], cmd.Args[2])
	cmd.Connection.Write([]byte("+OK\r\n"))
	return true
}

func (cmd *Command) Del() bool {
	if len(cmd.Args) < 2 {
		cmd.Connection.Write([]byte("-ERR wrong number of arguments for '" + cmd.Args[0] + "' command\r\n"))
		return true
	}

	count := 0

	for _, key := range cmd.Args[1:] {
		if _, ok := cache.LoadAndDelete(key); ok {
			count++
		}
	}

	cmd.Connection.Write([]byte(fmt.Sprintf(":%d\r\n", count)))
	return true
}

func (cmd *Command) Quit() bool {
	if len(cmd.Args) != 1 {
		cmd.Connection.Write([]byte("-ERR wrong number of arguments for '" + cmd.Args[0] + "' command\r\n"))
		return true
	}

	cmd.Connection.Write([]byte("+OK\r\n"))
	return false
}

func (cmd *Command) setExpiration(index int) error {
	option := strings.ToUpper(cmd.Args[index])
	value, _ := strconv.Atoi(cmd.Args[index+1])

	var duration time.Duration
	switch option {
	case "EX":
		duration = time.Second * time.Duration(value)
	case "PX":
		duration = time.Millisecond * time.Duration(value)
	default:
		return fmt.Errorf("expiration option is not valid")
	}

	go func() {
		time.Sleep(duration)
		cache.Delete(cmd.Args[1])
	}()

	return nil
}
