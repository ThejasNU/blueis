package parser

import (
	"bufio"
	"errors"
	"net"
)

type Parser struct{
	connection net.Conn
	reader *bufio.Reader
	line []byte
	index int
}

func NewParser(conn net.Conn) *Parser{
	return &Parser{
		connection: conn,
		reader: bufio.NewReader(conn),
		line: make([]byte, 0),
		index: 0,
	}
}

//tells whether we are at the end of the line
func (parser *Parser) atEnd() bool{
	return parser.index>=len(parser.line)
}

//return the current character
func (parser *Parser) current() byte{
	if parser.atEnd() {
		return '\r'
	}
	return parser.line[parser.index]
}

//read commands line from the input
func (parser *Parser) readLine() ([]byte,error) {
	//read till we get \r
	line,err:=parser.reader.ReadBytes('\r')
	if err!=nil{
		return nil,err
	}

	//checks for \n after \r
	if _,err=parser.reader.ReadByte();err!=nil{
		return nil,err
	}

	return line[:len(line)-1],nil
}

//reads argument from the current line
func (parser *Parser) parserArg() (s string,err error){
	for parser.current()==' '{
		parser.index++
	}

	if parser.current()=='"'{
		parser.index++
		buf,err:=parser.parseString()
		return string(buf),err
	}

	for !parser.atEnd() && parser.current()!=' ' && parser.current() != '\r'{
		s+=string(parser.current())
		parser.index++
	}

	return
}

//reads a string argument from the current line, special case of above function
func (parser *Parser) parseString() (stream []byte,err error){
	for parser.current()!='"' && !parser.atEnd(){
		cur:=parser.current()
		parser.index++
		next:=parser.current()

		if cur=='\\' && next=='"'{
			stream = append(stream, '"')
			parser.index++
		} else{
			stream=append(stream, cur)
		}
	}

	if parser.current()!='"'{
		return nil,errors.New("unbalanced quotes in request")
	}
	parser.index++

	return
}