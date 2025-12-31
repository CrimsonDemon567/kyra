package bytecode

import (
	"encoding/binary"
	"fmt"
	"kyra/internal/parser"
)

// ---------------------------
// Bytecode instruction set
// ---------------------------

const (
	OP_CONST = 0x01
	OP_ADD   = 0x02
	OP_SUB   = 0x03
	OP_MUL   = 0x04
	OP_DIV   = 0x05
	OP_MOD   = 0x06

	OP_EQ  = 0x07
	OP_NEQ = 0x08
	OP_LT  = 0x09
	OP_GT  = 0x0A
	OP_LE  = 0x0B
	OP_GE  = 0x0C

	OP_AND = 0x0D
	OP_OR  = 0x0E
	OP_NOT = 0x0F

	OP_LOAD  = 0x10
	OP_STORE = 0x11

	OP_CALL = 0x12
	OP_RET  = 0x13

	OP_JMP  = 0x14
	OP_JMPF = 0x15

	OP_POP  = 0x16
	OP_EXIT = 0x17
)

// ---------------------------
// Chunk structure
// ---------------------------

type Chunk struct {
	Code      []byte
	Constants []interface{}
	Names     map[string]int
}

func NewChunk() *Chunk {
	return &Chunk{
		Code:      []byte{},
		Constants: []interface{}{},
		Names:     map[string]int{},
	}
}

func (c *Chunk) emit(b byte) {
	c.Code = append(c.Code, b)
}

func (c *Chunk) emitInt(v int) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(v))
	c.Code = append(c.Code, buf...)
}

func (c *Chunk) addConst(v interface{}) int {
	c.Constants = append(c.Constants, v)
	return len(c.Constants) - 1
}

// ---------------------------
// Emitter entry
// ---------------------------

func Emit(ast *parser.AST) []byte {
	// Funktions-Tabelle für dieses Modul zurücksetzen
	resetFunctions()

	mainChunk := NewChunk()

	for _, stmt := range ast.TopLevel {
		emitStmt(mainChunk, stmt)
	}

	// Implizites return aus main
	mainChunk.emit(OP_RET)

	// Mit Funktionen + Main-Chunk zu KBC v2 encodieren
	return encodeModuleWithFunctions(mainChunk)
}

// ---------------------------
// Statement emission
// ---------------------------

func emitStmt(c *Chunk, stmt parser.Stmt) {
	switch s := stmt.(type) {

	case *parser.ExprStmt:
		emitExpr(c, s.Expr)
		c.emit(OP_POP)

	case *parser.LetStmt:
		emitExpr(c, s.Expr)
		slot := c.addConst(s.Name)
		c.emit(OP_STORE)
		c.emitInt(slot)

	case *parser.ReturnStmt:
		emitExpr(c, s.Value)
		c.emit(OP_RET)

	case *parser.ExitStmt:
		c.emit(OP_EXIT)

	case *parser.PassStmt:
		// no-op

	case *parser.IfStmt:
		emitExpr(c, s.Cond)
		c.emit(OP_JMPF)
		jumpPos := len(c.Code)
		c.emitInt(0)

		for _, st := range s.Then {
			emitStmt(c, st)
		}

		if len(s.Else) > 0 {
			c.emit(OP_JMP)
			elseJump := len(c.Code)
			c.emitInt(0)

			patchJump(c, jumpPos)

			for _, st := range s.Else {
				emitStmt(c, st)
			}

			patchJump(c, elseJump)
		} else {
			patchJump(c, jumpPos)
		}

	case *parser.WhileStmt:
		loopStart := len(c.Code)

		emitExpr(c, s.Cond)
		c.emit(OP_JMPF)
		exitPos := len(c.Code)
		c.emitInt(0)

		for _, st := range s.Body {
			emitStmt(c, st)
		}

		c.emit(OP_JMP)
		c.emitInt(loopStart)

		patchJump(c, exitPos)

	case *parser.ForStmt:
		// for i 10:
		emitExpr(c, s.Limit)
		limitSlot := c.addConst(s.VarName + "_limit")
		c.emit(OP_STORE)
		c.emitInt(limitSlot)

		iSlot := c.addConst(s.VarName)
		c.emit(OP_CONST)
		c.emitInt(c.addConst(float64(0)))
		c.emit(OP_STORE)
		c.emitInt(iSlot)

		loopStart := len(c.Code)

		// if i >= limit: break
		c.emit(OP_LOAD)
		c.emitInt(iSlot)
		c.emit(OP_LOAD)
		c.emitInt(limitSlot)
		c.emit(OP_GE)
		c.emit(OP_JMPF)
		exitPos := len(c.Code)
		c.emitInt(0)

		for _, st := range s.Body {
			emitStmt(c, st)
		}

		// i = i + 1
		c.emit(OP_LOAD)
		c.emitInt(iSlot)
		c.emit(OP_CONST)
		c.emitInt(c.addConst(float64(1)))
		c.emit(OP_ADD)
		c.emit(OP_STORE)
		c.emitInt(iSlot)

		c.emit(OP_JMP)
		c.emitInt(loopStart)

		patchJump(c, exitPos)

	case *parser.FuncDef:
		// Vollständige Funktionsunterstützung über functions.go
		emitFunctionDef(c, s)

	case *parser.FuncExprDef:
		emitFunctionExpr(c, s)

	case *parser.FuncOneLiner:
		emitFunctionOneLiner(c, s)

	default:
		panic(fmt.Sprintf("Unknown statement type: %T", stmt))
	}
}

