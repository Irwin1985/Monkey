package ast

import (
	"bytes"
	"monkey/token"
	"strings"
)

// --------------------------------------------------------------------------------------------------------	//
// Se crea una interface padre llamada Node que contiene un método público TokenLiteral()
// La interface Statement embebe la interface Node por lo que todo tipo struct derivado
// deberá implementar los métodos tanto de Node como de Statement.
// La interface Expression embebe la interface Node por lo que todo tipo struct derivado
// deberá implementar los métodos tanto de Node como de Expression.

// Los nodos de tipo statement implementarán a Statement.
// Los nodos de tipo expression implementarán a Expression.
// --------------------------------------------------------------------------------------------------------	//

// Interface Node => Es la interface raíz de todas las construcciones del lenguaje a crear.
type Node interface {
	// Sirve para ver el token al que pertenece el Nodo.
	TokenLiteral() string
	// Sirve para ver el AST generado por el Nodo.
	String() string
}

// Interface Statement => Implementa implicitamente a la interface Node.
type Statement interface {
	Node
	statementNode()
}

// Interface Expression => Interface para todos los Nodos de tipo Expression.
// Implementa implicitamente la interface Node.
type Expression interface {
	Node
	expressionNode()
}

// Primera implementación de la interface Node.
// este nodo Program será el nodo raíz del AST que producirá el Parser.
// Un programa está compuesto por una o más sentencias, cada una en forma
// de nodo de tipo Statement que será almacenado en el slice []Statement.

// Estructura Program => contendrá todas las sentencias que forman un script o programa.
// su gramática EBNF es:
// program 		= statements
// statements 	= statement {statement}
type Program struct {
	// Lista de todos los AST's que deriven de la interface Statement. Para conocer su tipo real
	// primero hay que castear su interface y luego su tipo. Ejemplo del AST Identifier:
	// exprStmt := program.Statements[0].(*ast.ExpressionStatement) y luego
	// identifier = exprStmt.Expression.(*ast.Identifier) y finalmente:
	// identifier.Name, identifier.TokenLiteral().
	// Identifier es un Tipo de dato que deriva de la interface Node<-Expression por lo tanto
	// es un dato contenido dentro del campo Expression del tipo ExpressionStatement.
	// program.Statements[0] hace referencia al dato AST ExpressionStatement.
	Statements []Statement
}

// Implementa la función String() de la interface Node.
func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// Implementa la función TokenLiteral() que cumple con Node.
func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

// Estructura LetStatement => se encargará de crear el AST para la gramática:
// letStatement = 'let' identifier expression
type LetStatement struct {
	// El token asociado: token.Type = LET, token.Literal = 'let'
	Token token.Token
	// Puntero al AST Identifier
	Name *Identifier
	// AST que implementa la interface Expression.
	Value Expression
}

// Cumple con la interface Statement.
func (ls *LetStatement) statementNode() {}

// Cumple con la interface Node.
func (ls *LetStatement) TokenLiteral() string {
	return ls.Token.Literal
}

// Implementa la función String() de la interface Node.
func (ls *LetStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")
	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

// Estructura Identifier => se encargará de crear el AST para la gramática:
// identifier = string
type Identifier struct {
	// El token asociado: token.Type = IDENT, token.Literal = 'any string'
	Token token.Token
	// string literal del indentificador.
	Value string
}

// Cumple con la interface Expression.
func (i *Identifier) expressionNode() {}

// Cumple con la interface Node.
func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
}

// Implementa la función String() de la interface Node.
func (i *Identifier) String() string {
	return i.Value
}

// Estructura ReturnStatement => se encargará de crear el AST para la gramática:
// ReturnStatement = 'return' expression ';'
type ReturnStatement struct {
	// El token asociado: token.Type = RETURN, token.Literal = 'return'
	Token       token.Token
	ReturnValue Expression
}

// Cumple con la interface Statement.
func (rs *ReturnStatement) statementNode() {}

// Cumple con la interface Node.
func (rs *ReturnStatement) TokenLiteral() string {
	return rs.Token.Literal
}

// Implementa la función String() de la interface Node.
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral() + " ")
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}
	out.WriteString(";")
	return out.String()
}

