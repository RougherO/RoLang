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
	// all error messages generated while parsing
	errors []error
	// pointers for reading tokens
	currToken token.Token
	nextToken token.Token
	// pratt table
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
	POSTFIX            // x() x++ x[1]
	DOT                // a.b
)

func New(lexer *lexer.Lexer) *Parser {
	p := &Parser{
		lexer:  lexer,
		errors: []error{},
	}

	// TODO put this into a global variable so that every time a parser
	// is created it refers to the same table instead of creating and filling
	// a new table
	p.table = [token.TOTAL]Entry{
		// prefix expression do not need a precedence
		token.LBRACE: {p.parseMapLiteral, nil, NONE},
		token.STRING: {p.parseStringLiteral, nil, NONE},
		token.IDENT:  {p.parseIdentifier, nil, NONE},
		token.FN:     {p.parseFunctionLiteral, nil, NONE},
		token.INT:    {p.parseIntegerLiteral, nil, NONE},
		token.FLOAT:  {p.parseFloatLiteral, nil, NONE},
		token.TRUE:   {p.parseBoolLiteral, nil, NONE},
		token.FALSE:  {p.parseBoolLiteral, nil, NONE},
		token.NULL:   {p.parseNullLiteral, nil, NONE},
		token.BANG:   {p.parsePrefixExpression, nil, NONE},
		token.MINUS:  {p.parsePrefixExpression, p.parseInfixExpression, SUM},
		token.LPAREN: {p.parseGroupedExpression, p.parseCallExpression, POSTFIX},
		token.LBRACK: {p.parseArrayLiteral, p.parseIndexExpression, POSTFIX},
		token.ASSIGN: {nil, p.parseAssignExpression, ASSIGN},
		token.PLUS:   {nil, p.parseInfixExpression, SUM},
		token.STAR:   {nil, p.parseInfixExpression, PRODUCT},
		token.SLASH:  {nil, p.parseInfixExpression, PRODUCT},
		token.EQ:     {nil, p.parseInfixExpression, EQUALS},
		token.NE:     {nil, p.parseInfixExpression, EQUALS},
		token.LT:     {nil, p.parseInfixExpression, COMPARE},
		token.LE:     {nil, p.parseInfixExpression, COMPARE},
		token.GT:     {nil, p.parseInfixExpression, COMPARE},
		token.GE:     {nil, p.parseInfixExpression, COMPARE},
		token.DOT:    {nil, p.parseInfixExpression, DOT},
	}

	// Read two tokens, to set currToken and nextToken
	p.readToken()
	p.readToken()

	return p
}

func (p *Parser) Parse() (*ast.Program, []error) {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	// Read until end of file
	for !p.hasToken(token.EOF) {
		stmt := p.ParseStatement()
		program.Statements = append(program.Statements, stmt)
		// Set parser on the first token of next statement
		p.readToken()
	}

	return program, p.errors
}

