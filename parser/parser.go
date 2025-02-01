package parser

import (
	"RoLang/ast"
	"RoLang/lexer"
	"RoLang/token"

	"fmt"
	"strconv"
)

type Parser struct {
	lexer *lexer.Lexer

	errors []string

	currToken token.Token
	nextToken token.Token

	table [token.TOTAL]Entry
}

type (
	Precedence   uint
	prefixParser func() ast.Expression
	infixParser  func(ast.Expression) ast.Expression
)

type Entry struct {
	prefix     prefixParser
	infix      infixParser
	precedence Precedence
}

const (
	NONE    Precedence = iota
	ASSIGN             // =
	EQUALS             // == !=
	COMPARE            // < > <= >=
	SUM                // + -
	PRODUCT            // * /
	PREFIX             // !x -x
	POSTFIX            // x() x++
)

func New(lexer *lexer.Lexer) *Parser {
	// Allocating on heap
	p := &Parser{
		lexer:  lexer,
		errors: []string{},
	}

	p.table = [token.TOTAL]Entry{
		// prefix expression do not need a precedence
		token.LPAREN: {p.parseGroupedExpression, nil, NONE},
		token.IDENT:  {p.parseIdentifier, nil, NONE},
		token.INT:    {p.parseIntegerLiteral, nil, NONE},
		token.FLOAT:  {p.parseFloatLiteral, nil, NONE},
		token.TRUE:   {p.parseBoolLiteral, nil, NONE},
		token.FALSE:  {p.parseBoolLiteral, nil, NONE},
		token.BANG:   {p.parsePrefixExpression, nil, NONE},
		token.MINUS:  {p.parsePrefixExpression, p.parseInfixExpression, SUM},
		token.PLUS:   {nil, p.parseInfixExpression, SUM},
		token.STAR:   {nil, p.parseInfixExpression, PRODUCT},
		token.SLASH:  {nil, p.parseInfixExpression, PRODUCT},
		token.EQ:     {nil, p.parseInfixExpression, EQUALS},
		token.NE:     {nil, p.parseInfixExpression, EQUALS},
		token.LT:     {nil, p.parseInfixExpression, COMPARE},
		token.LE:     {nil, p.parseInfixExpression, COMPARE},
		token.GT:     {nil, p.parseInfixExpression, COMPARE},
		token.GE:     {nil, p.parseInfixExpression, COMPARE},
	}

	// Read two tokens, to set currToken and nextToken
	p.readToken()
	p.readToken()

	return p
}

func (p *Parser) Parse() *ast.Program {
	// Allocating on heap
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	// Read until end of file
	for p.currToken.Type != token.EOF {
		stmt := p.parseStatement()
		program.Statements = append(program.Statements, stmt)

		// Set parser on the first token of next statement
		p.readToken()
	}

	return program
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.currToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.IF:
		return p.parseIfStatement()
	case token.LBRACE:
		return p.parseBlockStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.currToken}

	// match and consume an identifier
	if !p.expectToken(token.IDENT) {
		return nil
	}

	stmt.Ident = &ast.Identifier{
		Token: p.currToken,
		Value: p.currToken.Word,
	}

	// match and consume an equals
	if !p.expectToken(token.ASSIGN) {
		return nil
	}

	for p.currToken.Type != token.SEMCOL {
		p.readToken()
	}

	return stmt
}