// ---------------------------
// Expression emission
// ---------------------------

func emitExpr(c *Chunk, expr parser.Expr) {
	switch e := expr.(type) {

	case *parser.NumberExpr:
		val := parseNumber(e.Value)
		slot := c.addConst(val)
		c.emit(OP_CONST)
		c.emitInt(slot)

	case *parser.StringExpr:
		slot := c.addConst(e.Value)
		c.emit(OP_CONST)
		c.emitInt(slot)

	case *parser.BoolExpr:
		val := 0.0
		if e.Value {
			val = 1.0
		}
		slot := c.addConst(val)
		c.emit(OP_CONST)
		c.emitInt(slot)

	case *parser.IdentExpr:
		slot := c.addConst(e.Name)
		c.emit(OP_LOAD)
		c.emitInt(slot)

	case *parser.AssignExpr:
		emitExpr(c, e.Expr)
		slot := c.addConst(e.Name)
		c.emit(OP_STORE)
		c.emitInt(slot)

	case *parser.UnaryExpr:
		emitExpr(c, e.Expr)
		switch e.Op {
		case "-":
			c.emit(OP_CONST)
			c.emitInt(c.addConst(float64(-1)))
			c.emit(OP_MUL)
		case "!":
			c.emit(OP_NOT)
		default:
			panic("Unknown unary operator: " + e.Op)
		}

	case *parser.BinaryExpr:
		emitExpr(c, e.Left)
		emitExpr(c, e.Right)
		emitBinaryOp(c, e.Op)

	case *parser.CallExpr:
		for _, arg := range e.Args {
			emitExpr(c, arg)
		}
		c.emit(OP_CALL)
		c.emitInt(len(e.Args))

	case *parser.MemberExpr:
		panic("Member access not implemented yet")

	case *parser.ParenExpr:
		emitExpr(c, e.Expr)

	default:
		panic(fmt.Sprintf("Unknown expression type: %T", expr))
	}
}

func emitBinaryOp(c *Chunk, op string) {
	switch op {
	case "+":
		c.emit(OP_ADD)
	case "-":
		c.emit(OP_SUB)
	case "*":
		c.emit(OP_MUL)
	case "/":
		c.emit(OP_DIV)
	case "%":
		c.emit(OP_MOD)
	case "==":
		c.emit(OP_EQ)
	case "!=":
		c.emit(OP_NEQ)
	case "<":
		c.emit(OP_LT)
	case ">":
		c.emit(OP_GT)
	case "<=":
		c.emit(OP_LE)
	case ">=":
		c.emit(OP_GE)
	case "&&":
		c.emit(OP_AND)
	case "||":
		c.emit(OP_OR)
	default:
		panic("Unknown binary operator: " + op)
	}
}

// ---------------------------
// Helpers
// ---------------------------

func patchJump(c *Chunk, pos int) {
	target := len(c.Code)
	binary.LittleEndian.PutUint32(c.Code[pos:pos+4], uint32(target))
}

func parseNumber(s string) float64 {
	var v float64
	fmt.Sscanf(s, "%f", &v)
	return v
}
