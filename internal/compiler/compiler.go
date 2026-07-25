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

type mainStmt struct {
	code string
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
	macros    map[string]macroDef
}

type macroDef struct {
	params    []string
	restParam string
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
	namespace, functions, vars, stmts, err := compileForms(source)
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
	for _, stmt := range stmts {
		fmt.Fprintf(&out, "\t%s\n", stmt.code)
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

func compileForms(source string) (string, []functionDef, []varDef, []mainStmt, error) {
	ast, err := ParseFile(source)
	if err != nil {
		return "", nil, nil, nil, err
	}

	ctx, err := newCompileContext()
	if err != nil {
		return "", nil, nil, nil, err
	}
	namespace := ""
	functions := make([]functionDef, 0, len(ast.Forms))
	vars := make([]varDef, 0, len(ast.Forms))
	stmts := make([]mainStmt, 0, len(ast.Forms))

	for _, form := range ast.Forms {
		if list, ok := form.(ListExpr); ok && len(list.Elements) > 0 {
			if head, ok := list.Elements[0].(SymbolExpr); ok && head.Name == "defmacro" {
				name, def, err := compileDefmacro(list)
				if err != nil {
					return "", nil, nil, nil, err
				}
				ctx.macros[name] = def
				continue
			}
		}

		expanded, err := macroExpand(form, ctx, 0)
		if err != nil {
			return "", nil, nil, nil, err
		}

		form = expanded
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
			binding, kind, err := compileDef(list, ctx)
			if err != nil {
				return "", nil, nil, nil, err
			}
			ctx.globals[binding.goName] = kind
			vars = append(vars, binding)
		case "defmacro":
			name, def, err := compileDefmacro(list)
			if err != nil {
				return "", nil, nil, nil, err
			}
			ctx.macros[name] = def
		case "println":
			arg, err := strArgExprForGoCall(list.Elements[1:], ctx, nil)
			if err != nil {
				return "", nil, nil, nil, err
			}
			stmts = append(stmts, mainStmt{code: fmt.Sprintf("fmt.Println(%s)", arg)})
		case "print":
			arg, err := argumentExprForGoCall(list.Elements[1:], ctx, nil)
			if err != nil {
				return "", nil, nil, nil, err
			}
			stmts = append(stmts, mainStmt{code: fmt.Sprintf("fmt.Print(%s)", arg)})
		default:
			expr, err := exprToGo(form, ctx, nil)
			if err != nil {
				return "", nil, nil, nil, err
			}
			stmts = append(stmts, mainStmt{code: fmt.Sprintf("_ = %s", expr.code)})
		}
	}

	return namespace, functions, vars, stmts, nil
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
	localSymbols := make(map[string]exprKind, len(paramsExpr.Elements))
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
		localSymbols[goParam] = exprKindValue
		params = append(params, goParam)
	}

	fnCtx := compileContext{
		functions: make(map[string]functionDef, len(ctx.functions)+1),
		globals:   ctx.globals,
	}
	for name, def := range ctx.functions {
		fnCtx.functions[name] = def
	}
	fnCtx.functions[goName] = functionDef{goName: goName, params: params}

	body, err := exprToGo(form.Elements[3], fnCtx, localSymbols)
	if err != nil {
		return functionDef{}, err
	}
	if body.kind != exprKindValue {
		return functionDef{}, fmt.Errorf("defn body must evaluate to Value")
	}

	return functionDef{goName: goName, params: params, body: body.code}, nil
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
		return varDef{}, 0, false, fmt.Errorf("def value must evaluate to Value")
	}

	_, exists := ctx.globals[goName]
	return varDef{goName: goName, expr: valueExpr.code}, valueExpr.kind, !exists, nil
}

func compileDefmacro(form ListExpr) (string, macroDef, error) {
	if len(form.Elements) != 4 {
		return "", macroDef{}, fmt.Errorf("defmacro expects name, vector params, and body")
	}
	nameExpr, ok := form.Elements[1].(SymbolExpr)
	if !ok || nameExpr.Name == "" {
		return "", macroDef{}, fmt.Errorf("defmacro expects a macro name")
	}
	paramsExpr, ok := form.Elements[2].(VectorExpr)
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
		body:      form.Elements[3],
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
		return ListExpr{Elements: out}, nil
	case VectorExpr:
		out := make([]Expr, 0, len(value.Elements))
		for _, item := range value.Elements {
			expanded, err := macroExpand(item, ctx, depth)
			if err != nil {
				return nil, err
			}
			out = append(out, expanded)
		}
		return VectorExpr{Elements: out}, nil
	case MapExpr:
		out := make([]Expr, 0, len(value.Entries))
		for _, item := range value.Entries {
			expanded, err := macroExpand(item, ctx, depth)
			if err != nil {
				return nil, err
			}
			out = append(out, expanded)
		}
		return MapExpr{Entries: out}, nil
	case SetExpr:
		out := make([]Expr, 0, len(value.Elements))
		for _, item := range value.Elements {
			expanded, err := macroExpand(item, ctx, depth)
			if err != nil {
				return nil, err
			}
			out = append(out, expanded)
		}
		return SetExpr{Elements: out}, nil
	case HashFnExpr:
		expanded, err := macroExpand(value.Body, ctx, depth)
		if err != nil {
			return nil, err
		}
		return HashFnExpr{Body: expanded}, nil
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
		return ListExpr{Elements: out}
	case VectorExpr:
		out := make([]Expr, 0, len(value.Elements))
		for _, item := range value.Elements {
			out = append(out, copyExpr(item))
		}
		return VectorExpr{Elements: out}
	case MapExpr:
		out := make([]Expr, 0, len(value.Entries))
		for _, item := range value.Entries {
			out = append(out, copyExpr(item))
		}
		return MapExpr{Entries: out}
	case SetExpr:
		out := make([]Expr, 0, len(value.Elements))
		for _, item := range value.Elements {
			out = append(out, copyExpr(item))
		}
		return SetExpr{Elements: out}
	case HashFnExpr:
		return HashFnExpr{Body: copyExpr(value.Body)}
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
	case FloatExpr:
		if arg.Raw != "" {
			return goExpr{code: fmt.Sprintf("%s.NewDouble(%s)", runtimeAlias, arg.Raw), kind: exprKindValue}, nil
		}
		return goExpr{code: fmt.Sprintf("%s.NewDouble(%g)", runtimeAlias, arg.Value), kind: exprKindValue}, nil
	case KeywordExpr:
		return goExpr{code: fmt.Sprintf("%s.NewKeyword(%q)", runtimeAlias, arg.Name), kind: exprKindValue}, nil
	case QuotedSymbolExpr:
		return goExpr{code: fmt.Sprintf("%s.NewSymbol(%q)", runtimeAlias, arg.Name), kind: exprKindValue}, nil
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
		if err != nil {
			return goExpr{}, err
		}
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
		return goExpr{}, fmt.Errorf("unknown symbol %q", arg.Name)
	case ListExpr:
		return listExprToGo(arg, ctx, locals)
	default:
		return goExpr{}, fmt.Errorf("unsupported literal")
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
		case "map":
			return mapCallExprToGo(list.Elements[1:], ctx, locals)
		case "filter":
			return filterCallExprToGo(list.Elements[1:], ctx, locals)
		case "assoc":
			return assocExprToGo(list.Elements[1:], ctx, locals)
		case "dissoc":
			return dissocExprToGo(list.Elements[1:], ctx, locals)
		case "fn":
			return fnExprToGo(list.Elements[1:], ctx, locals)
		}
	}

	return callExprToGo(list.Elements[0], list.Elements[1:], ctx, locals)
}

