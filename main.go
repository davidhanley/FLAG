package main

import (
    "bufio"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strconv"
    "strings"
)

// ====================== AST ======================
type Node interface{}

type Number int64
type Symbol string
type List []Node

// ====================== Parser ======================
func tokenize(input string) []string {
    input = strings.ReplaceAll(input, "(", " ( ")
    input = strings.ReplaceAll(input, ")", " ) ")
    return strings.Fields(input)
}

func parse(tokens []string, pos *int) (Node, error) {
    if *pos >= len(tokens) {
        return nil, fmt.Errorf("unexpected end of input")
    }

    token := tokens[*pos]
    *pos++

    if token == "(" {
        var list List
        for *pos < len(tokens) && tokens[*pos] != ")" {
            node, err := parse(tokens, pos)
            if err != nil {
                return nil, err
            }
            list = append(list, node)
        }
        if *pos < len(tokens) && tokens[*pos] == ")" {
            *pos++
        }
        return list, nil
    }

    if n, err := strconv.ParseInt(token, 10, 64); err == nil {
        return Number(n), nil
    }

    return Symbol(token), nil
}

// ====================== Compiler ======================
func compileToGo(expr Node) string {
    switch v := expr.(type) {
    case Number:
        return fmt.Sprintf("Int(%d)", v)

    case List:
        if len(v) == 0 {
            return "Int(0)"
        }
        op := v[0].(Symbol)
        args := make([]string, len(v)-1)
        for i, arg := range v[1:] {
            args[i] = compileToGo(arg)
        }

        switch op {
        case "+":
            return fmt.Sprintf("Add(%s)", strings.Join(args, ", "))
        case "*":
            return fmt.Sprintf("Mul(%s)", strings.Join(args, ", "))
        case "-":
            return fmt.Sprintf("Sub(%s)", strings.Join(args, ", "))
        case "/":
            return fmt.Sprintf("Div(%s)", strings.Join(args, ", "))
        default:
            return "Int(0)"
        }
    default:
        return "Int(0)"
    }
}

// ====================== Runtime (now inside the generated file) ======================
const runtimeCode = `package main

import "fmt"

type Int int64

func Add(args ...Int) Int {
    sum := Int(0)
    for _, v := range args {
        sum += v
    }
    return sum
}

func Mul(args ...Int) Int {
    prod := Int(1)
    for _, v := range args {
        prod *= v
    }
    return prod
}

func Sub(args ...Int) Int {
    if len(args) == 0 {
        return 0
    }
    res := args[0]
    for _, v := range args[1:] {
        res -= v
    }
    return res
}

func Div(args ...Int) Int {
    if len(args) == 0 {
        return 0
    }
    res := args[0]
    for _, v := range args[1:] {
        if v != 0 {
            res /= v
        }
    }
    return res
}

func main() {
    result := %s
    fmt.Println(result)
}
`

// ====================== REPL ======================
func main() {
    fmt.Println("FLAG MVP REPL - Functional Lisp Accelerated Go")
    fmt.Println("Type expressions like: (+ 1 (* 2 3))")
    fmt.Println("Ctrl+C to exit\n")

    scanner := bufio.NewScanner(os.Stdin)
    for {
        fmt.Print("> ")
        if !scanner.Scan() {
            break
        }

        input := strings.TrimSpace(scanner.Text())
        if input == "" || input == "quit" || input == "exit" {
            continue
        }

        tokens := tokenize(input)
        pos := 0
        expr, err := parse(tokens, &pos)
        if err != nil {
            fmt.Println("Parse error:", err)
            continue
        }

        body := compileToGo(expr)
        goCode := fmt.Sprintf(runtimeCode, body)

        tmpFile := filepath.Join(os.TempDir(), "flag_mvp.go")
        if err := os.WriteFile(tmpFile, []byte(goCode), 0644); err != nil {
            fmt.Println("File write error:", err)
            continue
        }

        cmd := exec.Command("go", "run", tmpFile)
        output, err := cmd.CombinedOutput()
        if err != nil {
            fmt.Println("Compile/Run error:")
            fmt.Println(string(output))
        } else {
            fmt.Print(string(output))
        }
    }
}
