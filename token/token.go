package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// signature + literal
	IDENT = "IDENT" // add, foobar, x,y
	INT   = "INT"

	// operator
	ASSIGN = "="
	PLUS   = "+"

	// delimeter
	COMMA     = ","
	SEMICOLON = ";"

	// brunket
	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "}"

	// keyword
	FUNCTION = "FUNCTION"
	LET      = "LET"
)
