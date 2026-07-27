package compiler

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"strconv"
	"strings"
	"unicode"
)

const runtimeAlias = "flagrt"

type exprKind uint8

const (
	exprKindValue exprKind = iota + 1
	exprKindString
	exprKindBool
)

type goExpr struct {
	code string
	kind exprKind
}

func exprPos(expr Expr) (int, int, bool) {
	switch value := expr.(type) {
	case ListExpr:
		return value.Line, value.Col, value.Line > 0 && value.Col > 0
	case VectorExpr:
		return value.Line, value.Col, value.Line > 0 && value.Col > 0
	case MapExpr:
		return value.Line, value.Col, value.Line > 0 && value.Col > 0
	case SetExpr:
		return value.Line, value.Col, value.Line > 0 && value.Col > 0
	case HashFnExpr:
		return value.Line, value.Col, value.Line > 0 && value.Col > 0
	case SymbolExpr:
		return value.Line, value.Col, value.Line > 0 && value.Col > 0
	case KeywordExpr:
		return value.Line, value.Col, value.Line > 0 && value.Col > 0
	case QuotedSymbolExpr:
		return value.Line, value.Col, value.Line > 0 && value.Col > 0
	case QuotedListExpr:
		return value.Line, value.Col, value.Line > 0 && value.Col > 0
	case StringExpr:
		return value.Line, value.Col, value.Line > 0 && value.Col > 0
	case IntExpr:
		return value.Line, value.Col, value.Line > 0 && value.Col > 0
	case FloatExpr:
		return value.Line, value.Col, value.Line > 0 && value.Col > 0
	case RatioExpr:
		return value.Line, value.Col, value.Line > 0 && value.Col > 0
	default:
		return 0, 0, false
	}
}

func exprError(expr Expr, msg string) error {
	if line, col, ok := exprPos(expr); ok {
		return fmt.Errorf("%s at %d:%d", msg, line, col)
	}
	return fmt.Errorf("%s", msg)
}

func exprToSourceString(expr Expr) string {
	switch value := expr.(type) {
	case StringExpr:
		return strconv.Quote(value.Value)
	case IntExpr:
		return fmt.Sprintf("%d", value.Value)
	case FloatExpr:
		if value.Raw != "" {
			return value.Raw
		}
		return strconv.FormatFloat(value.Value, 'g', -1, 64)
	case RatioExpr:
		return fmt.Sprintf("%d/%d", value.Numerator, value.Denominator)
	case KeywordExpr:
		return ":" + value.Name
	case SymbolExpr:
		return value.Name
	case QuotedSymbolExpr:
		return "'" + value.Name
	case ListExpr:
		parts := make([]string, 0, len(value.Elements))
		for _, item := range value.Elements {
			parts = append(parts, exprToSourceString(item))
		}
		return "(" + strings.Join(parts, " ") + ")"
	case VectorExpr:
		parts := make([]string, 0, len(value.Elements))
		for _, item := range value.Elements {
			parts = append(parts, exprToSourceString(item))
		}
		return "[" + strings.Join(parts, " ") + "]"
	case MapExpr:
		parts := make([]string, 0, len(value.Entries))
		for _, item := range value.Entries {
			parts = append(parts, exprToSourceString(item))
		}
		return "{" + strings.Join(parts, " ") + "}"
	case SetExpr:
		parts := make([]string, 0, len(value.Elements))
		for _, item := range value.Elements {
			parts = append(parts, exprToSourceString(item))
		}
		return "#{" + strings.Join(parts, " ") + "}"
	case HashFnExpr:
		return "#(" + exprToSourceString(value.Body) + ")"
	default:
		return fmt.Sprintf("%v", expr)
	}
}

type mainStmt struct {
	code string
}

type testCase struct {
	goName     string
	label      string
	line       int
	bodySource string
}

type functionDef struct {
	goName       string
	variadicName string
	arityName    string
	hasRest      bool
	doc          string
	params       []string
	localInits   []string
	body         string
}

type varDef struct {
	goName string
	doc    string
	expr   string
}

type compileContext struct {
	functions         map[string]functionDef
	globals           map[string]exprKind
	macros            map[string]macroDef
	selfFunctionName  string
	selfFunctionArity int
	selfArityName     string
}

type macroDef struct {
	params    []string
	restParam string
	doc       string
	body      Expr
}

//go:embed macros.flag
var standardMacrosSource string

func newCompileContext() (compileContext, error) {
	ctx := compileContext{
		functions: make(map[string]functionDef),
		globals:   make(map[string]exprKind),
		macros:    make(map[string]macroDef),
	}
	if err := loadStandardMacros(&ctx); err != nil {
		return compileContext{}, err
	}
	return ctx, nil
}

func loadStandardMacros(ctx *compileContext) error {
	macroAST, err := ParseFile(standardMacrosSource)
	if err != nil {
		return fmt.Errorf("parse standard macros: %w", err)
	}
	for _, form := range macroAST.Forms {
		list, ok := form.(ListExpr)
		if !ok {
			return fmt.Errorf("invalid standard macro form")
		}
		head, ok := list.Elements[0].(SymbolExpr)
		if !ok || head.Name != "defmacro" {
			return fmt.Errorf("invalid standard macro declaration")
		}
		name, def, err := compileDefmacro(list)
		if err != nil {
			return err
		}
		ctx.macros[name] = def
	}
	return nil
}

// Compile translates a small FLAG source file into a Go program.
func Compile(source string) ([]byte, error) {
	namespace, functions, vars, stmts, tests, needsFmt, err := compileForms(source)
	if err != nil {
		return nil, err
	}
	if len(tests) > 0 {
		needsFmt = true
	}

	var out bytes.Buffer
	out.WriteString("package main\n\n")
	out.WriteString("import (\n")
	if needsFmt {
		out.WriteString("\t\"fmt\"\n")
	}
	if len(tests) > 0 {
		out.WriteString("\t\"os\"\n")
	}
	out.WriteString("\tflagrt \"flag-lang/runtime\"\n")
	out.WriteString(")\n\n")

	if namespace != "" {
		fmt.Fprintf(&out, "// Source namespace: %s\n", namespace)
	}

	for _, fn := range functions {
		if fn.doc != "" {
			fmt.Fprintf(&out, "// %s\n", fn.doc)
		}
		out.WriteString(renderFunctionDef(fn))
		out.WriteString("\n")
	}

	for _, v := range vars {
		if v.doc != "" {
			fmt.Fprintf(&out, "// %s\n", v.doc)
		}
		fmt.Fprintf(&out, "var %s = %s\n", v.goName, v.expr)
	}
	if len(vars) > 0 {
		out.WriteString("\n")
	}

	if len(tests) > 0 {
		out.WriteString("type flagTestCase struct {\n")
		out.WriteString("\tname string\n")
		out.WriteString("\tline int\n")
		out.WriteString("\tbody string\n")
		out.WriteString("\tfn func() flagrt.Value\n")
		out.WriteString("}\n\n")
		out.WriteString("func runFlagTestCase(tc flagTestCase) (passed bool) {\n")
		out.WriteString("\tdefer func() {\n")
		out.WriteString("\t\tif r := recover(); r != nil {\n")
		out.WriteString("\t\t\tfmt.Printf(\"FAIL %s\\n%s\\n%s\\n\", tc.name, tc.body, r)\n")
		out.WriteString("\t\t\treturn\n")
		out.WriteString("\t\t}\n")
		out.WriteString("\t\tpassed = true\n")
		out.WriteString("\t\tfmt.Printf(\"PASS %s\\n\", tc.name)\n")
		out.WriteString("\t}()\n")
		out.WriteString("\ttc.fn()\n")
		out.WriteString("\treturn passed\n")
		out.WriteString("}\n\n")
	}

	out.WriteString("func main() {\n")
	if len(tests) > 0 {
		out.WriteString("\ttests := []flagTestCase{\n")
		for _, tc := range tests {
			fmt.Fprintf(&out, "\t\t{name: %q, line: %d, body: %q, fn: %s},\n", tc.label, tc.line, tc.bodySource, tc.goName)
		}
		out.WriteString("\t}\n")
		out.WriteString("\tpassed := 0\n")
		out.WriteString("\tfailed := 0\n")
		out.WriteString("\tfor _, tc := range tests {\n")
		out.WriteString("\t\tif runFlagTestCase(tc) {\n")
		out.WriteString("\t\t\tpassed++\n")
		out.WriteString("\t\t} else {\n")
		out.WriteString("\t\t\tfailed++\n")
		out.WriteString("\t\t}\n")
		out.WriteString("\t}\n")
		out.WriteString("\tfmt.Printf(\"%d passed, %d failed\\n\", passed, failed)\n")
		out.WriteString("\tif failed > 0 {\n")
		out.WriteString("\t\tos.Exit(1)\n")
		out.WriteString("\t}\n")
	} else {
		for _, stmt := range stmts {
			fmt.Fprintf(&out, "\t%s\n", stmt.code)
		}
	}
	out.WriteString("}\n")

	formatted, err := format.Source(out.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated Go: %w", err)
	}

	return formatted, nil
}

// CompileExpression translates a single FLAG expression into a Go expression.
// Numeric expressions are wrapped in runtime conversion so the result is printable as any.
func CompileExpression(source string) (string, error) {
	ast, err := ParseFile(source)
	if err != nil {
		return "", err
	}
	if len(ast.Forms) != 1 {
		return "", fmt.Errorf("expected exactly one expression")
	}

	ctx, err := newCompileContext()
	if err != nil {
		return "", err
	}
	expr, err := exprToGo(ast.Forms[0], ctx, nil)
	if err != nil {
		return "", err
	}
	if expr.kind == exprKindValue {
		return fmt.Sprintf("%s.ValueToAny(%s)", runtimeAlias, expr.code), nil
	}
	return expr.code, nil
}

type ReplCompiler struct {
	ctx compileContext
}

type ReplCompiled struct {
	Setup      string
	ResultExpr string
}

func NewReplCompiler() *ReplCompiler {
	ctx, err := newCompileContext()
	if err != nil {
		panic(err)
	}
	return &ReplCompiler{
		ctx: ctx,
	}
}