func callExprToGo(calleeExpr Expr, argsExpr []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
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
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("function argument must evaluate to Value")
		}
		args = append(args, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.Call(%s)", runtimeAlias, strings.Join(append([]string{callee.code}, args...), ", ")), kind: exprKindValue}, nil
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

	bindings := make([]string, 0, len(bindingsExpr.Elements)/2)
	for i := 0; i < len(bindingsExpr.Elements); i += 2 {
		nameExpr, ok := bindingsExpr.Elements[i].(SymbolExpr)
		if !ok || nameExpr.Name == "" {
			return goExpr{}, fmt.Errorf("let binding name must be a symbol")
		}

		goName, err := toGoIdentifier(nameExpr.Name)
		if err != nil {
			return goExpr{}, err
		}

		valueExpr, err := exprToGo(bindingsExpr.Elements[i+1], ctx, localKinds)
		if err != nil {
			return goExpr{}, err
		}

		bindings = append(bindings, fmt.Sprintf("\t%s := %s\n", goName, valueExpr.code))
		localKinds[goName] = valueExpr.kind
	}

	bodyExprs := args[1:]
	if len(bodyExprs) == 0 {
		if len(bindingsExpr.Elements) == 0 {
			return goExpr{}, fmt.Errorf("let without body requires at least one binding")
		}
		lastBindingName := bindingsExpr.Elements[len(bindingsExpr.Elements)-2].(SymbolExpr)
		bodyExprs = []Expr{lastBindingName}
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
		return ListExpr{Elements: out}
	case VectorExpr:
		out := make([]Expr, 0, len(value.Elements))
		for _, item := range value.Elements {
			out = append(out, replaceHashFnPlaceholders(item))
		}
		return VectorExpr{Elements: out}
	case MapExpr:
		out := make([]Expr, 0, len(value.Entries))
		for _, item := range value.Entries {
			out = append(out, replaceHashFnPlaceholders(item))
		}
		return MapExpr{Entries: out}
	case SetExpr:
		out := make([]Expr, 0, len(value.Elements))
		for _, item := range value.Elements {
			out = append(out, replaceHashFnPlaceholders(item))
		}
		return SetExpr{Elements: out}
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
	params := make([]string, 0, len(paramsExpr.Elements))
	localKinds := make(map[string]exprKind, len(locals)+len(paramsExpr.Elements))
	for name, kind := range locals {
		localKinds[name] = kind
	}
	paramNames := make(map[string]struct{}, len(paramsExpr.Elements))

	for _, paramExpr := range paramsExpr.Elements {
		param, ok := paramExpr.(SymbolExpr)
		if !ok || param.Name == "" {
			return goExpr{}, fmt.Errorf("%s parameters must be symbols", label)
		}

		goParam, err := toGoIdentifier(param.Name)
		if err != nil {
			return goExpr{}, err
		}
		if _, exists := paramNames[goParam]; exists {
			return goExpr{}, fmt.Errorf("duplicate parameter %q", param.Name)
		}
		paramNames[goParam] = struct{}{}
		localKinds[goParam] = exprKindValue
		params = append(params, goParam)
	}

	body, err := exprToGo(bodyExpr, ctx, localKinds)
	if err != nil {
		return goExpr{}, err
	}
	if body.kind != exprKindValue {
		return goExpr{}, fmt.Errorf("%s body must evaluate to Value", label)
	}

	def := functionDef{
		goName: label,
		params: params,
		body:   body.code,
	}
	return goExpr{code: fmt.Sprintf("%s.NewFunction(%s)", runtimeAlias, renderFunctionLiteral(def)), kind: exprKindValue}, nil
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
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("map literal entries must evaluate to Value")
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
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("vector literal entries must evaluate to Value")
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
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("set literal entries must evaluate to Value")
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
