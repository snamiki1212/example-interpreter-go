package lexer

type Lexer struct {
	input        string
	position     int  // current cursor
	readPosition int  // next cursor
	ch           byte // current checking char
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) { // NOTE: when EOF
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}