func (p *Parser) parseIfStatement() ast.Statement {
	stmt := &ast.IfStatement{Token: p.currToken}
	// consume 'if'
	p.readToken()

	// no parenthesis is necessary we straight
	// away parse condition expression
	stmt.Condition = p.parseExpression(NONE)

	if !p.expectToken(token.LBRACE) {
		return nil
	}

	stmt.Then = p.parseBlockStatement()

	// check for 'else'
	if p.peekToken(token.ELSE) {
		// consume 'else'
		p.readToken()

		// check if next token is 'if'
		if p.matchToken(token.IF) {
			stmt.Else = p.parseIfStatement()
		} else if p.matchToken(token.LBRACE) {
			stmt.Else = p.parseBlockStatement()
		} else {
			p.report(fmt.Sprintf("expected 'if' or '{'. found %q", p.nextToken.Word))
			return nil
		}
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.currToken}

	// consume 'return' token
	p.readToken()

	for p.currToken.Type != token.SEMCOL {
		p.readToken()
	}

	return stmt
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.currToken}
	block.Statements = []ast.Statement{}

	// consume '{' token
	p.readToken()

	for p.currToken.Type != token.RBRACE && p.currToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}

		p.readToken() // read next statement's token
	}

	return block
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.currToken}

	stmt.Expression = p.parseExpression(NONE)

	if p.peekToken(token.SEMCOL) {
		p.readToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence Precedence) ast.Expression {
	prefix := p.table[p.currToken.Type].prefix
	if prefix == nil {
		p.noPrefixFuncError(p.currToken.Type)
		return nil
	}
	expr := prefix()

	// keep consuming tokens until we run into a semi colon or the next token's precedence is greater
	// than current token's precedence
	for !p.peekToken(token.SEMCOL) && precedence < p.table[p.nextToken.Type].precedence {
		infix := p.table[p.nextToken.Type].infix
		// only prefix expression
		if infix == nil {
			return expr
		}

		p.readToken()
		expr = infix(expr)
	}

	return expr
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expr := &ast.PrefixExpression{
		Token:    p.currToken,
		Operator: p.currToken.Word,
	}
	p.readToken()
	expr.Right = p.parseExpression(PREFIX)

	return expr
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expr := &ast.InfixExpression{
		Token:    p.currToken,
		Operator: p.currToken.Word,
		Left:     left,
	}

	// get current token's precedence
	precedence := p.table[p.currToken.Type].precedence
	// consume current token
	p.readToken()
	// start parsing the next token and use current token's precedence
	expr.Right = p.parseExpression(precedence)

	return expr
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.readToken()

	expr := p.parseExpression(NONE)

	if !p.expectToken(token.RPAREN) {
		return nil
	}

	return expr
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{
		Token: p.currToken,
		Value: p.currToken.Word,
	}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	l := &ast.IntegerLiteral{Token: p.currToken}

	value, err := strconv.ParseInt(p.currToken.Word, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer. %s",
			p.currToken.Word, err)
		p.errors = append(p.errors, msg)
		return nil
	}

	l.Value = value

	return l
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	l := &ast.FloatLiteral{Token: p.currToken}

	value, err := strconv.ParseFloat(p.currToken.Word, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float. %s",
			p.currToken.Word, err)
		p.errors = append(p.errors, msg)
		return nil
	}

	l.Value = value

	return l
}

func (p *Parser) parseBoolLiteral() ast.Expression {
	return &ast.BoolLiteral{
		Token: p.currToken,
		Value: p.currToken.Type == token.TRUE,
	}
}

func (p *Parser) matchToken(tokenType token.TokenType) bool {
	if p.peekToken(tokenType) {
		p.readToken()
		return true
	}

	return false
}

func (p *Parser) expectToken(tokenType token.TokenType) bool {
	if !p.matchToken(tokenType) {
		p.peekError(tokenType)
		return false
	}

	return true
}

func (p *Parser) peekToken(tokType token.TokenType) bool {
	return p.nextToken.Type == tokType
}

func (p *Parser) readToken() {
	p.currToken = p.nextToken
	p.nextToken = p.lexer.NextToken()
}

func (p *Parser) peekError(tokenType token.TokenType) {
	p.report(fmt.Sprintf("expected next token to be %q, got %q instead",
		token.TokenString[tokenType], p.nextToken.Word))
}

func (p *Parser) noPrefixFuncError(tokenType token.TokenType) {
	p.report(fmt.Sprintf("no prefix parse function for %q found",
		token.TokenString[tokenType]))
}

func (p *Parser) report(message string) {
	message = fmt.Sprintf("%s %s", p.currToken.Loc, message)
	p.errors = append(p.errors, message)
}