func (r *ReplCompiler) CompileLine(source string) (ReplCompiled, error) {
	ast, err := ParseFile(source)
	if err != nil {
		return ReplCompiled{}, err
	}
	if len(ast.Forms) != 1 {
		return ReplCompiled{}, fmt.Errorf("expected exactly one expression")
	}

	if list, ok := ast.Forms[0].(ListExpr); ok && len(list.Elements) > 0 {
		if head, ok := list.Elements[0].(SymbolExpr); ok {
			switch head.Name {
			case "def":
				binding, exprKind, isNew, err := compileDefForRepl(list, r.ctx)
				if err != nil {
					return ReplCompiled{}, err
				}
				r.ctx.globals[binding.goName] = exprKind
				setup := fmt.Sprintf("%s = %s", binding.goName, binding.expr)
				if isNew {
					setup = fmt.Sprintf("var %s flagrt.Value;;%s = %s", binding.goName, binding.goName, binding.expr)
				}
				return ReplCompiled{
					Setup:      setup,
					ResultExpr: fmt.Sprintf("%s.ValueToAny(%s)", runtimeAlias, binding.goName),
				}, nil
			case "deftest":
				def, err := compileDeftest(list, r.ctx)
				if err != nil {
					return ReplCompiled{}, err
				}
				_, exists := r.ctx.functions[def.goName]
				r.ctx.functions[def.goName] = def
				setupParts := make([]string, 0, 2)
				if !exists {
					setupParts = append(setupParts, fmt.Sprintf("func %s() flagrt.Value { return %s }", def.arityName, def.body))
				} else {
					setupParts = append(setupParts, fmt.Sprintf("%s = func() flagrt.Value { return %s }", def.arityName, def.body))
				}
				setupParts = append(setupParts, fmt.Sprintf("%s()", def.arityName))
				return ReplCompiled{Setup: strings.Join(setupParts, ";;")}, nil
			case "defn", "defn-":
				def, err := compileDefn(list, r.ctx)
				if err != nil {
					return ReplCompiled{}, err
				}
				_, exists := r.ctx.functions[def.goName]
				r.ctx.functions[def.goName] = def
				r.ctx.globals[def.goName] = exprKindValue

				setupParts := make([]string, 0, 6)
				if !exists {
					setupParts = append(setupParts,
						fmt.Sprintf("var %s %s", def.arityName, renderDirectFunctionType(def)),
						fmt.Sprintf("var %s func(args ...flagrt.Value) flagrt.Value", def.variadicName),
						fmt.Sprintf("var %s flagrt.Value", def.goName),
					)
				}
				setupParts = append(setupParts,
					fmt.Sprintf("%s = %s", def.arityName, renderDirectFunctionLiteral(def)),
					fmt.Sprintf("%s = %s", def.variadicName, renderVariadicFunctionLiteral(def)),
					fmt.Sprintf("%s = %s.NewFunction(%s)", def.goName, runtimeAlias, def.variadicName),
				)

				return ReplCompiled{
					Setup:      strings.Join(setupParts, ";;"),
					ResultExpr: fmt.Sprintf("%q", def.goName),
				}, nil
			case "defmacro":
				name, def, err := compileDefmacro(list)
				if err != nil {
					return ReplCompiled{}, err
				}
				r.ctx.macros[name] = def
				return ReplCompiled{ResultExpr: fmt.Sprintf("%q", name)}, nil
			}
		}
	}

	expanded, err := macroExpand(ast.Forms[0], r.ctx, 0)
	if err != nil {
		return ReplCompiled{}, err
	}

	expr, err := exprToGo(expanded, r.ctx, nil)
	if err != nil {
		return ReplCompiled{}, err
	}
	if expr.kind == exprKindValue {
		return ReplCompiled{ResultExpr: fmt.Sprintf("%s.ValueToAny(%s)", runtimeAlias, expr.code)}, nil
	}
	return ReplCompiled{ResultExpr: expr.code}, nil
}

func compileForms(source string) (string, []functionDef, []varDef, []mainStmt, []testCase, bool, error) {
	ast, err := ParseFile(source)
	if err != nil {
		return "", nil, nil, nil, nil, false, err
	}

	ctx, err := newCompileContext()
	if err != nil {
		return "", nil, nil, nil, nil, false, err
	}
	namespace := ""
	functions := make([]functionDef, 0, len(ast.Forms))
	vars := make([]varDef, 0, len(ast.Forms))
	stmts := make([]mainStmt, 0, len(ast.Forms))
	tests := make([]testCase, 0, len(ast.Forms))
	needsFmt := false

	for _, form := range ast.Forms {
		if list, ok := form.(ListExpr); ok && len(list.Elements) > 0 {
			if head, ok := list.Elements[0].(SymbolExpr); ok && head.Name == "defmacro" {
				name, def, err := compileDefmacro(list)
				if err != nil {
					return "", nil, nil, nil, nil, false, err
				}
				ctx.macros[name] = def
				continue
			}
		}

		expanded, err := macroExpand(form, ctx, 0)
		if err != nil {
			return "", nil, nil, nil, nil, false, err
		}

		form = expanded
		list, ok := form.(ListExpr)
		if !ok || len(list.Elements) == 0 {
			return "", nil, nil, nil, nil, false, fmt.Errorf("unsupported form")
		}

		head, ok := list.Elements[0].(SymbolExpr)
		if !ok {
			return "", nil, nil, nil, nil, false, fmt.Errorf("unsupported form")
		}

		switch head.Name {
		case "ns":
			if namespace != "" {
				return "", nil, nil, nil, nil, false, fmt.Errorf("namespace already declared")
			}
			if len(list.Elements) != 2 {
				return "", nil, nil, nil, nil, false, fmt.Errorf("ns expects one namespace symbol")
			}
			name, ok := list.Elements[1].(SymbolExpr)
			if !ok || name.Name == "" {
				return "", nil, nil, nil, nil, false, fmt.Errorf("namespace cannot be empty")
			}
			namespace = name.Name
		case "defn", "defn-":
			def, err := compileDefn(list, ctx)
			if err != nil {
				return "", nil, nil, nil, nil, false, err
			}
			if _, exists := ctx.functions[def.goName]; exists {
				return "", nil, nil, nil, nil, false, fmt.Errorf("function %q already defined", def.goName)
			}
			ctx.functions[def.goName] = def
			ctx.globals[def.goName] = exprKindValue
			functions = append(functions, def)
			vars = append(vars, varDef{
				goName: def.goName,
				expr:   fmt.Sprintf("%s.NewFunction(%s)", runtimeAlias, def.variadicName),
			})
		case "deftest":
			def, err := compileDeftest(list, ctx)
			if err != nil {
				return "", nil, nil, nil, nil, false, err
			}
			if _, exists := ctx.functions[def.goName]; exists {
				return "", nil, nil, nil, nil, false, fmt.Errorf("test %q already defined", def.goName)
			}
			ctx.functions[def.goName] = def
			functions = append(functions, def)
			testName := def.goName
			if len(list.Elements) > 1 {
				if sym, ok := list.Elements[1].(SymbolExpr); ok && sym.Name != "" {
					testName = sym.Name
				}
			}
			tests = append(tests, testCase{
				goName:     def.arityName,
				label:      testName,
				line:       list.Line,
				bodySource: exprToSourceString(ListExpr{Elements: list.Elements[2:], Line: list.Line, Col: list.Col}),
			})
		case "def":
			binding, kind, err := compileDef(list, ctx)
			if err != nil {
				return "", nil, nil, nil, nil, false, err
			}
			ctx.globals[binding.goName] = kind
			vars = append(vars, binding)
		case "defmacro":
			name, def, err := compileDefmacro(list)
			if err != nil {
				return "", nil, nil, nil, nil, false, err
			}
			ctx.macros[name] = def
		case "println":
			needsFmt = true
			arg, err := strArgExprForGoCall(list.Elements[1:], ctx, nil)
			if err != nil {
				return "", nil, nil, nil, nil, false, err
			}
			stmts = append(stmts, mainStmt{code: fmt.Sprintf("fmt.Println(%s)", arg)})
		case "print":
			needsFmt = true
			arg, err := argumentExprForGoCall(list.Elements[1:], ctx, nil)
			if err != nil {
				return "", nil, nil, nil, nil, false, err
			}
			stmts = append(stmts, mainStmt{code: fmt.Sprintf("fmt.Print(%s)", arg)})
		default:
			expr, err := exprToGo(form, ctx, nil)
			if err != nil {
				return "", nil, nil, nil, nil, false, err
			}
			stmts = append(stmts, mainStmt{code: fmt.Sprintf("_ = %s", expr.code)})
		}
	}

	return namespace, functions, vars, stmts, tests, needsFmt, nil
}

func compileDefn(form ListExpr, ctx compileContext) (functionDef, error) {
	if len(form.Elements) != 4 && len(form.Elements) != 5 {
		return functionDef{}, fmt.Errorf("defn expects name, optional docstring, vector params, and body")
	}

	nameExpr, ok := form.Elements[1].(SymbolExpr)
	if !ok || nameExpr.Name == "" {
		return functionDef{}, fmt.Errorf("defn expects a function name")
	}

	doc := ""
	paramsIndex := 2
	bodyIndex := 3
	if len(form.Elements) == 5 {
		docExpr, ok := form.Elements[2].(StringExpr)
		if !ok {
			return functionDef{}, fmt.Errorf("defn docstring must be a string")
		}
		doc = docExpr.Value
		paramsIndex = 3
		bodyIndex = 4
	}

	goName, err := toGoIdentifier(nameExpr.Name)
	if err != nil {
		return functionDef{}, err
	}

	paramsExpr, ok := form.Elements[paramsIndex].(VectorExpr)
	if !ok {
		return functionDef{}, fmt.Errorf("defn expects a parameter vector")
	}
	params, localSymbols, localInits, hasRest, err := bindLambdaParams(paramsExpr, ctx, nil, "defn")
	if err != nil {
		return functionDef{}, err
	}

	fnCtx := compileContext{
		functions:         make(map[string]functionDef, len(ctx.functions)+1),
		globals:           make(map[string]exprKind, len(ctx.globals)+1),
		macros:            ctx.macros,
		selfFunctionName:  goName,
		selfFunctionArity: len(params),
		selfArityName:     fmt.Sprintf("%s_arity_%d", goName, len(params)),
	}
	if hasRest {
		fnCtx.selfArityName = goName + "_variadic"
	}
	for name, kind := range ctx.globals {
		fnCtx.globals[name] = kind
	}
	fnCtx.globals[goName] = exprKindValue
	for name, def := range ctx.functions {
		fnCtx.functions[name] = def
	}
	fnCtx.functions[goName] = functionDef{
		goName:       goName,
		variadicName: goName + "_variadic",
		arityName:    fnCtx.selfArityName,
		hasRest:      hasRest,
		doc:          doc,
		params:       params,
	}

	body, err := exprToGo(form.Elements[bodyIndex], fnCtx, localSymbols)
	if err != nil {
		return functionDef{}, err
	}
	if body.kind != exprKindValue {
		return functionDef{}, fmt.Errorf("defn body must evaluate to Value")
	}

	return functionDef{
		goName:       goName,
		variadicName: goName + "_variadic",
		arityName:    fnCtx.selfArityName,
		hasRest:      hasRest,
		doc:          doc,
		params:       params,
		localInits:   localInits,
		body:         body.code,
	}, nil
}

func compileDef(form ListExpr, ctx compileContext) (varDef, exprKind, error) {
	binding, kind, _, err := compileDefForRepl(form, ctx)
	if err != nil {
		return varDef{}, 0, err
	}
	if kind != exprKindValue {
		return varDef{}, 0, fmt.Errorf("def value must evaluate to Value")
	}
	return binding, kind, nil
}

func compileDefForRepl(form ListExpr, ctx compileContext) (varDef, exprKind, bool, error) {
	if len(form.Elements) != 3 && len(form.Elements) != 4 {
		return varDef{}, 0, false, fmt.Errorf("def expects name, optional docstring, and value")
	}

	nameExpr, ok := form.Elements[1].(SymbolExpr)
	if !ok || nameExpr.Name == "" {
		return varDef{}, 0, false, fmt.Errorf("def expects a symbol name")
	}
	doc := ""
	valueIndex := 2
	if len(form.Elements) == 4 {
		docExpr, ok := form.Elements[2].(StringExpr)
		if !ok {
			return varDef{}, 0, false, fmt.Errorf("def docstring must be a string")
		}
		doc = docExpr.Value
		valueIndex = 3
	}
	goName, err := toGoIdentifier(nameExpr.Name)
	if err != nil {
		return varDef{}, 0, false, err
	}

	valueExpr, err := exprToGo(form.Elements[valueIndex], ctx, nil)
	if err != nil {
		return varDef{}, 0, false, err
	}
	if valueExpr.kind != exprKindValue {
		return varDef{}, 0, false, fmt.Errorf("def value must evaluate to Value")
	}

	_, exists := ctx.globals[goName]
	return varDef{goName: goName, doc: doc, expr: valueExpr.code}, valueExpr.kind, !exists, nil
}

