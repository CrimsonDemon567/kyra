package parser

import "kyra/internal/lexer"

// Parser consumes tokens and produces an AST.
type Parser struct {
    tokens []lexer.Token
    pos    int
}

func New(tokens []lexer.Token) *Parser {
    return &Parser{tokens: tokens}
}

// Parse returns a placeholder AST object.
// You will replace this with a real AST later.
func (p *Parser) Parse() interface{} {
    return nil
}
