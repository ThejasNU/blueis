package parser

import (
	"bufio"
	"errors"
	"net"
	"strconv"

	"github.com/ThejasNU/blueis/command"
)

type Parser struct {
	connection net.Conn
	reader     *bufio.Reader
	line       []byte
	index      int
}

func NewParser(conn net.Conn) *Parser {
	return &Parser{
		connection: conn,
		reader:     bufio.NewReader(conn),
		line:       make([]byte, 0),
		index:      0,
	}
}

//############ Helpers #######################################################

// tells whether we are at the end of the line
func (prsr *Parser) atEnd() bool {
	return prsr.index >= len(prsr.line)
}

// return the current character
func (prsr *Parser) current() byte {
	if prsr.atEnd() {
		return '\r'
	}
	return prsr.line[prsr.index]
}

//############ Arguments Parsers #######################################################

// read commands line from the input
func (prsr *Parser) readLine() ([]byte, error) {
	//read till we get \r
	line, err := prsr.reader.ReadBytes('\r')
	if err != nil {
		return nil, err
	}

	//checks for \n after \r
	if _, err = prsr.reader.ReadByte(); err != nil {
		return nil, err
	}

	return line[:len(line)-1], nil
}

// reads argument from the current line
func (prsr *Parser) parserArg() (s string, err error) {
	for prsr.current() == ' ' {
		prsr.index++
	}

	if prsr.current() == '"' {
		prsr.index++
		buf, err := prsr.parseString()
		return string(buf), err
	}

	for !prsr.atEnd() && prsr.current() != ' ' && prsr.current() != '\r' {
		s += string(prsr.current())
		prsr.index++
	}

	return
}

// reads a string argument from the current line, special case of above function
func (prsr *Parser) parseString() (stream []byte, err error) {
	for prsr.current() != '"' && !prsr.atEnd() {
		cur := prsr.current()
		prsr.index++
		next := prsr.current()

		if cur == '\\' && next == '"' {
			stream = append(stream, '"')
			prsr.index++
		} else {
			stream = append(stream, cur)
		}
	}

	if prsr.current() != '"' {
		return nil, errors.New("unbalanced quotes in request")
	}
	prsr.index++

	return
}

// ############ Commands Creator #######################################################
func (prsr *Parser) GetCommand() (*command.Command, error) {
	b, err := prsr.reader.ReadByte()

	if err != nil {
		return &command.Command{}, err
	}

	//if it starts with * ,RESP array is used or else a single line is input
	if b == '*' {
		return prsr.parseRespArray()
	} else {
		newLine, err := prsr.readLine()
		if err != nil {
			return &command.Command{}, err
		}
		prsr.index = 0
		prsr.line = append([]byte{}, b)
		prsr.line = append(prsr.line, newLine...)
		return prsr.parseInline()
	}
}

// parses single line commands
func (prsr *Parser) parseInline() (*command.Command, error) {
	for prsr.current() == ' ' {
		prsr.index++
	}

	cmd := command.Command{
		Connection: prsr.connection,
	}

	for !prsr.atEnd() {
		arg, err := prsr.parserArg()

		if err != nil {
			return &cmd, err
		}

		if arg != "" {
			cmd.Args = append(cmd.Args, arg)
		}
	}

	return &cmd, nil
}

// function to parse RESP commands
func (prsr *Parser) parseRespArray() (*command.Command, error) {
	cmd := command.Command{
		Connection: prsr.connection,
	}

	input, err := prsr.readLine()
	if err != nil {
		return &cmd, err
	}

	//first character after * tells how many elements are in RESP array
	inputArrayLength, _ := strconv.Atoi(string(input))

	for i := 0; i < inputArrayLength; i++ {
		symbol, err := prsr.reader.ReadByte()
		if err != nil {
			return &cmd, err
		}

		switch symbol {
		case ':':
			//denotes intergers
			arg, err := prsr.readLine()
			if err != nil {
				return &cmd, err
			}
			cmd.Args = append(cmd.Args, string(arg))

		case '$':
			//denotes multiple strings
			arg, err := prsr.readLine()
			if err != nil {
				return &cmd, err
			}
			length, _ := strconv.Atoi(string(arg))
			text := make([]byte, 0)
			for len(text) < length {
				line, err := prsr.readLine()
				if err != nil {
					return &cmd, err
				}
				text = append(text, line...)
			}
			cmd.Args = append(cmd.Args, string(text[:length]))
		case '*':
			//denotes array
			next, err := prsr.parseRespArray()
			if err != nil {
				return &cmd, err
			}
			cmd.Args = append(cmd.Args, next.Args...)
		}
	}

	return &cmd, nil
}
