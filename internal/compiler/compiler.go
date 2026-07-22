package compiler

import (
	"bytes"
	"fmt"
	"go/format"
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

type printCall struct {
	function string
	arg      string
}

type functionDef struct {
	goName string
	params []string
	body   string
}

type varDef struct {
	goName string
	expr   string
}

type compileContext struct {
	functions map[string]functionDef
	globals   map[string]exprKind
}

// Compile translates a small FLAG source file into a Go program.
func Compile(source string) ([]byte, error) {
	namespace, functions, vars, calls, err := compileForms(source)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	out.WriteString("package main\n\n")
	out.WriteString("import (\n")
	out.WriteString("\t\"fmt\"\n")
	out.WriteString("\tflagrt \"flag-lang/runtime\"\n")
	out.WriteString(")\n\n")

	if namespace != "" {
		fmt.Fprintf(&out, "// Source namespace: %s\n", namespace)
	}

	for _, fn := range functions {
		out.WriteString(renderFunctionDef(fn))
		out.WriteString("\n")
	}

	for _, v := range vars {
		fmt.Fprintf(&out, "var %s = %s\n", v.goName, v.expr)
	}
	if len(vars) > 0 {
		out.WriteString("\n")
	}

	out.WriteString("func main() {\n")
	for _, call := range calls {
		fmt.Fprintf(&out, "\tfmt.%s(%s)\n", call.function, call.arg)
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

	ctx := compileContext{
		functions: make(map[string]functionDef),
		globals:   make(map[string]exprKind),
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
	return &ReplCompiler{
		ctx: compileContext{
			functions: make(map[string]functionDef),
			globals:   make(map[string]exprKind),
		},
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
					setup = fmt.Sprintf("var %s flagrt.Value = %s", binding.goName, binding.expr)
				}
				return ReplCompiled{
					Setup:      setup,
					ResultExpr: fmt.Sprintf("%s.ValueToAny(%s)", runtimeAlias, binding.goName),
				}, nil
			case "defn":
				def, err := compileDefn(list, r.ctx)
				if err != nil {
					return ReplCompiled{}, err
				}
				_, exists := r.ctx.functions[def.goName]
				r.ctx.functions[def.goName] = def

				setup := fmt.Sprintf("%s = %s", def.goName, renderFunctionLiteral(def))
				if !exists {
					setup = fmt.Sprintf("var %s = %s", def.goName, renderFunctionLiteral(def))
				}
				return ReplCompiled{
					Setup:      setup,
					ResultExpr: fmt.Sprintf("%q", def.goName),
				}, nil
			}
		}
	}

	expr, err := exprToGo(ast.Forms[0], r.ctx, nil)
	if err != nil {
		return ReplCompiled{}, err
	}
	if expr.kind == exprKindValue {
		return ReplCompiled{ResultExpr: fmt.Sprintf("%s.ValueToAny(%s)", runtimeAlias, expr.code)}, nil
	}
	return ReplCompiled{ResultExpr: expr.code}, nil
}

func compileForms(source string) (string, []functionDef, []varDef, []printCall, error) {
	ast, err := ParseFile(source)
	if err != nil {
		return "", nil, nil, nil, err
	}

	ctx := compileContext{
		functions: make(map[string]functionDef),
		globals:   make(map[string]exprKind),
	}
	namespace := ""
	functions := make([]functionDef, 0, len(ast.Forms))
	vars := make([]varDef, 0, len(ast.Forms))
	calls := make([]printCall, 0, len(ast.Forms))

	for _, form := range ast.Forms {
		list, ok := form.(ListExpr)
		if !ok || len(list.Elements) == 0 {
			return "", nil, nil, nil, fmt.Errorf("unsupported form")
		}

		head, ok := list.Elements[0].(SymbolExpr)
		if !ok {
			return "", nil, nil, nil, fmt.Errorf("unsupported form")
		}

		switch head.Name {
		case "ns":
			if namespace != "" {
				return "", nil, nil, nil, fmt.Errorf("namespace already declared")
			}
			if len(list.Elements) != 2 {
				return "", nil, nil, nil, fmt.Errorf("ns expects one namespace symbol")
			}
			name, ok := list.Elements[1].(SymbolExpr)
			if !ok || name.Name == "" {
				return "", nil, nil, nil, fmt.Errorf("namespace cannot be empty")
			}
			namespace = name.Name
		case "defn":
			def, err := compileDefn(list, ctx)
			if err != nil {
				return "", nil, nil, nil, err
			}
			if _, exists := ctx.functions[def.goName]; exists {
				return "", nil, nil, nil, fmt.Errorf("function %q already defined", def.goName)
			}
			ctx.functions[def.goName] = def
			functions = append(functions, def)
		case "def":
			binding, err := compileDef(list, ctx)
			if err != nil {
				return "", nil, nil, nil, err
			}
			ctx.globals[binding.goName] = exprKindValue
			vars = append(vars, binding)
		case "println", "print":
			arg, err := argumentExprForGoCall(list.Elements[1:], ctx, nil)
			if err != nil {
				return "", nil, nil, nil, err
			}
			callName := "Print"
			if head.Name == "println" {
				callName = "Println"
			}
			calls = append(calls, printCall{function: callName, arg: arg})
		default:
			return "", nil, nil, nil, fmt.Errorf("unsupported form")
		}
	}

	return namespace, functions, vars, calls, nil
}

