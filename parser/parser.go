package parser

import (
	"bufio"
	"errors"
	"net"
	"strconv"

	"github.com/ThejasNU/blueis/command"
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

//############ Helpers #######################################################

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


//############ Arguments Parsers #######################################################

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

//############ Commands Creator #######################################################
func (parser *Parser) GetCommand() (*command.Command,error) {
	b,err := parser.reader.ReadByte()

	if err!=nil{
		return &command.Command{},err
	}

	//if it starts with * ,RESP array is used or else a single line is input
	if b=='*'{
		return parser.parseRespArray()
	} else{
		newLine,err:=parser.readLine()
		if err!=nil{
			return &command.Command{},err
		}
		parser.index=0
		parser.line=append([]byte{},b)
		parser.line=append(parser.line, newLine...)
		return parser.parseInline()
	}
}

//parses single line commands
func (parser *Parser) parseInline() (*command.Command,error){
	for parser.current()==' '{
		parser.index++
	}

	cmd:=command.Command{
		Connection: parser.connection,
	}

	for !parser.atEnd(){
		arg,err:=parser.parserArg()

		if err!=nil{
			return &cmd,err
		}

		if arg!=""{
			cmd.Args = append(cmd.Args, arg)
		}
	}

	return &cmd,nil
}

//function to parser RESP commands
func (parser *Parser) parseRespArray() (*command.Command,error){
	cmd:=command.Command{
		Connection: parser.connection,
	}

	input,err:=parser.readLine()
	if err!=nil{
		return &cmd,err
	}

	//first character after * tells how many elements are in RESP array
	inputArrayLength,_:=strconv.Atoi(string(input))

	for i:=0;i<inputArrayLength;i++{
		symbol,err:=parser.reader.ReadByte()
		if err!=nil{
			return &cmd,err
		}

		switch symbol{
		case ':':
			//denotes intergers
			arg,err:=parser.readLine()
			if err!=nil{
				return &cmd,err
			}
			cmd.Args = append(cmd.Args, string(arg))
		
		case '$':
			//denotes multiple strings
			arg,err:=parser.readLine()
			if err!=nil{
				return &cmd,err
			}
			length,_:=strconv.Atoi(string(arg))
			text:=make([]byte,0)
			for len(text)<length{
				line,err:=parser.readLine()
				if err!=nil{
					return &cmd,err
				}
				text = append(text, line...)
			}
			cmd.Args = append(cmd.Args, string(text[:length]))
		case '*':
			//denotes array
			next,err:=parser.parseRespArray()
			if err!=nil{
				return &cmd,err
			}
			cmd.Args = append(cmd.Args, next.Args...)
		}
	}
	
	return &cmd,nil
}