func compileDefmacro(form ListExpr) (string, macroDef, error) {
	if len(form.Elements) != 4 && len(form.Elements) != 5 {
		return "", macroDef{}, fmt.Errorf("defmacro expects name, optional docstring, vector params, and body")
	}
	nameExpr, ok := form.Elements[1].(SymbolExpr)
	if !ok || nameExpr.Name == "" {
		return "", macroDef{}, fmt.Errorf("defmacro expects a macro name")
	}
	doc := ""
	paramsIndex := 2
	bodyIndex := 3
	if len(form.Elements) == 5 {
		docExpr, ok := form.Elements[2].(StringExpr)
		if !ok {
			return "", macroDef{}, fmt.Errorf("defmacro docstring must be a string")
		}
		doc = docExpr.Value
		paramsIndex = 3
		bodyIndex = 4
	}
	paramsExpr, ok := form.Elements[paramsIndex].(VectorExpr)
	if !ok {
		return "", macroDef{}, fmt.Errorf("defmacro expects a parameter vector")
	}

	params := make([]string, 0, len(paramsExpr.Elements))
	restParam := ""
	for i := 0; i < len(paramsExpr.Elements); i++ {
		sym, ok := paramsExpr.Elements[i].(SymbolExpr)
		if !ok || sym.Name == "" {
			return "", macroDef{}, fmt.Errorf("defmacro parameters must be symbols")
		}
		if sym.Name == "&" {
			if restParam != "" || i != len(paramsExpr.Elements)-2 {
				return "", macroDef{}, fmt.Errorf("defmacro varargs must use [& name] at end")
			}
			next, ok := paramsExpr.Elements[i+1].(SymbolExpr)
			if !ok || next.Name == "" || next.Name == "&" {
				return "", macroDef{}, fmt.Errorf("defmacro varargs expects symbol after &")
			}
			restParam = next.Name
			break
		}
		params = append(params, sym.Name)
	}

	return nameExpr.Name, macroDef{
		params:    params,
		restParam: restParam,
		doc:       doc,
		body:      form.Elements[bodyIndex],
	}, nil
}

func compileDeftest(form ListExpr, ctx compileContext) (functionDef, error) {
	if len(form.Elements) < 3 {
		return functionDef{}, fmt.Errorf("deftest expects a name and body")
	}

	nameExpr, ok := form.Elements[1].(SymbolExpr)
	if !ok || nameExpr.Name == "" {
		return functionDef{}, fmt.Errorf("deftest expects a test name")
	}

	goName, err := toGoIdentifier(nameExpr.Name)
	if err != nil {
		return functionDef{}, err
	}

	body, err := testingBodyExprToGo(form.Elements[2:], ctx, nil)
	if err != nil {
		return functionDef{}, err
	}

	return functionDef{
		goName:       goName,
		variadicName: goName + "_variadic",
		arityName:    goName,
		body:         body.code,
	}, nil
}

func macroExpand(expr Expr, ctx compileContext, depth int) (Expr, error) {
	if depth > 100 {
		return nil, fmt.Errorf("macro expansion depth exceeded")
	}

	list, ok := expr.(ListExpr)
	if ok && len(list.Elements) > 0 {
		if head, ok := list.Elements[0].(SymbolExpr); ok {
			if macro, ok := ctx.macros[head.Name]; ok {
				expanded, err := applyMacro(macro, list.Elements[1:])
				if err != nil {
					return nil, err
				}
				return macroExpand(expanded, ctx, depth+1)
			}
		}
	}

	switch value := expr.(type) {
	case ListExpr:
		out := make([]Expr, 0, len(value.Elements))
		for _, item := range value.Elements {
			expanded, err := macroExpand(item, ctx, depth)
			if err != nil {
				return nil, err
			}
			out = append(out, expanded)
		}
		return ListExpr{Elements: out, Line: value.Line, Col: value.Col}, nil
	case VectorExpr:
		out := make([]Expr, 0, len(value.Elements))
		for _, item := range value.Elements {
			expanded, err := macroExpand(item, ctx, depth)
			if err != nil {
				return nil, err
			}
			out = append(out, expanded)
		}
		return VectorExpr{Elements: out, Line: value.Line, Col: value.Col}, nil
	case MapExpr:
		out := make([]Expr, 0, len(value.Entries))
		for _, item := range value.Entries {
			expanded, err := macroExpand(item, ctx, depth)
			if err != nil {
				return nil, err
			}
			out = append(out, expanded)
		}
		return MapExpr{Entries: out, Line: value.Line, Col: value.Col}, nil
	case SetExpr:
		out := make([]Expr, 0, len(value.Elements))
		for _, item := range value.Elements {
			expanded, err := macroExpand(item, ctx, depth)
			if err != nil {
				return nil, err
			}
			out = append(out, expanded)
		}
		return SetExpr{Elements: out, Line: value.Line, Col: value.Col}, nil
	case HashFnExpr:
		expanded, err := macroExpand(value.Body, ctx, depth)
		if err != nil {
			return nil, err
		}
		return HashFnExpr{Body: expanded, Line: value.Line, Col: value.Col}, nil
	default:
		return expr, nil
	}
}

func applyMacro(m macroDef, args []Expr) (Expr, error) {
	if m.restParam == "" && len(args) != len(m.params) {
		return nil, fmt.Errorf("macro expects exactly %d arguments", len(m.params))
	}
	if m.restParam != "" && len(args) < len(m.params) {
		return nil, fmt.Errorf("macro expects at least %d arguments", len(m.params))
	}

	values := make(map[string]Expr, len(m.params))
	for i, name := range m.params {
		values[name] = copyExpr(args[i])
	}
	restArgs := make([]Expr, 0)
	if m.restParam != "" {
		for _, arg := range args[len(m.params):] {
			restArgs = append(restArgs, copyExpr(arg))
		}
	}
	return substituteMacroExpr(m.body, values, m.restParam, restArgs)
}

func substituteMacroExpr(expr Expr, values map[string]Expr, restParam string, restArgs []Expr) (Expr, error) {
	switch value := expr.(type) {
	case SymbolExpr:
		if replacement, ok := values[value.Name]; ok {
			return copyExpr(replacement), nil
		}
		return value, nil
	case ListExpr:
		out := make([]Expr, 0, len(value.Elements))
		for _, item := range value.Elements {
			sym, isSym := item.(SymbolExpr)
			if isSym && restParam != "" && sym.Name == restParam {
				for _, restArg := range restArgs {
					out = append(out, copyExpr(restArg))
				}
				continue
			}
			sub, err := substituteMacroExpr(item, values, restParam, restArgs)
			if err != nil {
				return nil, err
			}
			out = append(out, sub)
		}
		if len(out) > 0 {
			if head, ok := out[0].(SymbolExpr); ok {
				if expanded, ok, err := applyMacroBuiltin(head.Name, out[1:]); ok || err != nil {
					return expanded, err
				}
			}
		}
		return ListExpr{Elements: out}, nil
	case VectorExpr:
		out := make([]Expr, 0, len(value.Elements))
		for _, item := range value.Elements {
			sub, err := substituteMacroExpr(item, values, restParam, restArgs)
			if err != nil {
				return nil, err
			}
			out = append(out, sub)
		}
		return VectorExpr{Elements: out}, nil
	case MapExpr:
		out := make([]Expr, 0, len(value.Entries))
		for _, item := range value.Entries {
			sub, err := substituteMacroExpr(item, values, restParam, restArgs)
			if err != nil {
				return nil, err
			}
			out = append(out, sub)
		}
		return MapExpr{Entries: out}, nil
	case SetExpr:
		out := make([]Expr, 0, len(value.Elements))
		for _, item := range value.Elements {
			sub, err := substituteMacroExpr(item, values, restParam, restArgs)
			if err != nil {
				return nil, err
			}
			out = append(out, sub)
		}
		return SetExpr{Elements: out}, nil
	case HashFnExpr:
		body, err := substituteMacroExpr(value.Body, values, restParam, restArgs)
		if err != nil {
			return nil, err
		}
		return HashFnExpr{Body: body}, nil
	default:
		return expr, nil
	}
}

func applyMacroBuiltin(name string, args []Expr) (Expr, bool, error) {
	switch name {
	case "macro-cond":
		expanded, err := expandCondMacro(args)
		return expanded, true, err
	case "macro-thread-first":
		expanded, err := expandThreadFirstMacro(args)
		return expanded, true, err
	case "macro-thread-last":
		expanded, err := expandThreadLastMacro(args)
		return expanded, true, err
	case "macro-some-thread-first":
		expanded, err := expandSomeThreadFirstMacro(args)
		return expanded, true, err
	default:
		return nil, false, nil
	}
}

func copyExpr(expr Expr) Expr {
	switch value := expr.(type) {
	case ListExpr:
		out := make([]Expr, 0, len(value.Elements))
		for _, item := range value.Elements {
			out = append(out, copyExpr(item))
		}
		return ListExpr{Elements: out, Line: value.Line, Col: value.Col}
	case VectorExpr:
		out := make([]Expr, 0, len(value.Elements))
		for _, item := range value.Elements {
			out = append(out, copyExpr(item))
		}
		return VectorExpr{Elements: out, Line: value.Line, Col: value.Col}
	case MapExpr:
		out := make([]Expr, 0, len(value.Entries))
		for _, item := range value.Entries {
			out = append(out, copyExpr(item))
		}
		return MapExpr{Entries: out, Line: value.Line, Col: value.Col}
	case SetExpr:
		out := make([]Expr, 0, len(value.Elements))
		for _, item := range value.Elements {
			out = append(out, copyExpr(item))
		}
		return SetExpr{Elements: out, Line: value.Line, Col: value.Col}
	case HashFnExpr:
		return HashFnExpr{Body: copyExpr(value.Body), Line: value.Line, Col: value.Col}
	default:
		return expr
	}
}

func expandCondMacro(args []Expr) (Expr, error) {
	if len(args) == 0 {
		return ListExpr{Elements: []Expr{SymbolExpr{Name: "if"}, SymbolExpr{Name: "false"}, IntExpr{Value: 0}}}, nil
	}
	if len(args)%2 != 0 {
		return nil, fmt.Errorf("cond expects test/expression pairs")
	}

	var acc Expr
	haveElse := false
	for i := len(args) - 2; i >= 0; i -= 2 {
		test := args[i]
		expr := args[i+1]
		if kw, ok := test.(KeywordExpr); ok && kw.Name == "else" {
			if i != len(args)-2 {
				return nil, fmt.Errorf("cond :else must be the final test")
			}
			acc = expr
			haveElse = true
			continue
		}

		if acc == nil {
			acc = ListExpr{Elements: []Expr{
				SymbolExpr{Name: "if"},
				test,
				expr,
			}}
		} else {
			acc = ListExpr{Elements: []Expr{
				SymbolExpr{Name: "if"},
				test,
				expr,
				acc,
			}}
		}
	}
	if acc == nil && haveElse {
		return nil, fmt.Errorf("cond requires at least one non-:else clause")
	}
	if acc == nil {
		return ListExpr{Elements: []Expr{
			SymbolExpr{Name: "if"},
			SymbolExpr{Name: "false"},
			IntExpr{Value: 0},
		}}, nil
	}

	return acc, nil
}

func expandThreadFirstMacro(args []Expr) (Expr, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("-> expects at least one argument")
	}
	acc := copyExpr(args[0])
	for _, step := range args[1:] {
		next, err := threadFirstStep(step, acc)
		if err != nil {
			return nil, err
		}
		acc = next
	}
	return acc, nil
}

func expandThreadLastMacro(args []Expr) (Expr, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("->> expects at least one argument")
	}
	acc := copyExpr(args[0])
	for _, step := range args[1:] {
		next, err := threadLastStep(step, acc)
		if err != nil {
			return nil, err
		}
		acc = next
	}
	return acc, nil
}