// Estructura ExpressionStatem-ent => se encargará de crear el AST para la gramática:
// ExpressionStatement 	= BooleanExpression
// BooleanExpression	= BooleanTerm 'and' BooleanTerm
// BooleanTerm			= BooleanFactor 'or' BooleanFactor
// BooleanFactor		= TRUE | FALSE | Relation
// Relation				= ArithmeticExpression [relOp ArithmeticExpression]
// relOp				= '<' | '>' | '<=' | '>=' | '!=' | '=='
type ExpressionStatement struct {
	// El token asociado a la expresión.
	Token token.Token
	// Contiene un AST que derive de la interface Expression. Para conocer el tipo real primero hay
	// que castearlo a su tipo original. Ejemplo: es.Expression.(*LetStatement) convierte el campo
	// Expression a un AST de tipo LetStatement.
	Expression Expression
}

// Cumple con la interface Statement.
func (es *ExpressionStatement) statementNode() {}

// Cumple con la interface Node.
func (es *ExpressionStatement) TokenLiteral() string {
	return es.Token.Literal
}

// Implementa el método String()
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

// Cumple con la interface Expression.
func (il *IntegerLiteral) expressionNode() {}

// Cumple con la interface Node.
func (il *IntegerLiteral) TokenLiteral() string {
	return il.Token.Literal
}

// Implementa el método String
func (il *IntegerLiteral) String() string {
	return il.Token.Literal
}

// PrefixExpression es el operador PREFIJO que por naturaleza
// posee un operando a la derecha de tipo Expression.
// Ejemplo: -5, !false
// donde '-' y '!' son los operadores PREFIJO.
type PrefixExpression struct {
	// El token asociado
	Token token.Token
	// El operador asociado
	Operator string
	// El operando es cualquier tipo que derive de Expression
	Right Expression
}

// Cumple con la interface Expression.
func (pe *PrefixExpression) expressionNode() {}

// Cumple con la interface Node.
func (pe *PrefixExpression) TokenLiteral() string {
	return pe.Token.Literal
}

// Implementa el método String
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

// ast.InfixExpression
type InfixExpression struct {
	Token    token.Token
	Operator string
	Left     Expression
	Right    Expression
}

// Cumple con la interface Expression
func (ie *InfixExpression) expressionNode() {}

// Cumple con la interface Node
func (ie *InfixExpression) TokenLiteral() string {
	return ie.Token.Literal
}

// Cumple con la interface Node
func (ie *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")
	return out.String()
}

// ast.Boolean
type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) String() string       { return b.Token.Literal }

// ast.IfExpression
type IfExpression struct {
	Token       token.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var out bytes.Buffer
	out.WriteString("if")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Consequence.String())
	if ie.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(ie.Alternative.String())
	}
	return out.String()
}

// BlockStatement
type BlockStatement struct {
	Token      token.Token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer

	for _, stmt := range bs.Statements {
		out.WriteString(stmt.String())
	}

	return out.String()
}

// FunctionLiteral
type FunctionLiteral struct {
	// El token literal es 'fn'
	Token      token.Token
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer
	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fl.Body.String())
	return out.String()
}

// CallExpression -> Llamadas a funciones
type CallExpression struct {
	Token     token.Token
	Function  Expression // identificador o función literal
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer
	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}

// String
type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return sl.Token.Literal }

// Array Literal
type ArrayLiteral struct {
	Token    token.Token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.TokenLiteral() }
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer
	elements := []string{}
	for _, el := range al.Elements {
		elements = append(elements, el.String())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}

// Array Index Operator Expression
type IndexExpression struct {
	Token token.Token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")
	return out.String()

}

// Hash Maps
type HashLiteral struct {
	Token token.Token
	Pairs map[Expression]Expression
}

func (hl *HashLiteral) expressionNode()      {}
func (hl *HashLiteral) TokenLiteral() string { return hl.Token.Literal }
func (hl *HashLiteral) String() string {
	var out bytes.Buffer
	pairs := []string{}
	for key, value := range hl.Pairs {
		pairs = append(pairs, key.String()+":"+value.String())
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}