func compileDefn(form ListExpr, ctx compileContext) (functionDef, error) {
	if len(form.Elements) != 4 {
		return functionDef{}, fmt.Errorf("defn expects name, vector params, and body")
	}

	nameExpr, ok := form.Elements[1].(SymbolExpr)
	if !ok || nameExpr.Name == "" {
		return functionDef{}, fmt.Errorf("defn expects a function name")
	}

	goName, err := toGoIdentifier(nameExpr.Name)
	if err != nil {
		return functionDef{}, err
	}

	paramsExpr, ok := form.Elements[2].(VectorExpr)
	if !ok {
		return functionDef{}, fmt.Errorf("defn expects a parameter vector")
	}

	params := make([]string, 0, len(paramsExpr.Elements))
	localSymbols := make(map[string]struct{}, len(paramsExpr.Elements))
	for _, paramExpr := range paramsExpr.Elements {
		param, ok := paramExpr.(SymbolExpr)
		if !ok || param.Name == "" {
			return functionDef{}, fmt.Errorf("defn parameters must be symbols")
		}

		goParam, err := toGoIdentifier(param.Name)
		if err != nil {
			return functionDef{}, err
		}
		if _, exists := localSymbols[goParam]; exists {
			return functionDef{}, fmt.Errorf("duplicate parameter %q", param.Name)
		}
		localSymbols[goParam] = struct{}{}
		params = append(params, goParam)
	}

	body, err := exprToGo(form.Elements[3], ctx, localSymbols)
	if err != nil {
		return functionDef{}, err
	}
	if body.kind != exprKindValue {
		return functionDef{}, fmt.Errorf("defn body must evaluate to numeric Value")
	}

	return functionDef{goName: goName, params: params, body: body.code}, nil
}

func compileDef(form ListExpr, ctx compileContext) (varDef, error) {
	binding, kind, _, err := compileDefForRepl(form, ctx)
	if err != nil {
		return varDef{}, err
	}
	if kind != exprKindValue {
		return varDef{}, fmt.Errorf("def value must evaluate to numeric Value")
	}
	return binding, nil
}

func compileDefForRepl(form ListExpr, ctx compileContext) (varDef, exprKind, bool, error) {
	if len(form.Elements) != 3 {
		return varDef{}, 0, false, fmt.Errorf("def expects name and value")
	}

	nameExpr, ok := form.Elements[1].(SymbolExpr)
	if !ok || nameExpr.Name == "" {
		return varDef{}, 0, false, fmt.Errorf("def expects a symbol name")
	}
	goName, err := toGoIdentifier(nameExpr.Name)
	if err != nil {
		return varDef{}, 0, false, err
	}

	valueExpr, err := exprToGo(form.Elements[2], ctx, nil)
	if err != nil {
		return varDef{}, 0, false, err
	}
	if valueExpr.kind != exprKindValue {
		return varDef{}, 0, false, fmt.Errorf("def value must evaluate to numeric Value")
	}

	_, exists := ctx.globals[goName]
	return varDef{goName: goName, expr: valueExpr.code}, valueExpr.kind, !exists, nil
}

