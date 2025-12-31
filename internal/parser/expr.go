package parser

import (
	"fmt"
	"kyra/internal/lexer"
)

// ---------------------------
// Pratt parser precedence
// ---------------------------

var precedences = map[lexer.TokenType]int{
	lexer.OR:       1,
	lexer.AND:      2,
	lexer.EQ:       3,
	lexer.NEQ:      3,
	lexer.LT:       4,
	lexer.GT:       4,
	lexer.LE:       4,
	lexer.GE:       4,
	lexer.PLUS:     5,
	lexer.MINUS:    5,
	lexer.STAR:     6,
	lexer.SLASH:    6,
	lexer.PERCENT:  6,
	lexer.ASSIGN:   0, // assignment is right-associative
	lexer.PLUS_EQ:  0,
	lexer.MINUS_EQ: 0,
	lexer.MUL_EQ:   0,
	lexer.DIV_EQ:   0,
}

func getPrecedence(tok lexer.Token) int {
	if p, ok := precedences[tok.Type]; ok {
		return p
	}
	return -1
}

// ---------------------------
// Entry
// ---------------------------

func parseExpr(p *Parser, minPrec int) Expr {
	left := parsePrefix(p)

	for {
		tok := p.peek()
		prec := getPrecedence(tok)

		if prec < minPrec {
			break
		}

		// Assignment is right-associative
		if tok.Type == lexer.ASSIGN ||
			tok.Type == lexer.PLUS_EQ ||
			tok.Type == lexer.MINUS_EQ ||
			tok.Type == lexer.MUL_EQ ||
			tok.Type == lexer.DIV_EQ {

			left = parseAssignment(p, left)
			continue
		}

		p.next() // consume operator
		right := parseExpr(p, prec+1)

		left = &BinaryExpr{
			Left:  left,
			Op:    tok.Lexeme,
			Right: right,
		}
	}

	return left
}

// ---------------------------
// Prefix parsing
// ---------------------------

func parsePrefix(p *Parser) Expr {
	tok := p.peek()

	switch tok.Type {

	case lexer.IDENT:
		return parseIdentifierOrCallOrMember(p)

	case lexer.NUMBER:
		p.next()
		return &NumberExpr{Value: tok.Lexeme}

	case lexer.STRING:
		p.next()
		return &StringExpr{Value: tok.Lexeme}

	case lexer.K_TRUE:
		p.next()
		return &BoolExpr{Value: true}

	case lexer.K_FALSE:
		p.next()
		return &BoolExpr{Value: false}

	case lexer.MINUS, lexer.BANG:
		p.next()
		right := parsePrefix(p)
		return &UnaryExpr{
			Op:   tok.Lexeme,
			Expr: right,
		}

	case lexer.LPAREN:
		p.next()
		expr := p.parseExpression()
		p.expect(lexer.RPAREN, "expected ')'")
		return &ParenExpr{Expr: expr}

	default:
		panic(fmt.Sprintf("Unexpected token in expression: %s (%s)", tok.Type, tok.Lexeme))
	}
}

// ---------------------------
// Identifier / Call / Member
// ---------------------------

func parseIdentifierOrCallOrMember(p *Parser) Expr {
	ident := p.expect(lexer.IDENT, "identifier").Lexeme
	var expr Expr = &IdentExpr{Name: ident}

	for {
		switch p.peek().Type {

		case lexer.LPAREN:
			expr = parseCall(p, expr)

		case lexer.DOT:
			p.next()
			name := p.expect(lexer.IDENT, "member name").Lexeme
			expr = &MemberExpr{
				Object: expr,
				Name:   name,
			}

		default:
			return expr
		}
	}
}

func parseCall(p *Parser, callee Expr) Expr {
	p.expect(lexer.LPAREN, "call")

	args := []Expr{}

	if p.peek().Type != lexer.RPAREN {
		for {
			arg := p.parseExpression()
			args = append(args, arg)

			if !p.match(lexer.COMMA) {
				break
			}
		}
	}

	p.expect(lexer.RPAREN, "expected ')' after call arguments")

	return &CallExpr{
		Callee: callee,
		Args:   args,
	}
}

// ---------------------------
// Assignment
// ---------------------------

func parseAssignment(p *Parser, left Expr) Expr {
	tok := p.next()

	ident, ok := left.(*IdentExpr)
	if !ok {
		panic("Left side of assignment must be an identifier")
	}

	value := p.parseExpression()

	return &AssignExpr{
		Name: ident.Name,
		Expr: value,
	}
}