func expandSomeThreadFirstMacro(args []Expr) (Expr, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("some-> expects at least one argument")
	}
	acc := copyExpr(args[0])
	for i, step := range args[1:] {
		tempName := fmt.Sprintf("__some_arrow_%d", i)
		tempSym := SymbolExpr{Name: tempName}
		next, err := threadFirstStep(step, tempSym)
		if err != nil {
			return nil, err
		}
		acc = ListExpr{Elements: []Expr{
			SymbolExpr{Name: "let"},
			VectorExpr{Elements: []Expr{
				tempSym,
				acc,
			}},
			ListExpr{Elements: []Expr{
				SymbolExpr{Name: "if"},
				ListExpr{Elements: []Expr{
					SymbolExpr{Name: "="},
					tempSym,
					SymbolExpr{Name: "nil"},
				}},
				SymbolExpr{Name: "nil"},
				next,
			}},
		}}
	}
	return acc, nil
}

func threadFirstStep(step Expr, acc Expr) (Expr, error) {
	switch value := step.(type) {
	case SymbolExpr:
		return ListExpr{Elements: []Expr{value, copyExpr(acc)}}, nil
	case ListExpr:
		if len(value.Elements) == 0 {
			return nil, fmt.Errorf("-> step must not be an empty list")
		}
		out := make([]Expr, 0, len(value.Elements)+1)
		out = append(out, copyExpr(value.Elements[0]))
		out = append(out, copyExpr(acc))
		for _, item := range value.Elements[1:] {
			out = append(out, copyExpr(item))
		}
		return ListExpr{Elements: out}, nil
	default:
		return nil, fmt.Errorf("-> step must be a symbol or list")
	}
}

func threadLastStep(step Expr, acc Expr) (Expr, error) {
	switch value := step.(type) {
	case SymbolExpr:
		return ListExpr{Elements: []Expr{value, copyExpr(acc)}}, nil
	case ListExpr:
		if len(value.Elements) == 0 {
			return nil, fmt.Errorf("->> step must not be an empty list")
		}
		out := make([]Expr, 0, len(value.Elements)+1)
		out = append(out, copyExpr(value.Elements[0]))
		for _, item := range value.Elements[1:] {
			out = append(out, copyExpr(item))
		}
		out = append(out, copyExpr(acc))
		return ListExpr{Elements: out}, nil
	default:
		return nil, fmt.Errorf("->> step must be a symbol or list")
	}
}

func argumentExprForGoCall(args []Expr, ctx compileContext, locals map[string]exprKind) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("expected one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return "", err
	}
	if arg.kind == exprKindValue {
		return fmt.Sprintf("%s.ValueToAny(%s)", runtimeAlias, arg.code), nil
	}
	return arg.code, nil
}

func strArgExprForGoCall(args []Expr, ctx compileContext, locals map[string]exprKind) (string, error) {
	parts, err := strArgParts(args, ctx, locals)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s.Str(%s)", runtimeAlias, strings.Join(parts, ", ")), nil
}

func exprToGo(expr Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	switch arg := expr.(type) {
	case StringExpr:
		return goExpr{code: fmt.Sprintf("%q", arg.Value), kind: exprKindString}, nil
	case IntExpr:
		return goExpr{code: fmt.Sprintf("%s.NewLong(%d)", runtimeAlias, arg.Value), kind: exprKindValue}, nil
	case RatioExpr:
		return goExpr{code: fmt.Sprintf("%s.NewRatio(%d, %d)", runtimeAlias, arg.Numerator, arg.Denominator), kind: exprKindValue}, nil
	case FloatExpr:
		if arg.Raw != "" {
			return goExpr{code: fmt.Sprintf("%s.NewDouble(%s)", runtimeAlias, arg.Raw), kind: exprKindValue}, nil
		}
		return goExpr{code: fmt.Sprintf("%s.NewDouble(%g)", runtimeAlias, arg.Value), kind: exprKindValue}, nil
	case KeywordExpr:
		return goExpr{code: fmt.Sprintf("%s.NewKeyword(%q)", runtimeAlias, arg.Name), kind: exprKindValue}, nil
	case QuotedSymbolExpr:
		return goExpr{code: fmt.Sprintf("%s.NewSymbol(%q)", runtimeAlias, arg.Name), kind: exprKindValue}, nil
	case QuotedListExpr:
		return quotedListExprToGo(arg)
	case VectorExpr:
		return vectorExprToGo(arg.Elements, ctx, locals)
	case MapExpr:
		return mapExprToGo(arg.Entries, ctx, locals)
	case SetExpr:
		return setExprToGo(arg.Elements, ctx, locals)
	case HashFnExpr:
		return hashFnExprToGo(arg, ctx, locals)
	case SymbolExpr:
		if arg.Name == "true" {
			return goExpr{code: fmt.Sprintf("%s.NewBool(true)", runtimeAlias), kind: exprKindValue}, nil
		}
		if arg.Name == "false" {
			return goExpr{code: fmt.Sprintf("%s.NewBool(false)", runtimeAlias), kind: exprKindValue}, nil
		}
		if arg.Name == "nil" {
			return goExpr{code: fmt.Sprintf("%s.NilValue()", runtimeAlias), kind: exprKindValue}, nil
		}
		ident, err := toGoIdentifier(arg.Name)
		if err == nil {
			if locals != nil {
				if kind, ok := locals[ident]; ok {
					return goExpr{code: ident, kind: kind}, nil
				}
			}
			if kind, ok := ctx.globals[ident]; ok {
				return goExpr{code: ident, kind: kind}, nil
			}
			if _, ok := ctx.functions[ident]; ok {
				return goExpr{code: fmt.Sprintf("%s.NewFunction(%s)", runtimeAlias, ident), kind: exprKindValue}, nil
			}
		}
		if isBuiltinFunctionSymbol(arg.Name) {
			return goExpr{code: fmt.Sprintf("%s.BuiltinFunction(%q)", runtimeAlias, arg.Name), kind: exprKindValue}, nil
		}
		if err != nil {
			return goExpr{}, err
		}
		return goExpr{}, fmt.Errorf("unknown symbol %q", arg.Name)
	case ListExpr:
		return listExprToGo(arg, ctx, locals)
	default:
		return goExpr{}, fmt.Errorf("unsupported literal")
	}
}

func quotedListExprToGo(arg QuotedListExpr) (goExpr, error) {
	parts := make([]string, 0, len(arg.Elements))
	for _, item := range arg.Elements {
		code, err := quotedLiteralToValueCode(item)
		if err != nil {
			return goExpr{}, err
		}
		parts = append(parts, code)
	}
	return goExpr{code: fmt.Sprintf("%s.NewList(%s)", runtimeAlias, strings.Join(parts, ", ")), kind: exprKindValue}, nil
}

func quotedLiteralToValueCode(expr Expr) (string, error) {
	switch value := expr.(type) {
	case IntExpr:
		return fmt.Sprintf("%s.NewLong(%d)", runtimeAlias, value.Value), nil
	case RatioExpr:
		return fmt.Sprintf("%s.NewRatio(%d, %d)", runtimeAlias, value.Numerator, value.Denominator), nil
	case FloatExpr:
		if value.Raw != "" {
			return fmt.Sprintf("%s.NewDouble(%s)", runtimeAlias, value.Raw), nil
		}
		return fmt.Sprintf("%s.NewDouble(%g)", runtimeAlias, value.Value), nil
	case KeywordExpr:
		return fmt.Sprintf("%s.NewKeyword(%q)", runtimeAlias, value.Name), nil
	case SymbolExpr:
		return fmt.Sprintf("%s.NewSymbol(%q)", runtimeAlias, value.Name), nil
	case QuotedSymbolExpr:
		return fmt.Sprintf("%s.NewSymbol(%q)", runtimeAlias, value.Name), nil
	case QuotedListExpr:
		out := make([]string, 0, len(value.Elements))
		for _, item := range value.Elements {
			part, err := quotedLiteralToValueCode(item)
			if err != nil {
				return "", err
			}
			out = append(out, part)
		}
		return fmt.Sprintf("%s.NewList(%s)", runtimeAlias, strings.Join(out, ", ")), nil
	case ListExpr:
		out := make([]string, 0, len(value.Elements))
		for _, item := range value.Elements {
			part, err := quotedLiteralToValueCode(item)
			if err != nil {
				return "", err
			}
			out = append(out, part)
		}
		return fmt.Sprintf("%s.NewList(%s)", runtimeAlias, strings.Join(out, ", ")), nil
	case VectorExpr:
		out := make([]string, 0, len(value.Elements))
		for _, item := range value.Elements {
			part, err := quotedLiteralToValueCode(item)
			if err != nil {
				return "", err
			}
			out = append(out, part)
		}
		return fmt.Sprintf("%s.NewArray(%s)", runtimeAlias, strings.Join(out, ", ")), nil
	case MapExpr:
		if len(value.Entries)%2 != 0 {
			return "", fmt.Errorf("quoted map literal expects an even number of forms")
		}
		out := make([]string, 0, len(value.Entries))
		for _, item := range value.Entries {
			part, err := quotedLiteralToValueCode(item)
			if err != nil {
				return "", err
			}
			out = append(out, part)
		}
		return fmt.Sprintf("%s.NewMap(%s)", runtimeAlias, strings.Join(out, ", ")), nil
	case SetExpr:
		out := make([]string, 0, len(value.Elements))
		for _, item := range value.Elements {
			part, err := quotedLiteralToValueCode(item)
			if err != nil {
				return "", err
			}
			out = append(out, part)
		}
		return fmt.Sprintf("%s.NewSet(%s)", runtimeAlias, strings.Join(out, ", ")), nil
	default:
		return "", fmt.Errorf("unsupported quoted literal %T", expr)
	}
}

func isBuiltinFunctionSymbol(name string) bool {
	switch name {
	case "+", "-", "*", "/", "%", "=", "<", ">",
		"first", "fist", "rest", "take", "drop",
		"map", "pmap", "filter", "reduce", "range", "get",
		"assoc", "dissoc", "open-file", "file-to-strings", "rand-int",
		"go-fn", "go-fn-args":
		return true
	default:
		return false
	}
}

func listExprToGo(list ListExpr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(list.Elements) == 0 {
		return goExpr{}, fmt.Errorf("unsupported form")
	}

	head, ok := list.Elements[0].(SymbolExpr)
	if ok {
		switch head.Name {
		case "+":
			return infixExprToGo(list.Elements[1:], runtimeAlias+".Add", ctx, locals)
		case "*":
			return infixExprToGo(list.Elements[1:], runtimeAlias+".Mul", ctx, locals)
		case "-":
			return infixExprToGo(list.Elements[1:], runtimeAlias+".Sub", ctx, locals)
		case "/":
			return infixExprToGo(list.Elements[1:], runtimeAlias+".Div", ctx, locals)
		case "%":
			return modExprToGo(list.Elements[1:], ctx, locals)
		case "=":
			return equalityExprToGo(list.Elements[1:], ctx, locals)
		case "<":
			return comparisonExprToGo(list.Elements[1:], runtimeAlias+".Lt", "<", ctx, locals)
		case ">":
			return comparisonExprToGo(list.Elements[1:], runtimeAlias+".Gt", ">", ctx, locals)
		case "str":
			return strExprToGo(list.Elements[1:], ctx, locals)
		case "println":
			return printlnExprToGo(list.Elements[1:], ctx, locals)
		case "testing":
			return testingExprToGo(list.Elements[1:], ctx, locals)
		case "is":
			return isExprToGo(list, ctx, locals)
		case "if":
			return ifExprToGo(list.Elements[1:], ctx, locals)
		case "do":
			return doExprToGo(list.Elements[1:], ctx, locals)
		case "let":
			return letExprToGo(list.Elements[1:], ctx, locals)
		case "symbol":
			return symbolExprToGo(list.Elements[1:], ctx, locals)
		case "name":
			return nameExprToGo(list.Elements[1:], ctx, locals)
		case "first", "fist":
			return firstExprToGo(list.Elements[1:], ctx, locals)
		case "rest":
			return restExprToGo(list.Elements[1:], ctx, locals)
		case "take":
			return takeExprToGo(list.Elements[1:], ctx, locals)
		case "drop":
			return dropExprToGo(list.Elements[1:], ctx, locals)
		case "map":
			return mapCallExprToGo(list.Elements[1:], ctx, locals)
		case "pmap":
			return pmapCallExprToGo(list.Elements[1:], ctx, locals)
		case "filter":
			return filterCallExprToGo(list.Elements[1:], ctx, locals)
		case "reduce":
			return reduceCallExprToGo(list.Elements[1:], ctx, locals)
		case "range":
			return rangeCallExprToGo(list.Elements[1:], ctx, locals)
		case "rand-int":
			return randIntExprToGo(list.Elements[1:], ctx, locals)
		case "assoc":
			return assocExprToGo(list.Elements[1:], ctx, locals)
		case "dissoc":
			return dissocExprToGo(list.Elements[1:], ctx, locals)
		case "to-json":
			return toJSONExprToGo(list.Elements[1:], ctx, locals)
		case "from-json":
			return fromJSONExprToGo(list.Elements[1:], ctx, locals)
		case "open-file":
			return openFileExprToGo(list.Elements[1:], ctx, locals)
		case "file-to-strings":
			return fileToStringsExprToGo(list.Elements[1:], ctx, locals)
		case "go-fn":
			return goFnExprToGo(list.Elements[1:], ctx, locals)
		case "go-fn-args":
			return goFnArgsExprToGo(list.Elements[1:], ctx, locals)
		case "fn":
			return fnExprToGo(list.Elements[1:], ctx, locals)
		}
	}

	return callExprToGo(list.Elements[0], list.Elements[1:], ctx, locals)
}

