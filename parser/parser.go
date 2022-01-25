package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // ex) add(2, 3)
)

// priority orders for infix parser
var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
}

type Parser struct {
	l      *lexer.Lexer
	errors []string

	// token
	curToken  token.Token // current token
	peekToken token.Token // next token

	// x-fix map
	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}

	// setup prefix
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)

	// setup infix
	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)

	// setup tokens
	p.nextToken() // for current
	p.nextToken() // for next

	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

///
/// for call expression
///
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseCallArguments()
	return exp
}

///
/// for call arguments
///
func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	if p.peekTokenIs(token.RPAREN) {
		// no args case
		p.nextToken()
		return args
	}

	p.nextToken() // skip ')'

	// first arg
	args = append(args, p.parseExpression(LOWEST))

	// second... args
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // skip prev arg
		p.nextToken() // skip ','
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RPAREN) {
		// invalid syntax
		return nil
	}

	return args
}

///
/// for function
///
func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	lit.Parameters = p.parseFunctionParameters()
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	lit.Body = p.parseBlockStatement()
	return lit
}

///
/// for arguments on func
///
func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers // no arguments pattern
	}
	p.nextToken() // skip '('

	// for first arg
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	// for second arg ~ ...
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // skip arg
		p.nextToken() // skip ','
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		// invalid syntax
		return nil
	}

	return identifiers
}

///
/// for GroupedExpression
///
func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return exp
}

///
/// for If
///
func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}
	if !p.expectPeek(token.LPAREN) { // NOTE: (
		return nil
	}
	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) { // NOTE: )
		return nil
	}
	if !p.expectPeek(token.LBRACE) { // NOTE: {
		return nil
	}
	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken() // skip 'else'
		if !p.expectPeek(token.LBRACE) {
			return nil
		}
		expression.Alternative = p.parseBlockStatement()
	}
	return expression
}

///
/// for BlockStatement
///
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}
	p.nextToken()
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	return block
}

///
/// for Statement
///
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

///
/// for Let
///
func (p *Parser) parseLetStatement() ast.Statement {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	// TODO: skip until semicolon
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

///
/// for Return
///
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()

	// TODO: skip till semicolon
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

///
/// for Expression-Statement
///
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	defer untrace(trace("parseExpressionStatement"))
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

///
/// for Expression
///
func (p *Parser) parseExpression(precedence int) ast.Expression {
	defer untrace(trace("parseExpression"))

	/// prefix block
	/// - run prefixParser if exists
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	/// infix block
	/// - run infixParser if exists
	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

///
/// for Identifier
///
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

///
/// for IntegerLiteral
///
func (p *Parser) parseIntegerLiteral() ast.Expression {
	defer untrace(trace("parseInteferLiteral"))
	lit := &ast.IntegerLiteral{Token: p.curToken}
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

///
/// for prefix expression
///
func (p *Parser) parsePrefixExpression() ast.Expression {
	defer untrace(trace("parsePrefixExpression"))
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

///
/// for infix expression
///
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	defer untrace(trace("parseInfixExpression"))
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	// if expression.Operator == "+" {
	// 	expression.Right = p.parseExpression(precedence - 1)
	// } else {
	expression.Right = p.parseExpression(precedence)
	// }
	return expression
}

///
/// for bool
///
func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

///
/// helpers
///
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// assert that next token type is same as arg type, and move to next token
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression // arg is left side val
)

func (p *Parser) registerPrefix(
	tokenType token.TokenType,
	fn prefixParseFn,
) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(
	tokenType token.TokenType,
	fn infixParseFn,
) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}
