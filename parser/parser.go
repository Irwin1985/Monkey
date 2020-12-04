package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
	"strconv"
)

const ( // Listado de constantes que definen el orden de precedencia de los operadores
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // < o >
	SUM         // +
	PRODUCT     // *
	PREFIX      // -x o !x
	CALL        // myFunction(X)
)

type Parser struct {
	// Instancia del Lexer.
	l *lexer.Lexer
	// Es el token actual.
	curToken token.Token
	// Es el siguiente token.
	peekToken token.Token
	// Lista de todos los errores encontrados por el Parser.
	errors []string
	// Listado de Tokens de tipo PREFIJO asociados a la función prefixParseFn.
	prefixParseFns map[token.TokenType]prefixParseFn
	// Listado de Tokens de tipo INFIJO asociados a la función infixParseFn.
	infixParseFns map[token.TokenType]infixParseFn
}

type (
	// Se llama cuando el token se encuentre en la posición PREFIJO.
	// Este es un tipo de dato que en lugar de tener un tipo nativo
	// como: bool, int, string, tiene una función.
	// esto se conoce como function first class citizen.
	prefixParseFn func() ast.Expression
	// Se llama cuando el token se encuentre en la posición INFIJO.
	infixParseFn func(ast.Expression) ast.Expression
)

// precedences es una tabla de precedencias que asocia los tipos de token con su orden
// de precedencia con respecto a los demás.
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

// registerPrefix es una función helper para registrar el tipo de token PREFIJO
// junto con su respectiva función de análisis prefixParseFn.
func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// registerInfix es una función helper para registrar el tipo de token INFIJO
// junto con su respectiva función de análisis infixParseFn.
func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// Este método revisa si el siguiente token tiene algún registro asociado
// en la tabla de precedencias. De lo contrario devuelve LOWEST.
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

// Y este método consulta si la tabla tiene un orden asociado para el token actual
// de lo contrario devuelve la precedencia más baja que es LOWEST.
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

// Crea una instancia de Parser
func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}
	// Registramos los tokens de tipo PREFIJO.

	// Primero inicializamos el map prefixParseFns
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	// Inicializamos el map para las operaciones INFIJO o BINARIAS.
	p.infixParseFns = make(map[token.TokenType]infixParseFn)

	// Registramos el token IDENT para identificadores.
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	// Registramos el token INT para enteros literales.
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	// Registramos el token BANG para las negaciones booleanas.
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	// Registramos el token MINUS para el operador PREFIJO '-'.
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	// Registramos el token TRUE
	p.registerPrefix(token.TRUE, p.parseBoolean)
	// Registramos el token FALSE
	p.registerPrefix(token.FALSE, p.parseBoolean)
	// Registramos el token LPAREN
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	// Registramos el token IF
	p.registerPrefix(token.IF, p.parseIfExpression)
	// Registramos el token FUNCTION
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	// Registramos el token STRING
	p.registerPrefix(token.STRING, p.parseStringLiteral)

	// Procedemos a registrar las operaciones INFIJO o BINARIAS.
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	// Registramos las llamadas a las funciones.
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	// leemos 2 tokens, uno para el actual y el otro para el siguiente.
	p.nextToken()
	p.nextToken()
	return p
}

// Analiza un string literal.
func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

// Analiza las llamadas a las funciones.
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseCallArguments()
	return exp
}

// Alaliza los argumentos.
func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}
	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return args
}

// Función asociada al token.FUNCTION que se encarga de crear
// el AST para ast.FunctionLiteral
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

// Se encarga de analizar los parámetros de una función
func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}
	p.nextToken()
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}

// Función asociada al token.IF que se encarga de crear
// el AST para ast.IfExpression.
// if <condition> <consequence> else <alternative>
func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	expression.Consequence = p.parseBlockStatement()
	// else support.
	if p.peekTokenIs(token.ELSE) {
		p.nextToken()
		if !p.expectPeek(token.LBRACE) {
			return nil
		}
		expression.Alternative = p.parseBlockStatement()
	}
	return expression
}