func callExprToGo(calleeExpr Expr, argsExpr []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if calleeSymbol, ok := calleeExpr.(SymbolExpr); ok &&
		calleeSymbol.Name == ctx.selfFunctionName &&
		len(argsExpr) == ctx.selfFunctionArity {
		args := make([]string, 0, len(argsExpr))
		for _, item := range argsExpr {
			part, err := exprToGo(item, ctx, locals)
			if err != nil {
				return goExpr{}, err
			}
			valueCode, err := functionArgToValueCode(part)
			if err != nil {
				return goExpr{}, err
			}
			args = append(args, valueCode)
		}
		return goExpr{code: fmt.Sprintf("%s(%s)", ctx.selfArityName, strings.Join(args, ", ")), kind: exprKindValue}, nil
	}

	callee, err := exprToGo(calleeExpr, ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	if callee.kind != exprKindValue {
		return goExpr{}, fmt.Errorf("first element in list must evaluate to a function")
	}

	args := make([]string, 0, len(argsExpr))
	for _, item := range argsExpr {
		part, err := exprToGo(item, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		valueCode, err := functionArgToValueCode(part)
		if err != nil {
			return goExpr{}, err
		}
		args = append(args, valueCode)
	}
	return goExpr{code: fmt.Sprintf("%s.Call(%s)", runtimeAlias, strings.Join(append([]string{callee.code}, args...), ", ")), kind: exprKindValue}, nil
}

func functionArgToValueCode(arg goExpr) (string, error) {
	switch arg.kind {
	case exprKindValue:
		return arg.code, nil
	case exprKindString:
		return fmt.Sprintf("%s.NewString(%s)", runtimeAlias, arg.code), nil
	default:
		return "", fmt.Errorf("function argument must evaluate to Value")
	}
}

func infixExprToGo(args []Expr, runtimeOp string, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) < 2 {
		return goExpr{}, fmt.Errorf("%s expects at least two arguments", runtimeOp)
	}

	first, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	if first.kind != exprKindValue {
		return goExpr{}, fmt.Errorf("numeric operator expects numeric Value arguments")
	}

	acc := first.code
	for _, item := range args[1:] {
		part, err := exprToGo(item, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("numeric operator expects numeric Value arguments")
		}
		acc = fmt.Sprintf("%s(%s, %s)", runtimeOp, acc, part.code)
	}
	return goExpr{code: acc, kind: exprKindValue}, nil
}

func modExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 2 {
		return goExpr{}, fmt.Errorf("%% expects exactly two arguments")
	}
	left, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	if left.kind != exprKindValue {
		return goExpr{}, fmt.Errorf("%% expects numeric Value arguments")
	}
	right, err := exprToGo(args[1], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	if right.kind != exprKindValue {
		return goExpr{}, fmt.Errorf("%% expects numeric Value arguments")
	}
	return goExpr{code: fmt.Sprintf("%s.Mod(%s, %s)", runtimeAlias, left.code, right.code), kind: exprKindValue}, nil
}

func equalityExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	boolExpr, err := comparisonExprToGo(args, runtimeAlias+".Eq", "=", ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	return goExpr{code: fmt.Sprintf("%s.NewBool(%s)", runtimeAlias, boolExpr.code), kind: exprKindValue}, nil
}

func comparisonExprToGo(args []Expr, runtimeOp, operator string, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) < 2 {
		return goExpr{}, fmt.Errorf("%s expects at least two arguments", operator)
	}

	parts := make([]string, 0, len(args))
	for _, item := range args {
		part, err := exprToGo(item, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("%s expects numeric Value arguments", operator)
		}
		parts = append(parts, part.code)
	}

	checks := make([]string, 0, len(parts)-1)
	for i := 0; i < len(parts)-1; i++ {
		checks = append(checks, fmt.Sprintf("%s(%s, %s)", runtimeOp, parts[i], parts[i+1]))
	}
	return goExpr{code: strings.Join(checks, " && "), kind: exprKindBool}, nil
}

func ifExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 2 && len(args) != 3 {
		return goExpr{}, fmt.Errorf("if expects test, true branch, and optional false branch")
	}

	condExpr, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	condition, err := truthyExprToGo(condExpr)
	if err != nil {
		return goExpr{}, err
	}

	trueExpr, err := exprToGo(args[1], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}

	var falseExpr goExpr
	if len(args) == 3 {
		falseExpr, err = exprToGo(args[2], ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if falseExpr.kind != trueExpr.kind {
			return goExpr{}, fmt.Errorf("if branches must produce the same kind")
		}
	} else {
		if trueExpr.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("if without false branch currently requires a Value result")
		}
		falseExpr = goExpr{code: runtimeAlias + ".NilValue()", kind: exprKindValue}
	}

	typeName, err := goTypeForExprKind(trueExpr.kind)
	if err != nil {
		return goExpr{}, err
	}

	var out strings.Builder
	fmt.Fprintf(&out, "func() %s {\n", typeName)
	fmt.Fprintf(&out, "\tif %s {\n", condition)
	fmt.Fprintf(&out, "\t\treturn %s\n", trueExpr.code)
	out.WriteString("\t}\n")
	fmt.Fprintf(&out, "\treturn %s\n", falseExpr.code)
	out.WriteString("}()")

	return goExpr{code: out.String(), kind: trueExpr.kind}, nil
}

func doExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) == 0 {
		return goExpr{code: runtimeAlias + ".NilValue()", kind: exprKindValue}, nil
	}

	compiled := make([]goExpr, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		compiled = append(compiled, part)
	}

	result := compiled[len(compiled)-1]
	typeName, err := goTypeForExprKind(result.kind)
	if err != nil {
		return goExpr{}, err
	}

	var out strings.Builder
	fmt.Fprintf(&out, "func() %s {\n", typeName)
	for i := 0; i < len(compiled)-1; i++ {
		fmt.Fprintf(&out, "\t_ = %s\n", compiled[i].code)
	}
	fmt.Fprintf(&out, "\treturn %s\n", result.code)
	out.WriteString("}()")
	return goExpr{code: out.String(), kind: result.kind}, nil
}

func letExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) < 1 {
		return goExpr{}, fmt.Errorf("let expects a binding vector")
	}

	bindingsExpr, ok := args[0].(VectorExpr)
	if !ok {
		return goExpr{}, fmt.Errorf("let expects a binding vector")
	}
	if len(bindingsExpr.Elements)%2 != 0 {
		return goExpr{}, fmt.Errorf("let binding vector expects name/value pairs")
	}

	localKinds := make(map[string]exprKind, len(locals)+len(bindingsExpr.Elements)/2)
	for name, kind := range locals {
		localKinds[name] = kind
	}

	bindings := make([]string, 0, len(bindingsExpr.Elements)*2)
	tempCounter := 0
	declared := make(map[string]struct{}, len(bindingsExpr.Elements))
	for i := 0; i < len(bindingsExpr.Elements); i += 2 {
		valueExpr, err := exprToGo(bindingsExpr.Elements[i+1], ctx, localKinds)
		if err != nil {
			return goExpr{}, err
		}
		if valueExpr.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("let binding value must evaluate to Value")
		}

		sourceName := fmt.Sprintf("__bind%d", tempCounter)
		tempCounter++
		bindings = append(bindings, fmt.Sprintf("\tvar %s = %s\n", sourceName, valueExpr.code))

		emitter := newDestructureEmitter(ctx, localKinds, &bindings, &tempCounter, declared)
		if err := emitter.bind(bindingsExpr.Elements[i], sourceName); err != nil {
			return goExpr{}, err
		}
	}

	bodyExprs := args[1:]
	if len(bodyExprs) == 0 {
		if len(bindingsExpr.Elements) == 0 {
			return goExpr{}, fmt.Errorf("let without body requires at least one binding")
		}
		lastPattern := bindingsExpr.Elements[len(bindingsExpr.Elements)-2]
		if sym, ok := lastPattern.(SymbolExpr); ok {
			bodyExprs = []Expr{sym}
		} else if kw, ok := lastPattern.(KeywordExpr); ok {
			bodyExprs = []Expr{kw}
		} else {
			bodyExprs = []Expr{SymbolExpr{Name: "nil"}}
		}
	}

	compiledBody := make([]goExpr, 0, len(bodyExprs))
	for _, bodyExpr := range bodyExprs {
		compiled, err := exprToGo(bodyExpr, ctx, localKinds)
		if err != nil {
			return goExpr{}, err
		}
		compiledBody = append(compiledBody, compiled)
	}

	result := compiledBody[len(compiledBody)-1]
	typeName, err := goTypeForExprKind(result.kind)
	if err != nil {
		return goExpr{}, err
	}

	var out strings.Builder
	fmt.Fprintf(&out, "func() %s {\n", typeName)
	for _, binding := range bindings {
		out.WriteString(binding)
	}
	for i := 0; i < len(compiledBody)-1; i++ {
		fmt.Fprintf(&out, "\t_ = %s\n", compiledBody[i].code)
	}
	fmt.Fprintf(&out, "\treturn %s\n", result.code)
	out.WriteString("}()")

	return goExpr{code: out.String(), kind: result.kind}, nil
}

type destructureEmitter struct {
	ctx         compileContext
	locals      map[string]exprKind
	lines       *[]string
	tempCounter *int
	declared    map[string]struct{}
}

type destructureKeyBinding struct {
	keyExpr     string
	bindingExpr Expr
}

func newDestructureEmitter(
	ctx compileContext,
	locals map[string]exprKind,
	lines *[]string,
	tempCounter *int,
	declared map[string]struct{},
) destructureEmitter {
	return destructureEmitter{
		ctx:         ctx,
		locals:      locals,
		lines:       lines,
		tempCounter: tempCounter,
		declared:    declared,
	}
}

func (e destructureEmitter) bind(pattern Expr, sourceCode string) error {
	switch p := pattern.(type) {
	case SymbolExpr:
		if p.Name == "&" {
			return fmt.Errorf("unexpected & in binding")
		}
		return e.bindSymbol(p, sourceCode)
	case VectorExpr:
		return e.bindVector(p, sourceCode)
	case MapExpr:
		return e.bindMap(p, sourceCode)
	default:
		return fmt.Errorf("unsupported binding form %T", pattern)
	}
}

