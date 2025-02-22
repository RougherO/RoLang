package ast

import (
	"RoLang/token"

	"fmt"
)

type (
	Node interface {
		Location() token.SrcLoc
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

	LoopStatement struct {
		Token     token.Token // `loop` keyword
		Condition Expression
		Body      *BlockStatement
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

	IndexExpression struct {
		Token token.Token // '[' token
		Left  Expression
		Index Expression
	}

	Identifier struct {
		Token token.Token
		Value string
	}

	AssignExpression struct {
		Token token.Token // '=' token
		Left  Expression
		Right Expression
	}

	FunctionLiteral struct {
		Token      token.Token
		Parameters []*Identifier
		Body       *BlockStatement
	}

	ArrayLiteral struct {
		Token    token.Token // '[' token
		Elements []Expression
	}

	MapElement struct {
		Key   Expression
		Value Expression
	}

	MapLiteral struct {
		Token    token.Token // '{' token
		Elements []MapElement
	}

	StringLiteral struct {
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

	NullLiteral struct {
		Token token.Token
	}
)

func (p *Program) String() string {
	var out string

	for _, s := range p.Statements {
		out += s.String()
	}

	return out
}

func (bs *BlockStatement) Location() token.SrcLoc {
	return bs.Token.Loc
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

func (ls *LetStatement) String() string {
	if ls.InitValue != nil {
		return fmt.Sprintf("let %s = %s;", ls.Ident.Value, ls.InitValue)
	}

	return fmt.Sprintf("let %s", ls.Ident.Value)
}

func (ls *LetStatement) Location() token.SrcLoc {
	return ls.Token.Loc
}

func (ls *LetStatement) Statement() {}

func (ls *LoopStatement) String() string {
	out := "loop "

	if ls.Condition != nil {
		out += ls.Condition.String()
	}

	out += ls.Body.String()

	return out
}

func (ls *LoopStatement) Location() token.SrcLoc {
	return ls.Token.Loc
}

func (ls *LoopStatement) Statement() {}

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

func (fs *FunctionStatement) Location() token.SrcLoc {
	return fs.Token.Loc
}

func (fs *FunctionStatement) Statement() {}

func (rs *ReturnStatement) String() string {
	if rs.ReturnValue != nil {
		return fmt.Sprintf("return %s;", rs.ReturnValue)
	}

	return "return;"
}

func (rs *ReturnStatement) Location() token.SrcLoc {
	return rs.Token.Loc
}

func (rs *ReturnStatement) Statement() {}

func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}

	return ""
}

func (es *ExpressionStatement) Location() token.SrcLoc {
	return es.Token.Loc
}

func (es *ExpressionStatement) Statement() {}

func (is *IfStatement) String() string {
	var out string
	out += fmt.Sprintf("if %s", is.Then)

	if is.Else != nil {
		out += fmt.Sprintf("else %s", is.Else)
	}

	return out
}

func (is *IfStatement) Location() token.SrcLoc {
	return is.Token.Loc
}

func (is *IfStatement) Statement() {}

func (ie *InfixExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", ie.Left, ie.Operator, ie.Right)
}

func (ie *InfixExpression) Location() token.SrcLoc {
	return ie.Token.Loc
}

func (ie *InfixExpression) Expression() {}

func (pe *PrefixExpression) String() string {
	return fmt.Sprintf("(%s%s)", pe.Operator, pe.Right)
}

func (pe *PrefixExpression) Location() token.SrcLoc {
	return pe.Token.Loc
}

func (pe *PrefixExpression) Expression() {}

func (id *Identifier) String() string {
	return id.Value
}

func (id *Identifier) Location() token.SrcLoc {
	return id.Token.Loc
}

func (id *Identifier) Expression() {}

func (ae *AssignExpression) String() string {
	return fmt.Sprintf("(%s = %s)", ae.Left, ae.Right)
}

func (ae *AssignExpression) Location() token.SrcLoc {
	return ae.Token.Loc
}

func (ae *AssignExpression) Expression() {}

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

func (ce *CallExpression) Location() token.SrcLoc {
	return ce.Token.Loc
}

func (ce *CallExpression) Expression() {}

func (ie *IndexExpression) String() string {
	return fmt.Sprintf("(%s[%s])", ie.Left, ie.Index)
}

func (ie *IndexExpression) Location() token.SrcLoc {
	return ie.Token.Loc
}

func (ie *IndexExpression) Expression() {}

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

func (fl *FunctionLiteral) Location() token.SrcLoc {
	return fl.Token.Loc
}

func (fl *FunctionLiteral) Expression() {}

func (il *IntegerLiteral) String() string {
	return il.Token.Word
}

func (il *IntegerLiteral) Location() token.SrcLoc {
	return il.Token.Loc
}

func (il *IntegerLiteral) Expression() {}

func (fl *FloatLiteral) String() string {
	return fl.Token.Word
}

func (fl *FloatLiteral) Location() token.SrcLoc {
	return fl.Token.Loc
}

func (fl *FloatLiteral) Expression() {}

func (al *ArrayLiteral) String() string {
	var out string
	out += "["
	for i, elem := range al.Elements {
		if i == 0 {
			out += elem.String()
		} else {
			out += ", " + elem.String()
		}
	}
	out += "]"

	return out
}

func (al *ArrayLiteral) Location() token.SrcLoc {
	return al.Token.Loc
}

func (al *ArrayLiteral) Expression() {}

func (me *MapElement) String() string {
	return fmt.Sprintf("%s:%s", me.Key, me.Value)
}

func (ml *MapLiteral) String() string {
	var out string
	out += "{"
	for i, elem := range ml.Elements {
		if i == 0 {
			out += elem.String()
		} else {
			out += ", " + elem.String()
		}
	}
	out += "}"

	return out
}

func (ml *MapLiteral) Location() token.SrcLoc {
	return ml.Token.Loc
}

func (ml *MapLiteral) Expression() {}

func (sl *StringLiteral) String() string {
	return sl.Token.Word
}

func (sl *StringLiteral) Location() token.SrcLoc {
	return sl.Token.Loc
}

func (sl *StringLiteral) Expression() {}

func (bl *BoolLiteral) String() string {
	return bl.Token.Word
}

func (bl *BoolLiteral) Location() token.SrcLoc {
	return bl.Token.Loc
}

func (bl *BoolLiteral) Expression() {}

func (nl *NullLiteral) String() string {
	return nl.Token.Word
}

func (nl *NullLiteral) Location() token.SrcLoc {
	return nl.Token.Loc
}

func (nl *NullLiteral) Expression() {}
