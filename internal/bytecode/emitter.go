package bytecode

// Emit returns a minimal valid Kyra bytecode header.
// This is enough for the bootstrap phase.
func Emit(ast interface{}) []byte {
    // "KBC" + version byte
    return []byte{0x4B, 0x42, 0x43, 0x01}
}