func (e destructureEmitter) bindSymbol(sym SymbolExpr, sourceCode string) error {
	if sym.Name == "" {
		return fmt.Errorf("binding symbol cannot be empty")
	}
	name, err := toGoIdentifier(sym.Name)
	if err != nil {
		return err
	}
	if _, exists := e.declared[name]; exists {
		return fmt.Errorf("duplicate binding %q", sym.Name)
	}
	e.declared[name] = struct{}{}
	e.locals[name] = exprKindValue
	*e.lines = append(*e.lines, fmt.Sprintf("\tvar %s = %s\n", name, sourceCode))
	return nil
}

func (e destructureEmitter) bindVector(pattern VectorExpr, sourceCode string) error {
	positionals := make([]Expr, 0, len(pattern.Elements))
	var restPattern Expr
	var asPattern Expr
	seenRest := false

	for i := 0; i < len(pattern.Elements); i++ {
		if kw, ok := pattern.Elements[i].(KeywordExpr); ok && kw.Name == "as" {
			if i+1 >= len(pattern.Elements) {
				return fmt.Errorf("vector destructuring :as expects a binding form")
			}
			if asPattern != nil {
				return fmt.Errorf("vector destructuring supports only one :as")
			}
			asPattern = pattern.Elements[i+1]
			i++
			continue
		}
		if sym, ok := pattern.Elements[i].(SymbolExpr); ok && sym.Name == "&" {
			if seenRest {
				return fmt.Errorf("vector destructuring supports only one & binding")
			}
			if i+1 >= len(pattern.Elements) {
				return fmt.Errorf("vector destructuring & expects a binding form")
			}
			restPattern = pattern.Elements[i+1]
			seenRest = true
			i++
			continue
		}
		if seenRest {
			return fmt.Errorf("vector destructuring only allows :as after & binding")
		}
		positionals = append(positionals, pattern.Elements[i])
	}

	if asPattern != nil {
		if err := e.bind(asPattern, sourceCode); err != nil {
			return err
		}
	}

	current := sourceCode
	for _, positional := range positionals {
		nextName := e.freshTemp("__dseq")
		*e.lines = append(*e.lines, fmt.Sprintf("\tvar %s = %s.SeqFirst(%s)\n", nextName, runtimeAlias, current))
		if err := e.bind(positional, nextName); err != nil {
			return err
		}
		current = fmt.Sprintf("%s.SeqRest(%s)", runtimeAlias, current)
	}

	if restPattern != nil {
		if err := e.bind(restPattern, current); err != nil {
			return err
		}
	}
	return nil
}

func (e destructureEmitter) bindMap(pattern MapExpr, sourceCode string) error {
	bindings := make([]destructureKeyBinding, 0, len(pattern.Entries)/2)
	var asPattern Expr
	defaults := make(map[string]Expr)

	for i := 0; i < len(pattern.Entries); i += 2 {
		if i+1 >= len(pattern.Entries) {
			return fmt.Errorf("map destructuring expects key/value pairs")
		}
		key := pattern.Entries[i]
		value := pattern.Entries[i+1]

		if kw, ok := key.(KeywordExpr); ok {
			switch kw.Name {
			case "as":
				asPattern = value
				continue
			case "or":
				defaultMap, ok := value.(MapExpr)
				if !ok {
					return fmt.Errorf("map destructuring :or expects a map")
				}
				def, err := collectDestructureDefaults(defaultMap)
				if err != nil {
					return err
				}
				for name, expr := range def {
					defaults[name] = expr
				}
				continue
			case "keys":
				entries, err := expandMapDestructureKeys(value)
				if err != nil {
					return err
				}
				bindings = append(bindings, entries...)
				continue
			case "syms":
				entries, err := expandMapDestructureSyms(value)
				if err != nil {
					return err
				}
				bindings = append(bindings, entries...)
				continue
			case "strs":
				entries, err := expandMapDestructureStrs(value)
				if err != nil {
					return err
				}
				bindings = append(bindings, entries...)
				continue
			}
		}

		keyExpr, err := compileDestructureMapKeyToValue(key)
		if err != nil {
			return err
		}
		bindings = append(bindings, destructureKeyBinding{keyExpr: keyExpr, bindingExpr: value})
	}

	if asPattern != nil {
		if err := e.bind(asPattern, sourceCode); err != nil {
			return err
		}
	}

	for _, binding := range bindings {
		valueExpr := fmt.Sprintf("%s.Get(%s, %s)", runtimeAlias, sourceCode, binding.keyExpr)
		if sym, ok := binding.bindingExpr.(SymbolExpr); ok {
			if defExpr, exists := defaults[sym.Name]; exists {
				defaultCode, err := exprToGo(defExpr, e.ctx, e.locals)
				if err != nil {
					return fmt.Errorf("map destructuring default for %q: %w", sym.Name, err)
				}
				if defaultCode.kind != exprKindValue {
					return fmt.Errorf("map destructuring default for %q must evaluate to Value", sym.Name)
				}
				valueExpr = fmt.Sprintf("%s.Get(%s, %s, %s)", runtimeAlias, sourceCode, binding.keyExpr, defaultCode.code)
			}
		}
		if err := e.bind(binding.bindingExpr, valueExpr); err != nil {
			return err
		}
	}

	return nil
}

func (e destructureEmitter) freshTemp(prefix string) string {
	name := fmt.Sprintf("%s%d", prefix, *e.tempCounter)
	*e.tempCounter = *e.tempCounter + 1
	return name
}

func collectDestructureDefaults(defaults MapExpr) (map[string]Expr, error) {
	if len(defaults.Entries)%2 != 0 {
		return nil, fmt.Errorf("map destructuring :or expects key/value pairs")
	}
	out := make(map[string]Expr, len(defaults.Entries)/2)
	for i := 0; i < len(defaults.Entries); i += 2 {
		name, err := destructureDefaultName(defaults.Entries[i])
		if err != nil {
			return nil, err
		}
		out[name] = defaults.Entries[i+1]
	}
	return out, nil
}

func destructureDefaultName(expr Expr) (string, error) {
	switch value := expr.(type) {
	case SymbolExpr:
		if value.Name == "" {
			return "", fmt.Errorf("map destructuring :or key cannot be empty")
		}
		return value.Name, nil
	case KeywordExpr:
		if value.Name == "" {
			return "", fmt.Errorf("map destructuring :or key cannot be empty")
		}
		return value.Name, nil
	default:
		return "", fmt.Errorf("map destructuring :or keys must be symbols or keywords")
	}
}

func expandMapDestructureKeys(expr Expr) ([]destructureKeyBinding, error) {
	vector, ok := expr.(VectorExpr)
	if !ok {
		return nil, fmt.Errorf("map destructuring :keys expects a vector")
	}
	out := make([]destructureKeyBinding, 0, len(vector.Elements))
	for _, entry := range vector.Elements {
		sym, ok := entry.(SymbolExpr)
		if !ok || sym.Name == "" {
			return nil, fmt.Errorf("map destructuring :keys entries must be symbols")
		}
		out = append(out, destructureKeyBinding{
			keyExpr:     fmt.Sprintf("%s.NewKeyword(%q)", runtimeAlias, sym.Name),
			bindingExpr: sym,
		})
	}
	return out, nil
}

func expandMapDestructureSyms(expr Expr) ([]destructureKeyBinding, error) {
	vector, ok := expr.(VectorExpr)
	if !ok {
		return nil, fmt.Errorf("map destructuring :syms expects a vector")
	}
	out := make([]destructureKeyBinding, 0, len(vector.Elements))
	for _, entry := range vector.Elements {
		sym, ok := entry.(SymbolExpr)
		if !ok || sym.Name == "" {
			return nil, fmt.Errorf("map destructuring :syms entries must be symbols")
		}
		out = append(out, destructureKeyBinding{
			keyExpr:     fmt.Sprintf("%s.NewSymbol(%q)", runtimeAlias, sym.Name),
			bindingExpr: sym,
		})
	}
	return out, nil
}

func expandMapDestructureStrs(expr Expr) ([]destructureKeyBinding, error) {
	vector, ok := expr.(VectorExpr)
	if !ok {
		return nil, fmt.Errorf("map destructuring :strs expects a vector")
	}
	out := make([]destructureKeyBinding, 0, len(vector.Elements))
	for _, entry := range vector.Elements {
		str, ok := entry.(StringExpr)
		if !ok {
			return nil, fmt.Errorf("map destructuring :strs entries must be strings")
		}
		name, err := toGoIdentifier(str.Value)
		if err != nil {
			return nil, fmt.Errorf("map destructuring :strs entry %q is not a valid symbol name", str.Value)
		}
		sym := SymbolExpr{Name: name}
		out = append(out, destructureKeyBinding{
			keyExpr:     fmt.Sprintf("%s.NewString(%q)", runtimeAlias, str.Value),
			bindingExpr: sym,
		})
	}
	return out, nil
}

func compileDestructureMapKeyToValue(expr Expr) (string, error) {
	switch value := expr.(type) {
	case KeywordExpr:
		return fmt.Sprintf("%s.NewKeyword(%q)", runtimeAlias, value.Name), nil
	case SymbolExpr:
		return fmt.Sprintf("%s.NewSymbol(%q)", runtimeAlias, value.Name), nil
	case QuotedSymbolExpr:
		return fmt.Sprintf("%s.NewSymbol(%q)", runtimeAlias, value.Name), nil
	case IntExpr:
		return fmt.Sprintf("%s.NewLong(%d)", runtimeAlias, value.Value), nil
	case FloatExpr:
		if value.Raw != "" {
			return fmt.Sprintf("%s.NewDouble(%s)", runtimeAlias, value.Raw), nil
		}
		return fmt.Sprintf("%s.NewDouble(%g)", runtimeAlias, value.Value), nil
	case StringExpr:
		return fmt.Sprintf("%s.NewString(%q)", runtimeAlias, value.Value), nil
	default:
		return "", fmt.Errorf("unsupported map destructuring key %T", expr)
	}
}

func symbolExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("symbol expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	switch arg.kind {
	case exprKindString, exprKindValue:
		return goExpr{code: fmt.Sprintf("%s.Symbol(%s)", runtimeAlias, arg.code), kind: exprKindValue}, nil
	default:
		return goExpr{}, fmt.Errorf("symbol expects a string, symbol, or keyword argument")
	}
}

func nameExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("name expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	switch arg.kind {
	case exprKindString, exprKindValue:
		return goExpr{code: fmt.Sprintf("%s.Name(%s)", runtimeAlias, arg.code), kind: exprKindString}, nil
	default:
		return goExpr{}, fmt.Errorf("name expects a string, symbol, or keyword argument")
	}
}

func firstExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("first expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	if arg.kind != exprKindValue {
		return goExpr{}, fmt.Errorf("first expects an argument that evaluates to Value")
	}
	return goExpr{code: fmt.Sprintf("%s.First(%s)", runtimeAlias, arg.code), kind: exprKindValue}, nil
}

func restExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("rest expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	if arg.kind != exprKindValue {
		return goExpr{}, fmt.Errorf("rest expects an argument that evaluates to Value")
	}
	return goExpr{code: fmt.Sprintf("%s.Rest(%s)", runtimeAlias, arg.code), kind: exprKindValue}, nil
}

func takeExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 2 {
		return goExpr{}, fmt.Errorf("take expects count and sequence")
	}
	countExpr, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	seqExpr, err := exprToGo(args[1], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	if countExpr.kind != exprKindValue || seqExpr.kind != exprKindValue {
		return goExpr{}, fmt.Errorf("take arguments must evaluate to Value")
	}
	return goExpr{code: fmt.Sprintf("%s.Take(%s, %s)", runtimeAlias, countExpr.code, seqExpr.code), kind: exprKindValue}, nil
}

func dropExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 2 {
		return goExpr{}, fmt.Errorf("drop expects count and sequence")
	}
	countExpr, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	seqExpr, err := exprToGo(args[1], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	if countExpr.kind != exprKindValue || seqExpr.kind != exprKindValue {
		return goExpr{}, fmt.Errorf("drop arguments must evaluate to Value")
	}
	return goExpr{code: fmt.Sprintf("%s.Drop(%s, %s)", runtimeAlias, countExpr.code, seqExpr.code), kind: exprKindValue}, nil
}

func toJSONExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("to-json expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	if arg.kind != exprKindValue {
		return goExpr{}, fmt.Errorf("to-json expects an argument that evaluates to Value")
	}
	return goExpr{code: fmt.Sprintf("%s.ToJSON(%s)", runtimeAlias, arg.code), kind: exprKindString}, nil
}

func fromJSONExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("from-json expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	if arg.kind != exprKindString {
		return goExpr{}, fmt.Errorf("from-json expects a string argument")
	}
	return goExpr{code: fmt.Sprintf("%s.FromJSON(%s)", runtimeAlias, arg.code), kind: exprKindValue}, nil
}

func openFileExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("open-file expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	switch arg.kind {
	case exprKindString:
		return goExpr{code: fmt.Sprintf("%s.OpenFile(%s)", runtimeAlias, arg.code), kind: exprKindValue}, nil
	case exprKindValue:
		return goExpr{code: fmt.Sprintf("%s.OpenFile(%s.Name(%s))", runtimeAlias, runtimeAlias, arg.code), kind: exprKindValue}, nil
	default:
		return goExpr{}, fmt.Errorf("open-file expects a string, symbol, or keyword argument")
	}
}

func fileToStringsExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("file-to-strings expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	switch arg.kind {
	case exprKindString:
		return goExpr{code: fmt.Sprintf("%s.FileToStringsPath(%s)", runtimeAlias, arg.code), kind: exprKindValue}, nil
	case exprKindValue:
		return goExpr{code: fmt.Sprintf("%s.FileToStrings(%s)", runtimeAlias, arg.code), kind: exprKindValue}, nil
	default:
		return goExpr{}, fmt.Errorf("file-to-strings expects a string, symbol, or keyword argument")
	}
}

func goFnExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("go-fn expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	switch arg.kind {
	case exprKindString:
		return goExpr{code: fmt.Sprintf("%s.GoFunction(%s)", runtimeAlias, arg.code), kind: exprKindValue}, nil
	case exprKindValue:
		return goExpr{code: fmt.Sprintf("%s.GoFunction(%s.Name(%s))", runtimeAlias, runtimeAlias, arg.code), kind: exprKindValue}, nil
	default:
		return goExpr{}, fmt.Errorf("go-fn expects a string, symbol, or keyword argument")
	}
}

func goFnArgsExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("go-fn-args expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	switch arg.kind {
	case exprKindString:
		return goExpr{code: fmt.Sprintf("%s.GoFunctionArgs(%s)", runtimeAlias, arg.code), kind: exprKindValue}, nil
	case exprKindValue:
		return goExpr{code: fmt.Sprintf("%s.GoFunctionArgs(%s.Name(%s))", runtimeAlias, runtimeAlias, arg.code), kind: exprKindValue}, nil
	default:
		return goExpr{}, fmt.Errorf("go-fn-args expects a string, symbol, or keyword argument")
	}
}

func mapCallExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) < 2 {
		return goExpr{}, fmt.Errorf("map expects function and at least one sequence")
	}

	parts := make([]string, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("map arguments must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.Map(%s)", runtimeAlias, strings.Join(parts, ", ")), kind: exprKindValue}, nil
}

func pmapCallExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) < 2 {
		return goExpr{}, fmt.Errorf("pmap expects function and at least one sequence")
	}

	parts := make([]string, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("pmap arguments must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.PMap(%s)", runtimeAlias, strings.Join(parts, ", ")), kind: exprKindValue}, nil
}

func filterCallExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 2 {
		return goExpr{}, fmt.Errorf("filter expects function and one sequence")
	}

	parts := make([]string, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("filter arguments must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.Filter(%s)", runtimeAlias, strings.Join(parts, ", ")), kind: exprKindValue}, nil
}

func reduceCallExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 2 && len(args) != 3 {
		return goExpr{}, fmt.Errorf("reduce expects function and collection, or function, initial value, and collection")
	}

	parts := make([]string, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("reduce arguments must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.Reduce(%s)", runtimeAlias, strings.Join(parts, ", ")), kind: exprKindValue}, nil
}

func rangeCallExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 0 && len(args) != 1 && len(args) != 2 {
		return goExpr{}, fmt.Errorf("range expects zero, one, or two arguments")
	}
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("range arguments must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.Range(%s)", runtimeAlias, strings.Join(parts, ", ")), kind: exprKindValue}, nil
}

func randIntExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("rand-int expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	if arg.kind != exprKindValue {
		return goExpr{}, fmt.Errorf("rand-int expects an argument that evaluates to Value")
	}
	return goExpr{code: fmt.Sprintf("%s.RandInt(%s)", runtimeAlias, arg.code), kind: exprKindValue}, nil
}

func fnExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 2 {
		return goExpr{}, fmt.Errorf("fn expects parameter vector and body expression")
	}
	paramsExpr, ok := args[0].(VectorExpr)
	if !ok {
		return goExpr{}, fmt.Errorf("fn expects a parameter vector")
	}
	return compileLambda(paramsExpr, args[1], ctx, locals, "fn")
}

func hashFnExprToGo(hashFn HashFnExpr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	maxParam := 0
	collectHashFnPlaceholders(hashFn.Body, &maxParam)

	params := make([]Expr, 0, maxParam)
	for i := 1; i <= maxParam; i++ {
		params = append(params, SymbolExpr{Name: hashFnParamName(i)})
	}
	body := replaceHashFnPlaceholders(hashFn.Body)
	return compileLambda(VectorExpr{Elements: params}, body, ctx, locals, "#()")
}

func collectHashFnPlaceholders(expr Expr, maxParam *int) {
	switch value := expr.(type) {
	case SymbolExpr:
		if index, ok := hashFnPlaceholderIndex(value.Name); ok && index > *maxParam {
			*maxParam = index
		}
	case ListExpr:
		for _, item := range value.Elements {
			collectHashFnPlaceholders(item, maxParam)
		}
	case VectorExpr:
		for _, item := range value.Elements {
			collectHashFnPlaceholders(item, maxParam)
		}
	case MapExpr:
		for _, item := range value.Entries {
			collectHashFnPlaceholders(item, maxParam)
		}
	case SetExpr:
		for _, item := range value.Elements {
			collectHashFnPlaceholders(item, maxParam)
		}
	case HashFnExpr:
		// Nested #() has its own placeholder scope.
	}
}

func replaceHashFnPlaceholders(expr Expr) Expr {
	switch value := expr.(type) {
	case SymbolExpr:
		if index, ok := hashFnPlaceholderIndex(value.Name); ok {
			return SymbolExpr{Name: hashFnParamName(index)}
		}
		return value
	case ListExpr:
		out := make([]Expr, 0, len(value.Elements))
		for _, item := range value.Elements {
			out = append(out, replaceHashFnPlaceholders(item))
		}
		return ListExpr{Elements: out, Line: value.Line, Col: value.Col}
	case VectorExpr:
		out := make([]Expr, 0, len(value.Elements))
		for _, item := range value.Elements {
			out = append(out, replaceHashFnPlaceholders(item))
		}
		return VectorExpr{Elements: out, Line: value.Line, Col: value.Col}
	case MapExpr:
		out := make([]Expr, 0, len(value.Entries))
		for _, item := range value.Entries {
			out = append(out, replaceHashFnPlaceholders(item))
		}
		return MapExpr{Entries: out, Line: value.Line, Col: value.Col}
	case SetExpr:
		out := make([]Expr, 0, len(value.Elements))
		for _, item := range value.Elements {
			out = append(out, replaceHashFnPlaceholders(item))
		}
		return SetExpr{Elements: out, Line: value.Line, Col: value.Col}
	case HashFnExpr:
		return value
	default:
		return expr
	}
}

func hashFnPlaceholderIndex(name string) (int, bool) {
	if name == "%" {
		return 1, true
	}
	if strings.HasPrefix(name, "%") && len(name) > 1 {
		index, err := strconv.Atoi(name[1:])
		if err == nil && index > 0 {
			return index, true
		}
	}
	return 0, false
}

func hashFnParamName(index int) string {
	return fmt.Sprintf("__p%d", index)
}

func compileLambda(paramsExpr VectorExpr, bodyExpr Expr, ctx compileContext, locals map[string]exprKind, label string) (goExpr, error) {
	params, localKinds, localInits, hasRest, err := bindLambdaParams(paramsExpr, ctx, locals, label)
	if err != nil {
		return goExpr{}, err
	}

	body, err := exprToGo(bodyExpr, ctx, localKinds)
	if err != nil {
		return goExpr{}, err
	}
	if body.kind != exprKindValue {
		return goExpr{}, fmt.Errorf("%s body must evaluate to Value", label)
	}

	def := functionDef{
		goName:     label,
		params:     params,
		localInits: localInits,
		body:       body.code,
	}
	if hasRest {
		def.arityName = label
		def.variadicName = label
		def.hasRest = true
	}
	return goExpr{code: fmt.Sprintf("%s.NewFunction(%s)", runtimeAlias, renderFunctionLiteral(def)), kind: exprKindValue}, nil
}

func bindLambdaParams(
	paramsExpr VectorExpr,
	ctx compileContext,
	locals map[string]exprKind,
	label string,
) ([]string, map[string]exprKind, []string, bool, error) {
	params := make([]string, 0, len(paramsExpr.Elements))
	localKinds := make(map[string]exprKind, len(locals)+len(paramsExpr.Elements))
	for name, kind := range locals {
		localKinds[name] = kind
	}
	declared := make(map[string]struct{}, len(paramsExpr.Elements))
	localInits := make([]string, 0, len(paramsExpr.Elements))
	tempCounter := 0
	hasRest := false

	for idx := 0; idx < len(paramsExpr.Elements); idx++ {
		paramExpr := paramsExpr.Elements[idx]
		if sym, ok := paramExpr.(SymbolExpr); ok && sym.Name == "&" {
			if hasRest {
				return nil, nil, nil, false, fmt.Errorf("%s parameters support only one & binding", label)
			}
			if idx+1 >= len(paramsExpr.Elements) {
				return nil, nil, nil, false, fmt.Errorf("%s & expects a binding form", label)
			}
			hasRest = true
			restPattern := paramsExpr.Elements[idx+1]
			restName := fmt.Sprintf("__rest%d", tempCounter)
			tempCounter++
			restSource := fmt.Sprintf("%s.NewArray(args[%d:]...)", runtimeAlias, len(params))
			if _, ok := restPattern.(MapExpr); ok {
				restSource = fmt.Sprintf("%s.NewMap(args[%d:]...)", runtimeAlias, len(params))
			}
			localKinds[restName] = exprKindValue
			localInits = append(localInits, fmt.Sprintf("\tvar %s = %s\n", restName, restSource))
			emitter := newDestructureEmitter(ctx, localKinds, &localInits, &tempCounter, declared)
			if err := emitter.bind(restPattern, restName); err != nil {
				return nil, nil, nil, false, fmt.Errorf("%s parameter %d: %w", label, idx+2, err)
			}
			break
		}

		if sym, ok := paramExpr.(SymbolExpr); ok && sym.Name != "" {
			goParam, err := toGoIdentifier(sym.Name)
			if err != nil {
				return nil, nil, nil, false, err
			}
			if _, exists := declared[goParam]; exists {
				return nil, nil, nil, false, fmt.Errorf("duplicate parameter %q", sym.Name)
			}
			declared[goParam] = struct{}{}
			localKinds[goParam] = exprKindValue
			params = append(params, goParam)
			continue
		}

		paramName := fmt.Sprintf("__arg%d", idx)
		params = append(params, paramName)
		emitter := newDestructureEmitter(ctx, localKinds, &localInits, &tempCounter, declared)
		if err := emitter.bind(paramExpr, paramName); err != nil {
			return nil, nil, nil, false, fmt.Errorf("%s parameter %d: %w", label, idx+1, err)
		}
	}

	return params, localKinds, localInits, hasRest, nil
}

