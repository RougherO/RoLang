package ast

import (
	"RoLang/token"

	"fmt"
)

type (
	Node interface {
		TokenWord() string
		String() string
	}

	Statement interface {
		Node
		Statement()
	}

	Expression interface {
		Node
		Expression()
	}
)

type (
	Program struct {
		Statements []Statement
	}

	BlockStatement struct {
		Token      token.Token
		Statements []Statement
	}

	FunctionStatement struct {
		Token token.Token
		Ident *Identifier
		Value *FunctionLiteral
	}

	LetStatement struct {
		Token     token.Token
		Ident     *Identifier
		InitValue Expression
	}

	ReturnStatement struct {
		Token       token.Token
		ReturnValue Expression
	}

	ExpressionStatement struct {
		Token      token.Token
		Expression Expression
	}

	IfStatement struct {
		Token     token.Token
		Condition Expression
		Then      *BlockStatement
		Else      Statement // block or expression statement
	}

	PrefixExpression struct {
		Token    token.Token
		Operator string
		Right    Expression
	}

	InfixExpression struct {
		Token    token.Token
		Operator string
		Left     Expression
		Right    Expression
	}

	CallExpression struct {
		Token     token.Token // '(' token
		Callee    Expression
		Arguments []Expression
	}

	Identifier struct {
		Token token.Token
		Value string
	}

	FunctionLiteral struct {
		Token      token.Token
		Parameters []*Identifier
		Body       *BlockStatement
	}

	IntegerLiteral struct {
		Token token.Token
		Value int64
	}

	FloatLiteral struct {
		Token token.Token
		Value float64
	}

	BoolLiteral struct {
		Token token.Token
		Value bool
	}
)

func (p *Program) TokenWord() string {
	if len(p.Statements) > 0 {
		var out string
		for _, stmt := range p.Statements {
			out += stmt.String()
		}
		return out
	} else {
		return ""
	}
}

func (p *Program) String() string {
	var out string

	for _, s := range p.Statements {
		out += s.String()
	}

	return out
}

func (bs *BlockStatement) TokenWord() string {
	return bs.Token.Word
}

func (bs *BlockStatement) String() string {
	var out string
	out += "{ "
	for _, stmt := range bs.Statements {
		out += stmt.String()
	}
	out += " }"
	return out
}

func (bs *BlockStatement) Statement() {}

func (ls *LetStatement) TokenWord() string {
	return ls.Token.Word
}

func (fs *FunctionStatement) TokenWord() string {
	return fs.Token.Word
}

func (fs *FunctionStatement) String() string {
	var params string
	for i, param := range fs.Value.Parameters {
		if i == 0 {
			params += param.String()
		} else {
			params += ", " + param.String()
		}
	}

	return fmt.Sprintf("fn %s(%s) %s", fs.Ident, params, fs.Value.Body)
}

func (fs *FunctionStatement) Statement() {}

func (ls *LetStatement) String() string {
	if ls.InitValue != nil {
		return fmt.Sprintf("let %s = %s;", ls.Ident.Value, ls.InitValue)
	}

	return fmt.Sprintf("let %s", ls.Ident.Value)
}

func (ls *LetStatement) Statement() {}

func (rs *ReturnStatement) TokenWord() string {
	return rs.Token.Word
}

func (rs *ReturnStatement) String() string {
	if rs.ReturnValue != nil {
		return fmt.Sprintf("return %s;", rs.ReturnValue)
	}

	return "return;"
}

func (rs *ReturnStatement) Statement() {}

func (es *ExpressionStatement) TokenWord() string {
	return es.Token.Word
}

func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}

	return ""
}

func (es *ExpressionStatement) Statement() {}

func (is *IfStatement) TokenWord() string {
	return is.Token.Word
}

func (is *IfStatement) String() string {
	var out string
	out += fmt.Sprintf("if %s", is.Then)

	if is.Else != nil {
		out += fmt.Sprintf("else %s", is.Else)
	}

	return out
}

func (is *IfStatement) Statement() {}

func (ie *InfixExpression) TokenWord() string {
	return ie.Token.Word
}

func (ie *InfixExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", ie.Left, ie.Operator, ie.Right)
}

func (ie *InfixExpression) Expression() {}

func (pe *PrefixExpression) TokenWord() string {
	return pe.Token.Word
}

func (pe *PrefixExpression) String() string {
	return fmt.Sprintf("(%s%s)", pe.Operator, pe.Right)
}

func (pe *PrefixExpression) Expression() {}

func (id *Identifier) TokenWord() string {
	return id.Token.Word
}

func (id *Identifier) String() string {
	return id.Value
}

func (id *Identifier) Expression() {}

func (ce *CallExpression) TokenWord() string {
	return ce.Token.Word
}

func (ce *CallExpression) String() string {
	var args string
	for i, arg := range ce.Arguments {
		if i == 0 {
			args += arg.String()
		} else {
			args += ", " + arg.String()
		}
	}

	return fmt.Sprintf("%s(%s)", ce.Callee, args)
}

func (ce *CallExpression) Expression() {}

func (fl *FunctionLiteral) TokenWord() string {
	return fl.Token.Word
}

func (fl *FunctionLiteral) String() string {
	var params string
	for i, param := range fl.Parameters {
		if i == 0 {
			params += param.String()
		} else {
			params += ", " + param.String()
		}
	}
	return fmt.Sprintf("fn (%s) %s", params, fl.Body)
}

func (fl *FunctionLiteral) Expression() {}

func (il *IntegerLiteral) TokenWord() string {
	return il.Token.Word
}

func (il *IntegerLiteral) String() string {
	return il.TokenWord()
}

func (il *IntegerLiteral) Expression() {}

func (fl *FloatLiteral) TokenWord() string {
	return fl.Token.Word
}

func (fl *FloatLiteral) String() string {
	return fl.TokenWord()
}

func (fl *FloatLiteral) Expression() {}

func (bl *BoolLiteral) TokenWord() string {
	return bl.Token.Word
}

func (bl *BoolLiteral) String() string {
	return bl.Token.Word
}

func (bl *BoolLiteral) Expression() {}
