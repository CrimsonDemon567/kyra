package kvm

import (
	"encoding/binary"
	"fmt"
	"math"
)

// ---------------------------
// VM structures
// ---------------------------

type VM struct {
	constants []interface{}
	code      []byte

	ip    int
	sp    int
	stack []interface{}

	callStack []Frame
}

type Frame struct {
	ipBackup int
	spBackup int
	code     []byte
	consts   []interface{}
}

// ---------------------------
// VM entry
// ---------------------------

func New(code []byte) *VM {
	vm := &VM{
		stack:     make([]interface{}, 0, 1024),
		callStack: []Frame{},
	}
	vm.loadModule(code)
	return vm
}

func (vm *VM) loadModule(code []byte) {
	// Header: KBC + version
	if string(code[:3]) != "KBC" {
		panic("Invalid bytecode header")
	}

	version := code[3]
	if version != 2 {
		panic(fmt.Sprintf("Unsupported KBC version: %d", version))
	}

	offset := 4

	// Function count
	fnCount := int(binary.LittleEndian.Uint32(code[offset:]))
	offset += 4

	// Skip function chunks (lazy load)
	for i := 0; i < fnCount; i++ {
		// constants
		cCount := int(binary.LittleEndian.Uint32(code[offset:]))
		offset += 4

		for j := 0; j < cCount; j++ {
			kind := code[offset]
			offset++

			switch kind {
			case 1: // string
				l := int(binary.LittleEndian.Uint32(code[offset:]))
				offset += 4 + l
			case 2: // float64
				offset += 8
			case 3: // int
				offset += 4
			default:
				panic("Unknown constant type in function chunk")
			}
		}

		// code
		l := int(binary.LittleEndian.Uint32(code[offset:]))
		offset += 4 + l
	}

	// Main chunk
	cCount := int(binary.LittleEndian.Uint32(code[offset:]))
	offset += 4

	vm.constants = make([]interface{}, cCount)

	for i := 0; i < cCount; i++ {
		kind := code[offset]
		offset++

		switch kind {
		case 1: // string
			l := int(binary.LittleEndian.Uint32(code[offset:]))
			offset += 4
			str := string(code[offset : offset+l])
			offset += l
			vm.constants[i] = str

		case 2: // float64
			bits := binary.LittleEndian.Uint64(code[offset:])
			offset += 8
			vm.constants[i] = math.Float64frombits(bits)

		case 3: // int
			v := int(binary.LittleEndian.Uint32(code[offset:]))
			offset += 4
			vm.constants[i] = v

		default:
			panic("Unknown constant type in main chunk")
		}
	}

	codeLen := int(binary.LittleEndian.Uint32(code[offset:]))
	offset += 4

	vm.code = code[offset : offset+codeLen]
	vm.ip = 0
	vm.sp = 0
}

// ---------------------------
// Stack helpers
// ---------------------------

func (vm *VM) push(v interface{}) {
	vm.stack = append(vm.stack, v)
	vm.sp++
}

func (vm *VM) pop() interface{} {
	if vm.sp == 0 {
		panic("Stack underflow")
	}
	vm.sp--
	v := vm.stack[vm.sp]
	vm.stack = vm.stack[:vm.sp]
	return v
}

func (vm *VM) peek() interface{} {
	if vm.sp == 0 {
		panic("Stack empty")
	}
	return vm.stack[vm.sp-1]
}

// ---------------------------
// Execution
// ---------------------------