func mapExprToGo(entries []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(entries)%2 != 0 {
		return goExpr{}, fmt.Errorf("map literal expects key/value pairs")
	}

	parts := make([]string, 0, len(entries))
	for _, entry := range entries {
		part, err := exprToGo(entry, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind == exprKindString {
			part = goExpr{code: fmt.Sprintf("%s.NewString(%s)", runtimeAlias, part.code), kind: exprKindValue}
		}
		if part.kind != exprKindValue {
			return goExpr{}, exprError(entry, "map literal entries must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.NewMap(%s)", runtimeAlias, strings.Join(parts, ", ")), kind: exprKindValue}, nil
}

func vectorExprToGo(elements []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	parts := make([]string, 0, len(elements))
	for _, element := range elements {
		part, err := exprToGo(element, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind == exprKindString {
			part = goExpr{code: fmt.Sprintf("%s.NewString(%s)", runtimeAlias, part.code), kind: exprKindValue}
		}
		if part.kind != exprKindValue {
			return goExpr{}, exprError(element, "vector literal entries must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.NewArray(%s)", runtimeAlias, strings.Join(parts, ", ")), kind: exprKindValue}, nil
}

func setExprToGo(elements []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	parts := make([]string, 0, len(elements))
	for _, element := range elements {
		part, err := exprToGo(element, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind == exprKindString {
			part = goExpr{code: fmt.Sprintf("%s.NewString(%s)", runtimeAlias, part.code), kind: exprKindValue}
		}
		if part.kind != exprKindValue {
			return goExpr{}, exprError(element, "set literal entries must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.NewSet(%s)", runtimeAlias, strings.Join(parts, ", ")), kind: exprKindValue}, nil
}

func assocExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) < 3 || len(args)%2 == 0 {
		return goExpr{}, fmt.Errorf("assoc expects collection and key/value pairs")
	}

	parts := make([]string, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("assoc arguments must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.MapAssoc(%s)", runtimeAlias, strings.Join(parts, ", ")), kind: exprKindValue}, nil
}

func dissocExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) < 1 {
		return goExpr{}, fmt.Errorf("dissoc expects a collection argument")
	}

	parts := make([]string, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("dissoc arguments must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.MapDissoc(%s)", runtimeAlias, strings.Join(parts, ", ")), kind: exprKindValue}, nil
}

func strExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	parts, err := strArgParts(args, ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	return goExpr{code: fmt.Sprintf("%s.Str(%s)", runtimeAlias, strings.Join(parts, ", ")), kind: exprKindString}, nil
}

func printlnExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	parts, err := strArgParts(args, ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	return goExpr{code: fmt.Sprintf("%s.Println(%s)", runtimeAlias, strings.Join(parts, ", ")), kind: exprKindValue}, nil
}

func testingExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) < 2 {
		return goExpr{}, fmt.Errorf("testing expects a label and body")
	}
	label, ok := args[0].(StringExpr)
	if !ok {
		return goExpr{}, fmt.Errorf("testing expects a string label")
	}
	if label.Value == "" {
		return goExpr{}, fmt.Errorf("testing label cannot be empty")
	}
	_ = label

	body, err := doExprToGo(args[1:], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}

	var out strings.Builder
	fmt.Fprintf(&out, "func() %s.Value {\n", runtimeAlias)
	fmt.Fprintf(&out, "\t_ = %s\n", body.code)
	fmt.Fprintf(&out, "\treturn %s.NilValue()\n", runtimeAlias)
	out.WriteString("}()")

	return goExpr{code: out.String(), kind: exprKindValue}, nil
}

func testingBodyExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) == 0 {
		return goExpr{code: fmt.Sprintf("%s.NilValue()", runtimeAlias), kind: exprKindValue}, nil
	}
	body, err := doExprToGo(args, ctx, locals)
	if err != nil {
		return goExpr{}, err
	}

	var out strings.Builder
	fmt.Fprintf(&out, "func() %s.Value {\n", runtimeAlias)
	fmt.Fprintf(&out, "\t_ = %s\n", body.code)
	fmt.Fprintf(&out, "\treturn %s.NilValue()\n", runtimeAlias)
	out.WriteString("}()")
	return goExpr{code: out.String(), kind: exprKindValue}, nil
}

func isExprToGo(form ListExpr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	args := form.Elements[1:]
	if len(args) != 1 && len(args) != 2 {
		return goExpr{}, fmt.Errorf("is expects an expression and optional message")
	}

	testExpr, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	condition, err := truthyExprToGo(testExpr)
	if err != nil {
		return goExpr{}, err
	}

	message := exprToSourceString(args[0])
	if len(args) == 2 {
		msg, ok := args[1].(StringExpr)
		if !ok {
			return goExpr{}, fmt.Errorf("is message must be a string")
		}
		message = msg.Value
	}
	if line, col, ok := exprPos(form); ok {
		message = fmt.Sprintf("at %d:%d: %s", line, col, message)
	}

	var out strings.Builder
	fmt.Fprintf(&out, "func() %s.Value {\n", runtimeAlias)
	fmt.Fprintf(&out, "\tif !(%s) {\n", condition)
	fmt.Fprintf(&out, "\t\tpanic(%q)\n", message)
	out.WriteString("\t}\n")
	fmt.Fprintf(&out, "\treturn %s.NilValue()\n", runtimeAlias)
	out.WriteString("}()")

	return goExpr{code: out.String(), kind: exprKindValue}, nil
}

func strArgParts(args []Expr, ctx compileContext, locals map[string]exprKind) ([]string, error) {
	parts := make([]string, 0, len(args))
	for _, item := range args {
		part, err := exprToGo(item, ctx, locals)
		if err != nil {
			return nil, err
		}
		parts = append(parts, part.code)
	}
	return parts, nil
}

func truthyExprToGo(expr goExpr) (string, error) {
	switch expr.kind {
	case exprKindBool:
		return expr.code, nil
	case exprKindValue:
		return fmt.Sprintf("%s.IsTruthy(%s)", runtimeAlias, expr.code), nil
	case exprKindString:
		return "true", nil
	default:
		return "", fmt.Errorf("unsupported if test expression")
	}
}

func goTypeForExprKind(kind exprKind) (string, error) {
	switch kind {
	case exprKindValue:
		return runtimeAlias + ".Value", nil
	case exprKindBool:
		return "bool", nil
	case exprKindString:
		return "string", nil
	default:
		return "", fmt.Errorf("unsupported expression kind")
	}
}

func toGoIdentifier(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("empty symbol")
	}

	var out strings.Builder
	runes := []rune(name)
	for i, ch := range runes {
		if i == 0 {
			if ch != '_' && !unicode.IsLetter(ch) {
				return "", fmt.Errorf("unsupported symbol %q", name)
			}
			out.WriteRune(ch)
			continue
		}
		if ch == '-' {
			out.WriteByte('_')
			continue
		}
		if ch != '_' && !unicode.IsLetter(ch) && !unicode.IsDigit(ch) {
			return "", fmt.Errorf("unsupported symbol %q", name)
		}
		out.WriteRune(ch)
	}
	return out.String(), nil
}

func renderFunctionDef(fn functionDef) string {
	if fn.hasRest {
		return fmt.Sprintf("func %s(args ...flagrt.Value) flagrt.Value {\n%s}\n",
			fn.variadicName,
			renderStandaloneVariadicBody(fn),
		)
	}
	return fmt.Sprintf("func %s(%s) flagrt.Value {\n%s\treturn %s\n}\n\nfunc %s(args ...flagrt.Value) flagrt.Value {\n%s}\n",
		fn.arityName,
		renderDirectParamList(fn),
		renderLocalInits(fn.localInits),
		fn.body,
		fn.variadicName,
		renderVariadicFunctionBody(fn),
	)
}

func renderFunctionLiteral(fn functionDef) string {
	return fmt.Sprintf("func(args ...flagrt.Value) flagrt.Value {\n%s}", renderStandaloneVariadicBody(fn))
}

func renderDirectFunctionType(fn functionDef) string {
	return fmt.Sprintf("func(%s) flagrt.Value", renderDirectParamTypes(fn))
}

func renderDirectFunctionLiteral(fn functionDef) string {
	return fmt.Sprintf("func(%s) flagrt.Value {\n%s\treturn %s\n}", renderDirectParamList(fn), renderLocalInits(fn.localInits), fn.body)
}

func renderVariadicFunctionLiteral(fn functionDef) string {
	return fmt.Sprintf("func(args ...flagrt.Value) flagrt.Value {\n%s}", renderVariadicFunctionBody(fn))
}

func renderVariadicFunctionBody(fn functionDef) string {
	var body strings.Builder
	fmt.Fprintf(&body, "\tif len(args) != %d {\n", len(fn.params))
	fmt.Fprintf(&body, "\t\tpanic(%q)\n", fmt.Sprintf("%s expects exactly %d arguments", fn.goName, len(fn.params)))
	body.WriteString("\t}\n")
	if len(fn.params) == 0 {
		fmt.Fprintf(&body, "\treturn %s()\n", fn.arityName)
		return body.String()
	}
	callArgs := make([]string, 0, len(fn.params))
	for index := range fn.params {
		callArgs = append(callArgs, fmt.Sprintf("args[%d]", index))
	}
	fmt.Fprintf(&body, "\treturn %s(%s)\n", fn.arityName, strings.Join(callArgs, ", "))
	return body.String()
}

func renderStandaloneVariadicBody(fn functionDef) string {
	var body strings.Builder
	if fn.hasRest {
		fmt.Fprintf(&body, "\tif len(args) < %d {\n", len(fn.params))
		fmt.Fprintf(&body, "\t\tpanic(%q)\n", fmt.Sprintf("%s expects at least %d arguments", fn.goName, len(fn.params)))
		body.WriteString("\t}\n")
	} else {
		fmt.Fprintf(&body, "\tif len(args) != %d {\n", len(fn.params))
		fmt.Fprintf(&body, "\t\tpanic(%q)\n", fmt.Sprintf("%s expects exactly %d arguments", fn.goName, len(fn.params)))
		body.WriteString("\t}\n")
	}
	for index, param := range fn.params {
		fmt.Fprintf(&body, "\t%s := args[%d]\n", param, index)
	}
	for _, init := range fn.localInits {
		body.WriteString(init)
	}
	fmt.Fprintf(&body, "\treturn %s\n", fn.body)
	return body.String()
}

func renderLocalInits(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	var out strings.Builder
	for _, line := range lines {
		out.WriteString(line)
	}
	return out.String()
}

func renderDirectParamList(fn functionDef) string {
	if len(fn.params) == 0 {
		return ""
	}
	parts := make([]string, 0, len(fn.params))
	for _, param := range fn.params {
		parts = append(parts, fmt.Sprintf("%s flagrt.Value", param))
	}
	return strings.Join(parts, ", ")
}

func renderDirectParamTypes(fn functionDef) string {
	if len(fn.params) == 0 {
		return ""
	}
	parts := make([]string, 0, len(fn.params))
	for range fn.params {
		parts = append(parts, "flagrt.Value")
	}
	return strings.Join(parts, ", ")
}
