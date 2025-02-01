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

	Identifier struct {
		Token token.Token
		Value string
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
		return p.Statements[0].TokenWord()
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

func (ls *LetStatement) TokenWord() string {
	return ls.Token.Word
}

func (ls *LetStatement) String() string {
	if ls.InitValue != nil {
		return fmt.Sprintf("let %s = %s;", ls.Ident.Value, ls.InitValue)
	}

	return fmt.Sprintf("%s %s", ls.TokenWord(), ls.Ident.Value)
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
