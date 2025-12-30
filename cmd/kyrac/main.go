package main

import (
    "flag"
    "fmt"
    "os"

    "kyra/internal/bytecode"
    "kyra/internal/kar"
    "kyra/internal/lexer"
    "kyra/internal/parser"
)

func main() {
    modeModule := flag.String("m", "", "Compile a single Kyra file to .kbc")
    modeKar := flag.String("kar", "", "Build a .kar executable archive")
    flag.Parse()

    if *modeModule != "" {
        compileModule(*modeModule)
        return
    }

    if *modeKar != "" {
        buildKar(*modeKar)
        return
    }

    fmt.Println("Usage:")
    fmt.Println("  kyrac -m <file.kyra>")
    fmt.Println("  kyrac -kar <project-folder>")
}

func compileModule(path string) {
    src, err := os.ReadFile(path)
    if err != nil {
        panic(err)
    }

    // Lexing
    lx := lexer.New(string(src))
    tokens := lx.Lex()

    // Parsing
    p := parser.New(tokens)
    ast := p.Parse()

    // Bytecode emission
    bc := bytecode.Emit(ast)

    out := path[:len(path)-5] + ".kbc"
    os.WriteFile(out, bc, 0644)

    fmt.Println("Compiled:", out)
}

func buildKar(project string) {
    err := kar.Build(project)
    if err != nil {
        panic(err)
    }
}