func (p *Parser) ParseStatement() ast.Statement {
	switch p.currToken.Type {
	case token.LET:
		a := p.parseLetStatement()
		return a
	case token.RETURN:
		return p.parseReturnStatement()
	case token.IF:
		return p.parseIfStatement()
	case token.LBRACE:
		return p.parseBlockStatement()
	case token.FN:
		if p.peekToken(token.LPAREN) {
			return p.parseExpressionStatement()
		}
		return p.parseFunctionStatement()
	case token.LOOP:
		return p.parseLoopStatement()
	case token.BREAK:
		return p.parseJumpStatement(true)
	case token.CONT:
		return p.parseJumpStatement(false)
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) ParseExpression(precedence Precedence) ast.Expression {
	prefix := p.table[p.currToken.Type].prefix
	if prefix == nil {
		p.noPrefixFuncError(p.currToken.Type)
		return nil
	}

	expr := prefix()
	if expr == nil {
		return nil
	}

	// keep consuming tokens until next token's precedence
	// is greater than current token's precedence
	for precedence < p.table[p.nextToken.Type].precedence {
		infix := p.table[p.nextToken.Type].infix
		if infix == nil { // only prefix expression
			return expr
		}

		p.readToken()
		expr = infix(expr)
	}

	return expr
}

func (p *Parser) parseFunctionStatement() *ast.FunctionStatement {
	stmt := &ast.FunctionStatement{Token: p.currToken}

	if !p.expectToken(token.IDENT) {
		return nil
	}

	stmt.Ident = &ast.Identifier{
		Token: p.currToken,
		Value: p.currToken.Word,
	}

	// assertive check for '('
	if !p.expectToken(token.LPAREN) {
		return nil
	}

	parameters := p.parseFunctionParameters()
	if parameters == nil {
		return nil
	}

	// assertive check for '{'
	if !p.expectToken(token.LBRACE) {
		return nil
	}

	body := p.parseBlockStatement()
	if body == nil {
		return nil
	}

	stmt.Value = &ast.FunctionLiteral{
		Token:      stmt.Token,
		Parameters: parameters,
		Body:       body,
	}

	return stmt
}

func (p *Parser) parseLoopStatement() ast.Statement {
	stmt := &ast.LoopStatement{Token: p.currToken}

	if !p.peekToken(token.LBRACE) {
		p.readToken() // consume `loop`
		cond := p.ParseExpression(NONE)
		if cond == nil {
			return nil
		}
		stmt.Condition = cond
	}

	if !p.expectToken(token.LBRACE) {
		return nil
	}

	body := p.parseBlockStatement()
	if body == nil {
		return nil
	}

	stmt.Body = body
	return stmt
}

func (p *Parser) parseJumpStatement(isBreak bool) ast.Statement {
	stmt := &ast.JumpStatement{
		Token:   p.currToken,
		IsBreak: isBreak,
	}

	if !p.expectToken(token.SEMCOL) {
		return nil
	}

	return stmt
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
	p.readToken()

	initValue := p.ParseExpression(NONE)
	if initValue == nil {
		return nil
	}

	stmt.InitValue = initValue

	// optional semi-colon token
	if !p.expectToken(token.SEMCOL) {
		return nil
	}

	return stmt
}

func (p *Parser) parseIfStatement() ast.Statement {
	stmt := &ast.IfStatement{Token: p.currToken}
	// consume 'if'
	p.readToken()

	// no parenthesis is necessary we straight
	// away parse condition expression
	condition := p.ParseExpression(NONE)
	if condition == nil {
		return nil
	}

	stmt.Condition = condition

	if !p.expectToken(token.LBRACE) {
		return nil
	}

	then := p.parseBlockStatement()
	stmt.Then = then

	// check for 'else'
	if p.peekToken(token.ELSE) {
		// consume 'else'
		p.readToken()

		// check if next token is 'if'
		if p.matchToken(token.IF) {
			elze := p.parseIfStatement()
			if elze == nil {
				return nil
			}

			stmt.Else = elze
		} else if p.matchToken(token.LBRACE) {
			elze := p.parseBlockStatement()
			if elze == nil {
				return nil
			}

			stmt.Else = elze
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
	if p.hasToken(token.SEMCOL) {
		return stmt
	}

	returnValue := p.ParseExpression(NONE)
	if returnValue == nil {
		return nil
	}

	stmt.ReturnValue = returnValue
	if !p.expectToken(token.SEMCOL) {
		return nil
	}

	return stmt
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.currToken}
	block.Statements = []ast.Statement{}

	// consume '{' token
	p.readToken()

	for !p.hasToken(token.RBRACE) && !p.hasToken(token.EOF) {
		stmt := p.ParseStatement()
		if stmt == nil {
			return nil
		}
		block.Statements = append(block.Statements, stmt)

		p.readToken() // read next statement's token
	}

	if p.hasToken(token.EOF) {
		p.report("expect '}' at end of block reached end of file")
		return nil
	}

	return block
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.currToken}

	expr := p.ParseExpression(NONE)
	if expr == nil {
		return nil
	}

	stmt.Expression = expr

	if !p.expectToken(token.SEMCOL) {
		return nil
	}

	return stmt
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expr := &ast.PrefixExpression{
		Token:    p.currToken,
		Operator: p.currToken.Word,
	}
	p.readToken()
	right := p.ParseExpression(PREFIX)
	expr.Right = right

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
	right := p.ParseExpression(precedence)
	if right == nil {
		return nil
	}
	expr.Right = right

	return expr
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	expr := &ast.IndexExpression{
		Token: p.currToken,
		Left:  left,
	}

	p.readToken() // consume '['

	index := p.ParseExpression(NONE)
	if index == nil {
		return nil
	}

	if !p.expectToken(token.RBRACK) {
		return nil
	}

	expr.Index = index
	return expr
}

func (p *Parser) parseAssignExpression(left ast.Expression) ast.Expression {
	expr := &ast.AssignExpression{
		Token: p.currToken,
		Left:  left,
	}

	p.readToken() // consume '='

	// to make assignment right associative we reduce the
	// precedence before parsing the right hand side
	// to parse all assignments of rhs then assign it to this
	right := p.ParseExpression(ASSIGN - 1)
	if right == nil {
		return nil
	}

	expr.Right = right

	return expr
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.readToken()

	expr := p.ParseExpression(NONE)
	if expr == nil {
		return nil
	}

	if !p.expectToken(token.RPAREN) {
		return nil
	}

	return expr
}

func (p *Parser) parseCallExpression(callee ast.Expression) ast.Expression {
	expr := &ast.CallExpression{Token: p.currToken, Callee: callee}

	args := p.parseCallArguments()
	if args == nil {
		return nil
	}
	expr.Arguments = args

	return expr
}

func (p *Parser) parseCallArguments() []ast.Expression {
	// p.readToken() // consume '('

	args := []ast.Expression{}
	for {
		if p.peekToken(token.RPAREN) {
			break
		}
		p.readToken()

		arg := p.ParseExpression(NONE)
		if arg == nil {
			return nil
		}

		args = append(args, arg)
		if !p.peekToken(token.COMMA) {
			break
		}
		p.readToken()
	}

	if !p.expectToken(token.RPAREN) {
		return nil
	}

	return args
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{
		Token: p.currToken,
		Value: p.currToken.Word,
	}
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	fn := &ast.FunctionLiteral{Token: p.currToken}

	if !p.expectToken(token.LPAREN) {
		return nil
	}

	parameters := p.parseFunctionParameters()
	if parameters == nil {
		return nil
	}

	fn.Parameters = parameters

	if !p.expectToken(token.LBRACE) {
		return nil
	}

	body := p.parseBlockStatement()
	if body == nil {
		return nil
	}
	fn.Body = body

	return fn
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	idents := []*ast.Identifier{}

	for {
		if p.peekToken(token.RPAREN) {
			break
		}

		if !p.expectToken(token.IDENT) {
			return nil
		}

		ident := &ast.Identifier{Token: p.currToken, Value: p.currToken.Word}
		idents = append(idents, ident)

		if !p.peekToken(token.COMMA) {
			break
		}
		p.readToken() // read ','
	}

	if !p.expectToken(token.RPAREN) {
		return nil
	}

	return idents
}

func (p *Parser) parseMapLiteral() ast.Expression {
	return &ast.MapLiteral{
		Token:    p.currToken,
		Elements: p.parseMapElements(),
	}
}

func (p *Parser) parseMapElements() []ast.MapElement {
	elems := []ast.MapElement{}

	for {
		if p.peekToken(token.RBRACE) {
			break
		}

		p.readToken()

		keyExpr := p.ParseExpression(NONE)
		if keyExpr == nil {
			return nil
		}

		if !p.expectToken(token.COLON) {
			return nil
		}
		p.readToken()

		valueExpr := p.ParseExpression(NONE)
		if valueExpr == nil {
			return nil
		}
		elems = append(elems, ast.MapElement{
			Key:   keyExpr,
			Value: valueExpr,
		})

		if !p.peekToken(token.COMMA) {
			break
		}
		p.readToken() // consume ','
	}

	if !p.expectToken(token.RBRACE) {
		return nil
	}

	return elems
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	return &ast.ArrayLiteral{
		Token:    p.currToken,
		Elements: p.parseArrayElements(),
	}
}

func (p *Parser) parseArrayElements() []ast.Expression {
	elems := []ast.Expression{}

	for {
		if p.peekToken(token.RBRACK) {
			break
		}

		p.readToken()

		expr := p.ParseExpression(NONE)
		elems = append(elems, expr)

		if !p.peekToken(token.COMMA) {
			break
		}
		p.readToken() // read ','
	}

	if !p.expectToken(token.RBRACK) {
		return nil
	}

	return elems
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{
		Token: p.currToken,
		Value: p.currToken.Word,
	}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	l := &ast.IntegerLiteral{Token: p.currToken}

	value, err := strconv.ParseInt(p.currToken.Word, 0, 64)
	if err != nil {
		err = fmt.Errorf("could not parse %q as integer. %s",
			p.currToken.Word, err)
		p.errors = append(p.errors, err)
		return nil
	}

	l.Value = value

	return l
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	l := &ast.FloatLiteral{Token: p.currToken}

	value, err := strconv.ParseFloat(p.currToken.Word, 64)
	if err != nil {
		err = fmt.Errorf("could not parse %q as float. %s",
			p.currToken.Word, err)
		p.errors = append(p.errors, err)
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

func (p *Parser) parseNullLiteral() ast.Expression {
	return &ast.NullLiteral{
		Token: p.currToken,
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

func (p *Parser) hasToken(tokenType token.TokenType) bool {
	return p.currToken.Type == tokenType
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
	err := fmt.Errorf("%s %s", p.nextToken.Loc, message)
	p.errors = append(p.errors, err)
}