// Analiza uno o varios bloques de código.
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{} // Inicializa el slice de Statement.
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

// Analiza un identificador. Crea un AST de tipo Expression
// porque los identificadores en Monkey son expresiones también.
// al tipo Identifier le asigna su token correspondiente y su literal.
// ejemplo: Token = token.IDENT, Value = 'foo'
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// Analiza un literal booleano.
func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

// Analiza una expresión agrupada '(' expression ')'
func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken() // Se salta el LPAREN '('
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return exp
}

// Analiza un entero literal. Crea un AST de tipo IntegerLiteral
// porque los enteros en Monkey son expresiones también.
// al tipo IntegerLiteral se le asigna su token correspondiente y su literal.
// el campo token.Literal se convierte a int64 usando strconv.parseInt(str, base, int)
// ejemplo: Token = token.INT, Value = 5
func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
	}
	lit.Value = value
	return lit
}

// Crea un AST de tipo Expression
// llena sus datos con p.curToken
// y llama a parseExpression() para que analice el operador
// de su derecha. Le pasa la precedencia de PREFIJO.
func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

// Este es el famoso analisis para las expresiones INFIJO o BINARIAS.
// veamos de qué se trata :P
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

// El curToken lo iguala a peekToken y avanza peekToken
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// ParseProgram crea el AST para la siguiente gramática:
// program = statement {statement}
func (p *Parser) ParseProgram() *ast.Program {
	programNode := &ast.Program{}
	programNode.Statements = []ast.Statement{}
	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			programNode.Statements = append(programNode.Statements, stmt)
		}
		p.nextToken()
	}
	return programNode
}

// Analiza el Statement actual. Primero verifica de qué tipo es
// y luego llama a su respectivo analizador.
func (p *Parser) parseStatement() ast.Statement {
	// ¿por qué en cada rama hay un return?
	// porque no conocemos el tipo de AST a crear
	// y es mejor devolverlo integro en lugar de
	// crearlo en un tipo generico como Node
	// porque después habrá que sumarle otro casteo.
	// CONCLUSIÓN: a veces las malas prácticas nos simplifican la vida :)
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default: // Asumimos que es un statement porque en Monkey solo hay 2 tipos de statement. (let y return)
		return p.parseExpressionStatement()
	}
}

// Analiza y crea un AST de tipo ast.LetStatement
// usando la siguiente gramática:
// letStatement = 'let' identifier '=' expression
func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

// Analiza y crea un AST de tipo ast.ReturnStatement
// usando la siguiente gramática:
// returnStatement = 'return' expression
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken() // eat the 'return' keyword.
	stmt.ReturnValue = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

// Crea un AST de tipo ExpressionStatement
// una expresión en Monkey puede ser cualquiera de estas
// a + b; x - y; -3 + 2; add(x, y) - sub(x, y); foo; bar - foo;
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	expressionStmt := &ast.ExpressionStatement{Token: p.curToken}
	expressionStmt.Expression = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return expressionStmt
}

// Analiza las expresiones en el lenguaje Monkey.
// Devuelve un AST de tipo Expression
// Busca en el HashMap el token actual y si lo encuentra
// entonces recupera la fución y la invoca.
func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefixFn := p.prefixParseFns[p.curToken.Type]
	if prefixFn == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefixFn()
	for !p.curTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}
	return leftExp
}

// Compara el token recibido con el siguiente token (peekToken).
// Si son iguales avanza el Token, sino entonces registra el error.
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

// Verifica si el token recibido coincide con peekToken.
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// Verifica si el token recibido coincide con currentToken.
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

// Registra un error cuando no existan funciones asociadas al token recibido.
func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

// Retorna la lista de los posibles errores encontrados durante el análisis.
func (p *Parser) Errors() []string {
	return p.errors
}

// Registra el error en la lista de errores.
func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead.", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}