func argumentExprForGoCall(args []Expr, ctx compileContext, locals map[string]struct{}) (string, error) {
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

func exprToGo(expr Expr, ctx compileContext, locals map[string]struct{}) (goExpr, error) {
	switch arg := expr.(type) {
	case StringExpr:
		return goExpr{code: fmt.Sprintf("%q", arg.Value), kind: exprKindString}, nil
	case IntExpr:
		return goExpr{code: fmt.Sprintf("%s.NewLong(%d)", runtimeAlias, arg.Value), kind: exprKindValue}, nil
	case FloatExpr:
		if arg.Raw != "" {
			return goExpr{code: fmt.Sprintf("%s.NewDouble(%s)", runtimeAlias, arg.Raw), kind: exprKindValue}, nil
		}
		return goExpr{code: fmt.Sprintf("%s.NewDouble(%g)", runtimeAlias, arg.Value), kind: exprKindValue}, nil
	case SymbolExpr:
		ident, err := toGoIdentifier(arg.Name)
		if err != nil {
			return goExpr{}, err
		}
		if locals != nil {
			if _, ok := locals[ident]; ok {
				return goExpr{code: ident, kind: exprKindValue}, nil
			}
		}
		if kind, ok := ctx.globals[ident]; ok {
			return goExpr{code: ident, kind: kind}, nil
		}
		return goExpr{}, fmt.Errorf("unknown symbol %q", arg.Name)
	case ListExpr:
		return listExprToGo(arg, ctx, locals)
	default:
		return goExpr{}, fmt.Errorf("unsupported literal")
	}
}

func listExprToGo(list ListExpr, ctx compileContext, locals map[string]struct{}) (goExpr, error) {
	if len(list.Elements) == 0 {
		return goExpr{}, fmt.Errorf("unsupported form")
	}

	head, ok := list.Elements[0].(SymbolExpr)
	if !ok {
		return goExpr{}, fmt.Errorf("unsupported form")
	}

	switch head.Name {
	case "+":
		return infixExprToGo(list.Elements[1:], runtimeAlias+".Add", ctx, locals)
	case "*":
		return infixExprToGo(list.Elements[1:], runtimeAlias+".Mul", ctx, locals)
	case "-":
		return infixExprToGo(list.Elements[1:], runtimeAlias+".Sub", ctx, locals)
	case "/":
		return infixExprToGo(list.Elements[1:], runtimeAlias+".Div", ctx, locals)
	case "=":
		return equalityExprToGo(list.Elements[1:], ctx, locals)
	default:
		goName, err := toGoIdentifier(head.Name)
		if err != nil {
			return goExpr{}, err
		}
		if _, ok := ctx.functions[goName]; !ok {
			return goExpr{}, fmt.Errorf("unknown function %q", head.Name)
		}

		args := make([]string, 0, len(list.Elements)-1)
		for _, item := range list.Elements[1:] {
			part, err := exprToGo(item, ctx, locals)
			if err != nil {
				return goExpr{}, err
			}
			if part.kind != exprKindValue {
				return goExpr{}, fmt.Errorf("function argument must evaluate to numeric Value")
			}
			args = append(args, part.code)
		}
		return goExpr{code: fmt.Sprintf("%s(%s)", goName, strings.Join(args, ", ")), kind: exprKindValue}, nil
	}
}

func infixExprToGo(args []Expr, runtimeOp string, ctx compileContext, locals map[string]struct{}) (goExpr, error) {
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

func equalityExprToGo(args []Expr, ctx compileContext, locals map[string]struct{}) (goExpr, error) {
	if len(args) < 2 {
		return goExpr{}, fmt.Errorf("= expects at least two arguments")
	}

	parts := make([]string, 0, len(args))
	for _, item := range args {
		part, err := exprToGo(item, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("= expects numeric Value arguments")
		}
		parts = append(parts, part.code)
	}

	checks := make([]string, 0, len(parts)-1)
	for i := 0; i < len(parts)-1; i++ {
		checks = append(checks, fmt.Sprintf("%s.Eq(%s, %s)", runtimeAlias, parts[i], parts[i+1]))
	}
	return goExpr{code: strings.Join(checks, " && "), kind: exprKindBool}, nil
}

func toGoIdentifier(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("empty symbol")
	}

	runes := []rune(name)
	for i, ch := range runes {
		if i == 0 {
			if ch != '_' && !unicode.IsLetter(ch) {
				return "", fmt.Errorf("unsupported symbol %q", name)
			}
			continue
		}
		if ch != '_' && !unicode.IsLetter(ch) && !unicode.IsDigit(ch) {
			return "", fmt.Errorf("unsupported symbol %q", name)
		}
	}
	return name, nil
}

func renderFunctionDef(fn functionDef) string {
	return fmt.Sprintf("func %s(args ...flagrt.Value) flagrt.Value {\n%s}\n", fn.goName, renderFunctionBody(fn))
}

func renderFunctionLiteral(fn functionDef) string {
	return fmt.Sprintf("func(args ...flagrt.Value) flagrt.Value {\n%s}", renderFunctionBody(fn))
}

func renderFunctionBody(fn functionDef) string {
	var body strings.Builder
	if len(fn.params) > 0 {
		fmt.Fprintf(&body, "\tif len(args) != %d {\n", len(fn.params))
		fmt.Fprintf(&body, "\t\tpanic(%q)\n", fmt.Sprintf("%s expects exactly %d arguments", fn.goName, len(fn.params)))
		body.WriteString("\t}\n")
		for index, param := range fn.params {
			fmt.Fprintf(&body, "\t%s := args[%d]\n", param, index)
		}
	}
	fmt.Fprintf(&body, "\treturn %s\n", fn.body)
	return body.String()
}
