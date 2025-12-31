package parser

import (
	"fmt"
	"kyra/internal/lexer"
)

// parseUse parses a full use-statement.
// Examples:
//   use math
//   use kyra/add
//   use sdt/random
//
// Rules:
// - If the first identifier is "sdt", the module is loaded from the stdlib.
// - Otherwise, modules are resolved relative to the project.
func (p *Parser) parseUse() *UseStmt {
	p.expect(lexer.K_USE, "use statement")

	parts := []string{}
	isStdlib := false

	// Detect stdlib prefix: sdt/
	if p.peek().Type == lexer.IDENT && p.peek().Lexeme == "sdt" {
		isStdlib = true
		p.next()
		p.expect(lexer.SLASH, "expected '/' after sdt")
	}

	// Parse module path: a/b/c
	for {
		tok := p.expect(lexer.IDENT, "module name")
		parts = append(parts, tok.Lexeme)

		if !p.match(lexer.SLASH) {
			break
		}
	}

	// Optional newline after use
	for p.match(lexer.NEWLINE) {
	}

	return &UseStmt{
		Path:     parts,
		IsStdlib: isStdlib,
	}
}

// ---------------------------
// Utility: pretty-print module path
// ---------------------------

func (u *UseStmt) String() string {
	p := ""
	for i, s := range u.Path {
		if i > 0 {
			p += "/"
		}
		p += s
	}
	if u.IsStdlib {
		return fmt.Sprintf("use sdt/%s", p)
	}
	return fmt.Sprintf("use %s", p)
}