func (vm *VM) Run() interface{} {
	for vm.ip < len(vm.code) {
		op := vm.code[vm.ip]
		vm.ip++

		switch op {

		case 0x01: // OP_CONST
			idx := vm.readInt()
			vm.push(vm.constants[idx])

		case 0x02: // OP_ADD
			b := vm.pop().(float64)
			a := vm.pop().(float64)
			vm.push(a + b)

		case 0x03: // OP_SUB
			b := vm.pop().(float64)
			a := vm.pop().(float64)
			vm.push(a - b)

		case 0x04: // OP_MUL
			b := vm.pop().(float64)
			a := vm.pop().(float64)
			vm.push(a * b)

		case 0x05: // OP_DIV
			b := vm.pop().(float64)
			a := vm.pop().(float64)
			vm.push(a / b)

		case 0x06: // OP_MOD
			b := vm.pop().(float64)
			a := vm.pop().(float64)
			vm.push(math.Mod(a, b))

		case 0x07: // OP_EQ
			b := vm.pop()
			a := vm.pop()
			vm.push(boolToFloat(a == b))

		case 0x08: // OP_NEQ
			b := vm.pop()
			a := vm.pop()
			vm.push(boolToFloat(a != b))

		case 0x09: // OP_LT
			b := vm.pop().(float64)
			a := vm.pop().(float64)
			vm.push(boolToFloat(a < b))

		case 0x0A: // OP_GT
			b := vm.pop().(float64)
			a := vm.pop().(float64)
			vm.push(boolToFloat(a > b))

		case 0x0B: // OP_LE
			b := vm.pop().(float64)
			a := vm.pop().(float64)
			vm.push(boolToFloat(a <= b))

		case 0x0C: // OP_GE
			b := vm.pop().(float64)
			a := vm.pop().(float64)
			vm.push(boolToFloat(a >= b))

		case 0x0D: // OP_AND
			b := vm.pop().(float64)
			a := vm.pop().(float64)
			vm.push(boolToFloat(a != 0 && b != 0))

		case 0x0E: // OP_OR
			b := vm.pop().(float64)
			a := vm.pop().(float64)
			vm.push(boolToFloat(a != 0 || b != 0))

		case 0x0F: // OP_NOT
			a := vm.pop().(float64)
			vm.push(boolToFloat(a == 0))

		case 0x10: // OP_LOAD
			idx := vm.readInt()
			vm.push(vm.constants[idx])

		case 0x11: // OP_STORE
			idx := vm.readInt()
			val := vm.pop()
			vm.constants[idx] = val

		case 0x12: // OP_CALL
			argCount := vm.readInt()
			fnID := int(vm.pop().(float64))
			vm.callFunction(fnID, argCount)

		case 0x13: // OP_RET
			if len(vm.callStack) == 0 {
				return vm.pop()
			}
			vm.returnFromFunction()

		case 0x14: // OP_JMP
			target := vm.readInt()
			vm.ip = target

		case 0x15: // OP_JMPF
			target := vm.readInt()
			cond := vm.pop().(float64)
			if cond == 0 {
				vm.ip = target
			}

		case 0x16: // OP_POP
			vm.pop()

		case 0x17: // OP_EXIT
			return nil

		default:
			panic(fmt.Sprintf("Unknown opcode: %02X", op))
		}
	}

	return nil
}

// ---------------------------
// Function calls
// ---------------------------

func (vm *VM) callFunction(fnID int, argCount int) {
	fnCode, fnConsts := loadFunction(fnID)

	frame := Frame{
		ipBackup: vm.ip,
		spBackup: vm.sp - argCount,
		code:     vm.code,
		consts:   vm.constants,
	}

	vm.callStack = append(vm.callStack, frame)

	vm.code = fnCode
	vm.constants = fnConsts
	vm.ip = 0
	vm.sp = 0

	for i := 0; i < argCount; i++ {
		vm.push(vm.stack[frame.spBackup+i])
	}
}

func (vm *VM) returnFromFunction() {
	ret := vm.pop()

	frame := vm.callStack[len(vm.callStack)-1]
	vm.callStack = vm.callStack[:len(vm.callStack)-1]

	vm.code = frame.code
	vm.constants = frame.consts
	vm.ip = frame.ipBackup
	vm.sp = frame.spBackup

	vm.push(ret)
}

// ---------------------------
// Helpers
// ---------------------------

func (vm *VM) readInt() int {
	v := int(binary.LittleEndian.Uint32(vm.code[vm.ip:]))
	vm.ip += 4
	return v
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
