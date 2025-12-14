package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/VzoelFox/morphlang/pkg/lexer"
)

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
)

var precedences = map[lexer.TokenType]int{
	lexer.EQ:       EQUALS,
	lexer.NOT_EQ:   EQUALS,
	lexer.LT:       LESSGREATER,
	lexer.GT:       LESSGREATER,
	lexer.LTE:      LESSGREATER,
	lexer.GTE:      LESSGREATER,
	lexer.PLUS:     SUM,
	lexer.MINUS:    SUM,
	lexer.SLASH:    PRODUCT,
	lexer.ASTERISK: PRODUCT,
	lexer.LPAREN:   CALL,
}

type (
	prefixParseFn func() Expression
	infixParseFn  func(Expression) Expression
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  lexer.Token
	peekToken lexer.Token

	prefixParseFns map[lexer.TokenType]prefixParseFn
	infixParseFns  map[lexer.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[lexer.TokenType]prefixParseFn)
	p.registerPrefix(lexer.IDENT, p.parseIdentifier)
	p.registerPrefix(lexer.INT, p.parseIntegerLiteral)
	p.registerPrefix(lexer.STRING, p.parseStringLiteral)
	p.registerPrefix(lexer.INTERP_START, p.parseStringLiteral)
	p.registerPrefix(lexer.BENAR, p.parseBoolean)
	p.registerPrefix(lexer.SALAH, p.parseBoolean)
	p.registerPrefix(lexer.BANG, p.parsePrefixExpression)
	p.registerPrefix(lexer.MINUS, p.parsePrefixExpression)
	p.registerPrefix(lexer.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(lexer.JIKA, p.parseIfExpression)
	p.registerPrefix(lexer.SELAMA, p.parseWhileExpression)
	p.registerPrefix(lexer.FUNGSI, p.parseFunctionLiteral)

	p.infixParseFns = make(map[lexer.TokenType]infixParseFn)
	p.registerInfix(lexer.PLUS, p.parseInfixExpression)
	p.registerInfix(lexer.MINUS, p.parseInfixExpression)
	p.registerInfix(lexer.SLASH, p.parseInfixExpression)
	p.registerInfix(lexer.ASTERISK, p.parseInfixExpression)
	p.registerInfix(lexer.EQ, p.parseInfixExpression)
	p.registerInfix(lexer.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(lexer.LT, p.parseInfixExpression)
	p.registerInfix(lexer.GT, p.parseInfixExpression)
	p.registerInfix(lexer.LTE, p.parseInfixExpression)
	p.registerInfix(lexer.GTE, p.parseInfixExpression)
	p.registerInfix(lexer.LPAREN, p.parseCallExpression)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) addDetailedError(tok lexer.Token, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	lineContent := p.getLineContent(tok.Line)

	pointer := ""
	for i := 0; i < tok.Column-1; i++ {
		pointer += " "
	}
	pointer += "^"

	formatted := fmt.Sprintf("Error [%d:%d]: %s\n  %d | %s\n       %s",
		tok.Line, tok.Column, msg, tok.Line, lineContent, pointer)

	p.errors = append(p.errors, formatted)
}

func (p *Parser) getLineContent(line int) string {
	lines := strings.Split(p.l.Input(), "\n")
	if line >= 1 && line <= len(lines) {
		return lines[line-1]
	}
	return ""
}

func (p *Parser) peekError(t lexer.TokenType) {
	p.addDetailedError(p.peekToken, "expected next token to be %s, got %s instead", t, p.peekToken.Type)
}

func (p *Parser) curError(t lexer.TokenType) {
	p.addDetailedError(p.curToken, "expected token to be %s, got %s instead", t, p.curToken.Type)
}

func (p *Parser) registerPrefix(tokenType lexer.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType lexer.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) ParseProgram() *Program {
	program := &Program{}
	program.Statements = []Statement{}

	for p.curToken.Type != lexer.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() Statement {
	switch p.curToken.Type {
	case lexer.KEMBALIKAN:
		return p.parseReturnStatement()
	case lexer.IDENT:
		if p.peekTokenIs(lexer.ASSIGN) {
			return p.parseAssignmentStatement()
		}
		return p.parseExpressionStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseAssignmentStatement() *AssignmentStatement {
	stmt := &AssignmentStatement{Token: p.curToken}
	stmt.Name = &Identifier{Token: p.curToken, Value: p.curToken.Literal}

	p.nextToken() // eat IDENT
	p.nextToken() // eat =

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ReturnStatement {
	stmt := &ReturnStatement{Token: p.curToken}

	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() *ExpressionStatement {
	stmt := &ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(lexer.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
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

func (p *Parser) parseIdentifier() Expression {
	return &Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() Expression {
	lit := &IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		p.addDetailedError(p.curToken, "could not parse %q as integer", p.curToken.Literal)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() Expression {
	// Optimization: If simple string (starts with STRING and no following parts)
	if p.curTokenIs(lexer.STRING) && p.peekToken.Type != lexer.INTERP_START && p.peekToken.Type != lexer.STRING {
		return &StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
	}

	is := &InterpolatedString{Token: p.curToken, Parts: []Expression{}}

	processToken := func() bool {
		if p.curTokenIs(lexer.STRING) {
			is.Parts = append(is.Parts, &StringLiteral{Token: p.curToken, Value: p.curToken.Literal})
			return true
		}
		if p.curTokenIs(lexer.INTERP_START) {
			p.nextToken() // move to expression
			expr := p.parseExpression(LOWEST)
			is.Parts = append(is.Parts, expr)
			if !p.expectPeek(lexer.RBRACE) {
				return false
			}
			return true
		}
		return false
	}

	if !processToken() {
		return nil
	}

	for {
		if p.peekTokenIs(lexer.INTERP_START) || p.peekTokenIs(lexer.STRING) {
			p.nextToken()
			if !processToken() {
				return nil
			}
		} else {
			break
		}
	}
	return is
}

func (p *Parser) parseBoolean() Expression {
	return &BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(lexer.BENAR)}
}

func (p *Parser) parsePrefixExpression() Expression {
	expression := &PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseWhileExpression() Expression {
	expression := &WhileExpression{Token: p.curToken}
	p.nextToken() // eat selama

	expression.Condition = p.parseExpression(LOWEST)

	p.nextToken() // move to block start

	expression.Body = p.parseBlockStatement()

	if p.curTokenIs(lexer.AKHIR) {
		// Consumed by loop? No, loop breaks.
		// We verify it is AKHIR, but do NOT consume it past this node.
	} else {
		p.curError(lexer.AKHIR)
	}

	return expression
}

func (p *Parser) parseInfixExpression(left Expression) Expression {
	expression := &InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	// Strict Whitespace Check
	if isBinaryOp(p.curToken.Type) {
		if !p.curToken.HasLeadingSpace {
			p.addDetailedError(p.curToken, "Binary operator '%s' requires space before it", p.curToken.Literal)
		}
		if !p.peekToken.HasLeadingSpace {
			p.addDetailedError(p.curToken, "Binary operator '%s' requires space after it", p.curToken.Literal)
		}
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func isBinaryOp(t lexer.TokenType) bool {
	switch t {
	case lexer.PLUS, lexer.MINUS, lexer.SLASH, lexer.ASTERISK,
		lexer.EQ, lexer.NOT_EQ, lexer.LT, lexer.GT, lexer.LTE, lexer.GTE:
		return true
	}
	return false
}

func (p *Parser) parseGroupedExpression() Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseIfExpression() Expression {
	expression := &IfExpression{Token: p.curToken}
	// curToken is JIKA or ATAU_JIKA
	p.nextToken() // eat jika/atau_jika

	expression.Condition = p.parseExpression(LOWEST)

	// Advance to start of block
	p.nextToken()

	expression.Consequence = p.parseBlockStatement()

	if p.curTokenIs(lexer.LAINNYA) {
		p.nextToken() // eat lainnya
		expression.Alternative = p.parseBlockStatement()

		// Expect AKHIR after lainnya block
		if p.curTokenIs(lexer.AKHIR) {
			// Do not consume
		} else {
			p.curError(lexer.AKHIR)
		}
	} else if p.curTokenIs(lexer.ATAU_JIKA) {
		// chain
		child := p.parseIfExpression()
		expression.Alternative = &BlockStatement{
			Statements: []Statement{
				&ExpressionStatement{Expression: child},
			},
		}
		// child parseIfExpression finishes at AKHIR.
	} else {
		// Expect AKHIR
		if p.curTokenIs(lexer.AKHIR) {
			// Do not consume
		} else {
			p.curError(lexer.AKHIR)
		}
	}

	return expression
}

func (p *Parser) parseFunctionLiteral() Expression {
	lit := &FunctionLiteral{Token: p.curToken}

	if p.peekTokenIs(lexer.IDENT) {
		p.nextToken()
		lit.Name = p.curToken.Literal
	}

	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	p.nextToken() // eat ) to move to block

	lit.Body = p.parseBlockStatement()

	if p.curTokenIs(lexer.AKHIR) {
		// Do not consume
	} else {
		p.curError(lexer.AKHIR)
	}

	return lit
}

func (p *Parser) parseFunctionParameters() []*Identifier {
	identifiers := []*Identifier{}

	if p.peekTokenIs(lexer.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	ident := &Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	return identifiers
}

func (p *Parser) parseCallExpression(function Expression) Expression {
	exp := &CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseCallArguments()
	return exp
}

func (p *Parser) parseCallArguments() []Expression {
	args := []Expression{}

	if p.peekTokenIs(lexer.RPAREN) {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	return args
}

func (p *Parser) parseBlockStatement() *BlockStatement {
	block := &BlockStatement{Token: p.curToken}
	block.Statements = []Statement{}

	for !p.curTokenIs(lexer.AKHIR) && !p.curTokenIs(lexer.LAINNYA) && !p.curTokenIs(lexer.ATAU_JIKA) && !p.curTokenIs(lexer.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) noPrefixParseFnError(t lexer.TokenType) {
	p.addDetailedError(p.curToken, "no prefix parse function for %s found", t)
}
