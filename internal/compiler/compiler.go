package compiler

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"go/token"
	"path/filepath"
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
	exprKindMutableValue
)

type goExpr struct {
	code string
	kind exprKind
}

func exprPos(expr Expr) (int, int, bool) {
	switch value := expr.(type) {
	case ListExpr:
		return value.Line, value.Col, value.Line > 0
	case VectorExpr:
		return value.Line, value.Col, value.Line > 0
	case MapExpr:
		return value.Line, value.Col, value.Line > 0
	case SetExpr:
		return value.Line, value.Col, value.Line > 0
	case HashFnExpr:
		return value.Line, value.Col, value.Line > 0
	case SymbolExpr:
		return value.Line, value.Col, value.Line > 0
	case KeywordExpr:
		return value.Line, value.Col, value.Line > 0
	case QuotedSymbolExpr:
		return value.Line, value.Col, value.Line > 0
	case QuotedListExpr:
		return value.Line, value.Col, value.Line > 0
	case StringExpr:
		return value.Line, value.Col, value.Line > 0
	case CharExpr:
		return value.Line, value.Col, value.Line > 0
	case IntExpr:
		return value.Line, value.Col, value.Line > 0
	case BigIntExpr:
		return value.Line, value.Col, value.Line > 0
	case FloatExpr:
		return value.Line, value.Col, value.Line > 0
	case RatioExpr:
		return value.Line, value.Col, value.Line > 0
	case MetaExpr:
		return value.Line, value.Col, value.Line > 0
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

func annotateExprError(expr Expr, err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if strings.Contains(msg, " at ") {
		return err
	}
	return exprError(expr, msg)
}

func exprToSourceString(expr Expr) string {
	switch value := expr.(type) {
	case StringExpr:
		return strconv.Quote(value.Value)
	case CharExpr:
		switch value.Value {
		case ' ':
			return `\space`
		case '\n':
			return `\newline`
		case '\t':
			return `\tab`
		default:
			return "\\" + string(value.Value)
		}
	case IntExpr:
		return fmt.Sprintf("%d", value.Value)
	case BigIntExpr:
		return value.Value + "N"
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
	case MetaExpr:
		return "^" + exprToSourceString(value.Meta) + " " + exprToSourceString(value.Target)
	default:
		return fmt.Sprintf("%v", expr)
	}
}

type mainStmt struct {
	code string
	kind exprKind // 0 means a statement (println/print); otherwise an expression
}

type testCase struct {
	goName     string
	label      string
	line       int
	bodySource string
}

type functionDef struct {
	flagName     string // original FLAG name (e.g. "move", "flag_main")
	goName       string
	variadicName string
	arityName    string
	hasRest      bool
	doc          string
	params       []string
	localInits   []string
	body         string
	goSignature  string // custom Go signature for interop wrappers (empty = use standard)
}

type varDef struct {
	flagName string
	goName   string
	doc      string
	expr     string
}

type compileContext struct {
	functions         map[string]functionDef
	globals           map[string]exprKind
	macros            map[string]macroDef
	prologueFns       []functionDef
	prologueVars      []varDef
	selfFunctionName  string // FLAG source name for self-recursion matching
	selfFunctionArity int
	selfArityName     string
	// loopBindingNames tracks active loop/recur bindings for the current
	// expression context (nil when not inside a loop form).
	loopBindingNames []string
	// namespace is the current module :namespace for Go name mangling.
	// Empty means legacy single-file mode (no module prefix on idents).
	namespace string
	// moduleSymbols maps FLAG names as written in this module (bare local
	// names, :refer names, and qualified import names like "chess/move") to
	// the Go identifier of the binding.
	moduleSymbols map[string]string
	// constants, when non-nil, hoists each distinct constant literal (keywords,
	// symbols, strings, ratios, bigints, and collections built entirely from
	// constants) to a single package-level var so it is constructed once at
	// program start instead of on every evaluation. It is a pointer so nested
	// scopes (defn bodies, etc.) share one interner, and identical constants
	// across all merged flag files collapse to the same var. Left nil for the
	// REPL / single-expression paths, which emit constants inline.
	constants *constantInterner
	// recordTypes maps generated Go struct names to FLAG record names.
	recordTypes map[string]string
	// allowRedefine lets the REPL overwrite functions already in ctx.
	allowRedefine bool
}

// bindModuleName records a FLAG name -> Go ident mapping for this module.
func (ctx *compileContext) bindModuleName(flagName, goName string) error {
	if ctx.moduleSymbols == nil {
		return nil
	}
	if prev, ok := ctx.moduleSymbols[flagName]; ok && prev != goName {
		return fmt.Errorf("symbol %q already bound to %s", flagName, prev)
	}
	ctx.moduleSymbols[flagName] = goName
	if ctx.namespace != "" && !strings.Contains(flagName, "/") {
		qualified := ctx.namespace + "/" + flagName
		if prev, ok := ctx.moduleSymbols[qualified]; ok && prev != goName {
			return fmt.Errorf("symbol %q already bound to %s", qualified, prev)
		}
		ctx.moduleSymbols[qualified] = goName
	}
	return nil
}

// resolveModuleSymbol returns the Go identifier for a FLAG symbol name when
// it is registered in the module symbol table.
func (ctx *compileContext) resolveModuleSymbol(flagName string) (string, bool) {
	if ctx.moduleSymbols == nil {
		return "", false
	}
	goName, ok := ctx.moduleSymbols[flagName]
	return goName, ok
}

func copyCompileContext(ctx compileContext) compileContext {
	out := compileContext{
		functions:         make(map[string]functionDef, len(ctx.functions)),
		globals:           make(map[string]exprKind, len(ctx.globals)),
		macros:            make(map[string]macroDef, len(ctx.macros)),
		prologueFns:       ctx.prologueFns,
		prologueVars:      ctx.prologueVars,
		constants:         ctx.constants,
		namespace:         ctx.namespace,
		selfFunctionName:  ctx.selfFunctionName,
		selfFunctionArity: ctx.selfFunctionArity,
		selfArityName:     ctx.selfArityName,
		allowRedefine:     ctx.allowRedefine,
	}
	if len(ctx.loopBindingNames) > 0 {
		out.loopBindingNames = append([]string(nil), ctx.loopBindingNames...)
	}
	for k, v := range ctx.functions {
		out.functions[k] = v
	}
	for k, v := range ctx.globals {
		out.globals[k] = v
	}
	for k, v := range ctx.macros {
		out.macros[k] = v
	}
	if ctx.moduleSymbols != nil {
		out.moduleSymbols = make(map[string]string, len(ctx.moduleSymbols))
		for k, v := range ctx.moduleSymbols {
			out.moduleSymbols[k] = v
		}
	}
	if ctx.recordTypes != nil {
		out.recordTypes = ctx.recordTypes
	}
	return out
}

// constantInterner collects the distinct constant constructor expressions used
// in a program and assigns each a stable Go var name, deduplicating by the
// generated constructor code.
type constantInterner struct {
	order  []constDecl       // var declarations, in first-seen order
	byCode map[string]string // constructor code -> Go var name
	used   map[string]bool   // Go var names already handed out (collision guard)
}

type constDecl struct {
	name string
	code string
}

func newConstantInterner() *constantInterner {
	return &constantInterner{byCode: map[string]string{}, used: map[string]bool{}}
}

// ref returns the Go var name for a constant constructor, allocating one on
// first use. hint seeds a readable, valid identifier; dedup is by code.
func (c *constantInterner) ref(hint, code string) string {
	if v, ok := c.byCode[code]; ok {
		return v
	}
	base := "flag" + hint
	v := base
	for i := 1; c.used[v]; i++ {
		v = fmt.Sprintf("%s_%d", base, i)
	}
	c.byCode[code] = v
	c.used[v] = true
	c.order = append(c.order, constDecl{name: v, code: code})
	return v
}

// decls returns one package-level var definition per interned constant.
func (c *constantInterner) decls() []varDef {
	out := make([]varDef, 0, len(c.order))
	for _, d := range c.order {
		out = append(out, varDef{goName: d.name, expr: d.code})
	}
	return out
}

func sanitizeKeywordIdent(name string) string {
	var b strings.Builder
	for _, r := range name {
		if r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String()
}

// constCode returns a reference to the hoisted package-level var for a constant
// constructor when interning is enabled, otherwise the inline constructor code.
func (ctx compileContext) constCode(hint, code string) string {
	if ctx.constants == nil {
		return code
	}
	return ctx.constants.ref(hint, code)
}

// keywordCode returns the Go expression for a keyword literal, hoisting it when
// interning is enabled. Keyword var names stay flagKw_<name> for readability.
func (ctx compileContext) keywordCode(name string) string {
	return ctx.constCode("Kw_"+sanitizeKeywordIdent(name), fmt.Sprintf("%s.NewKeyword(%q)", runtimeAlias, name))
}

// stringLiteralValueCode returns the Value-constructing code for a string
// literal source, hoisting it when interning is enabled. It returns ok=false
// when source is not a plain string literal (e.g. a runtime-computed string),
// in which case the caller emits an inline NewString of its own.
func (ctx compileContext) stringLiteralValueCode(source Expr, quotedCode string) (string, bool) {
	str, ok := source.(StringExpr)
	if !ok {
		return "", false
	}
	hint := "Str_" + sanitizeKeywordIdent(truncateHint(str.Value))
	return ctx.constCode(hint, fmt.Sprintf("%s.NewString(%s)", runtimeAlias, quotedCode)), true
}

func truncateHint(s string) string {
	const max = 24
	if len(s) <= max {
		return s
	}
	return s[:max]
}

// isConstExpr reports whether expr is a compile-time constant: a literal, or a
// collection literal built entirely from constants. Symbols are constant only
// for the built-in literals true/false/nil.
func isConstExpr(expr Expr) bool {
	switch e := expr.(type) {
	case StringExpr, CharExpr, IntExpr, FloatExpr, RatioExpr, BigIntExpr, KeywordExpr, QuotedSymbolExpr, QuotedListExpr:
		return true
	case SymbolExpr:
		return e.Name == "true" || e.Name == "false" || e.Name == "nil"
	case VectorExpr:
		return allConstExprs(e.Elements)
	case SetExpr:
		return allConstExprs(e.Elements)
	case MapExpr:
		return allConstExprs(e.Entries)
	case MetaExpr:
		return isConstExpr(e.Target)
	default:
		return false
	}
}

func allConstExprs(exprs []Expr) bool {
	for _, e := range exprs {
		if !isConstExpr(e) {
			return false
		}
	}
	return true
}

type macroDef struct {
	params    []string
	restParam string
	doc       string
	body      Expr
}

//go:embed prologue.flag
var standardPrologueSource string

func newCompileContext() (compileContext, error) {
	ctx := compileContext{
		functions:     make(map[string]functionDef),
		globals:       make(map[string]exprKind),
		macros:        make(map[string]macroDef),
		moduleSymbols: make(map[string]string),
		recordTypes:   make(map[string]string),
	}
	if err := loadStandardPrologue(&ctx); err != nil {
		return compileContext{}, err
	}
	return ctx, nil
}

func loadStandardPrologue(ctx *compileContext) error {
	prologueAST, err := ParseFile(standardPrologueSource)
	if err != nil {
		return fmt.Errorf("parse compiler prologue: %w", err)
	}
	for _, form := range prologueAST.Forms {
		list, ok := form.(ListExpr)
		if !ok || len(list.Elements) == 0 {
			return fmt.Errorf("invalid compiler prologue form")
		}
		head, ok := list.Elements[0].(SymbolExpr)
		if !ok {
			return fmt.Errorf("invalid compiler prologue form")
		}
		switch head.Name {
		case "defmacro":
			name, def, err := compileDefmacro(list)
			if err != nil {
				return err
			}
			ctx.macros[name] = def
		case "defn":
			def, err := compileDefn(list, *ctx)
			if err != nil {
				return fmt.Errorf("compile compiler prologue: %w", err)
			}
			ctx.functions[def.goName] = def
			ctx.globals[def.goName] = exprKindValue
			ctx.prologueFns = append(ctx.prologueFns, def)
			ctx.prologueVars = append(ctx.prologueVars, varDef{
				flagName: def.flagName,
				goName:   def.goName,
				expr:     fmt.Sprintf("%s.NewFunction(%s)", runtimeAlias, def.variadicName),
			})
		default:
			return fmt.Errorf("invalid compiler prologue form %q", head.Name)
		}
	}
	return nil
}

func withPrologue(result compileResult, ctx compileContext) compileResult {
	if len(ctx.prologueFns) == 0 && len(ctx.prologueVars) == 0 {
		return result
	}
	result.functions = append(append([]functionDef{}, ctx.prologueFns...), result.functions...)
	result.vars = append(append([]varDef{}, ctx.prologueVars...), result.vars...)
	return result
}

// Compile translates a FLAG source string into a Go program.
// Module headers are supported; :imports require CompileProgram with a file path.
func Compile(source string) ([]byte, error) {
	result, err := compileSource(source, "")
	if err != nil {
		return nil, err
	}
	return emitGoProgram(result)
}

// CompileTokens translates a FLAG source-token stream into a Go program.
func CompileTokens(tokens <-chan SourceToken) ([]byte, error) {
	result, err := compileTokenStream(tokens, "")
	if err != nil {
		return nil, err
	}
	return emitGoProgram(result)
}

// CompileProgram loads entryPath and its import graph, then emits one Go program.
func CompileProgram(entryPath string) ([]byte, error) {
	result, err := compileProgram(entryPath, nil)
	if err != nil {
		return nil, err
	}
	return emitGoProgram(result)
}

// CompileProgramWithTests is like CompileProgram, but appends the body forms of
// each test file to the entry module (after stripping their module headers).
// That keeps same-module test access to private defs while still resolving imports.
func CompileProgramWithTests(entryPath string, testPaths []string) ([]byte, error) {
	result, err := compileProgram(entryPath, testPaths)
	if err != nil {
		return nil, err
	}
	return emitGoProgram(result)
}

type compileResult struct {
	namespace string
	typeDecls []string
	functions []functionDef
	vars      []varDef
	stmts     []mainStmt
	tests     []testCase
	needsFmt  bool
}

func emitGoProgram(result compileResult) ([]byte, error) {
	namespace, functions, vars, stmts, tests, needsFmt := result.namespace, result.functions, result.vars, result.stmts, result.tests, result.needsFmt
	if len(tests) > 0 {
		needsFmt = true
	}
	// Program entry: prefer flag_main, then main (common Clojure/Go style).
	// Match on FLAG names so module-mangled go names (e.g. c_frs_core__main) work.
	var entryFunction *functionDef
	for i := range functions {
		if functions[i].flagName == "flag_main" {
			entryFunction = &functions[i]
			break
		}
	}
	if entryFunction == nil {
		for i := range functions {
			if functions[i].flagName == "main" || functions[i].goName == "flag_main" || functions[i].goName == "main" {
				entryFunction = &functions[i]
				break
			}
		}
	}

	var out bytes.Buffer
	out.WriteString("package main\n\n")
	out.WriteString("import (\n")
	if needsFmt {
		out.WriteString("\t\"fmt\"\n")
	}
	out.WriteString("\t\"strings\"\n")
	if len(tests) > 0 || entryFunction != nil {
		out.WriteString("\t\"os\"\n")
	}
	out.WriteString("\tflagrt \"flag-lang/runtime\"\n")
	out.WriteString(")\n\n")

	if namespace != "" {
		fmt.Fprintf(&out, "// Source namespace: %s\n", namespace)
	}

	for _, decl := range result.typeDecls {
		out.WriteString(decl)
		if !strings.HasSuffix(decl, "\n") {
			out.WriteByte('\n')
		}
		out.WriteByte('\n')
	}

	for _, fn := range functions {
		if fn.doc != "" {
			writeDocComment(&out, fn.doc)
		}
		out.WriteString(renderFunctionDef(fn))
		out.WriteString("\n")
	}

	for _, v := range vars {
		if v.doc != "" {
			writeDocComment(&out, v.doc)
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
		out.WriteString("\t\t\tfmt.Printf(\"FAIL %s (line %d)\\n%s\\n%s\\n\", tc.name, tc.line, tc.body, r)\n")
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
	out.WriteString("\t_ = strings.HasSuffix\n")
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
			if stmt.kind == 0 {
				fmt.Fprintf(&out, "\t%s\n", stmt.code)
			} else {
				fmt.Fprintf(&out, "\t_ = %s\n", stmt.code)
			}
		}
		if entryFunction != nil {
			if entryFunction.hasRest || len(entryFunction.params) > 0 {
				out.WriteString("\targs := make([]flagrt.Value, 0, len(os.Args)-1)\n")
				out.WriteString("\tfor _, arg := range os.Args[1:] {\n")
				out.WriteString("\t\targs = append(args, flagrt.NewString(arg))\n")
				out.WriteString("\t}\n")
				fmt.Fprintf(&out, "\t_ = flagrt.Call(%s, args...)\n", entryFunction.goName)
			} else {
				fmt.Fprintf(&out, "\t_ = flagrt.Call(%s)\n", entryFunction.goName)
			}
		}
	}
	out.WriteString("}\n")

	formatted, err := format.Source(out.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated Go: %w", err)
	}

	return formatted, nil
}

func writeDocComment(out *bytes.Buffer, doc string) {
	for _, line := range strings.Split(doc, "\n") {
		if line == "" {
			out.WriteString("//\n")
			continue
		}
		fmt.Fprintf(out, "// %s\n", line)
	}
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
	ctx           compileContext
	loadedModules map[string]bool
	moduleDefs    map[string]map[string]string
	moduleMacros  map[string]map[string]macroDef
	modulesByPath map[string]*Module
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
		ctx:           ctx,
		loadedModules: map[string]bool{},
		moduleDefs:    map[string]map[string]string{},
		moduleMacros:  map[string]map[string]macroDef{},
		modulesByPath: map[string]*Module{},
	}
}

func (r *ReplCompiler) PrologueSetup() ReplCompiled {
	parts := make([]string, 0, len(r.ctx.prologueFns)*4)
	for _, def := range r.ctx.prologueFns {
		parts = append(parts,
			fmt.Sprintf("var %s func(args ...flagrt.Value) flagrt.Value", def.variadicName),
			fmt.Sprintf("var %s flagrt.Value", def.goName),
			fmt.Sprintf("%s = %s", def.variadicName, renderFunctionLiteral(def)),
			fmt.Sprintf("%s = flagrt.NewFunction(%s)", def.goName, def.variadicName),
		)
	}
	return ReplCompiled{Setup: strings.Join(parts, ";;")}
}

func (r *ReplCompiler) CompileLine(source string) (ReplCompiled, error) {
	ast, err := ParseFile(source)
	if err != nil {
		return ReplCompiled{}, err
	}
	if len(ast.Forms) != 1 {
		return ReplCompiled{}, fmt.Errorf("expected exactly one expression")
	}

	return r.compileReplForm(ast.Forms[0])
}

func (r *ReplCompiler) ImportSpec(source, importerPath string) ([]ReplCompiled, error) {
	ast, err := ParseFile(source)
	if err != nil {
		return nil, err
	}
	if len(ast.Forms) != 1 {
		return nil, fmt.Errorf("expected exactly one import spec")
	}
	spec, err := parseImportSpec(ast.Forms[0])
	if err != nil {
		return nil, err
	}
	return r.importResolvedSpec(spec, importerPath)
}

func (r *ReplCompiler) LoadFile(path string) ([]ReplCompiled, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve %s: %w", path, err)
	}
	mod, err := parseModuleTokenStream(absPath, TokenizeFileToChannel(absPath))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", absPath, err)
	}

	var compiled []ReplCompiled
	if mod.HasModuleHeader {
		for _, spec := range mod.Header.Imports {
			imported, err := r.importResolvedSpec(spec, absPath)
			if err != nil {
				return nil, err
			}
			compiled = append(compiled, imported...)
		}
	}

	prevNamespace := r.ctx.namespace
	switch {
	case mod.HasModuleHeader:
		r.ctx.namespace = mod.Header.Namespace
	case mod.LegacyNS != "":
		r.ctx.namespace = mod.LegacyNS
	}
	defer func() {
		r.ctx.namespace = prevNamespace
	}()
	knownFns, knownVars := r.knownBindings()
	r.ctx.allowRedefine = true
	result, _, definedMacros, err := compileModuleBody(mod, &r.ctx, true)
	if err != nil {
		return nil, err
	}
	compiled = append(compiled, r.replCompileResultSetups(result, knownFns, knownVars)...)
	if len(result.stmts) > 0 || len(result.tests) > 0 {
		compiled = append(compiled, replCompiledFromBody(result, definedMacros))
	}
	return compiled, nil
}

func (r *ReplCompiler) compileReplForm(form Expr) (ReplCompiled, error) {
	knownFns, knownVars := r.knownBindings()
	r.ctx.allowRedefine = true
	result, _, definedMacros, err := compileModuleBody(&Module{Forms: []Expr{form}}, &r.ctx, true)
	if err != nil {
		return ReplCompiled{}, err
	}
	steps := r.replCompileResultSetups(result, knownFns, knownVars)
	out := replCompiledFromBody(result, definedMacros)
	setupParts := make([]string, 0, len(steps)+1)
	for _, step := range steps {
		if step.Setup != "" {
			setupParts = append(setupParts, step.Setup)
		}
	}
	if out.Setup != "" {
		setupParts = append(setupParts, out.Setup)
		out.Setup = ""
	}
	out.Setup = strings.Join(setupParts, ";;")
	return out, nil
}

func (r *ReplCompiler) knownBindings() (map[string]functionDef, map[string]exprKind) {
	fns := make(map[string]functionDef, len(r.ctx.functions))
	for name, def := range r.ctx.functions {
		fns[name] = def
	}
	vars := make(map[string]exprKind, len(r.ctx.globals))
	for name, kind := range r.ctx.globals {
		vars[name] = kind
	}
	return fns, vars
}

func (r *ReplCompiler) importResolvedSpec(spec ImportSpec, importerPath string) ([]ReplCompiled, error) {
	resolved, err := resolveImportPath(importerPath, spec.Path)
	if err != nil {
		return nil, fmt.Errorf("import %q: %w", spec.Path, err)
	}
	loaded, err := r.ensureModuleLoaded(resolved, map[string]bool{})
	if err != nil {
		return nil, err
	}
	spec.Path = resolved
	mod := &Module{
		Path:            importerPath,
		HasModuleHeader: true,
		Header: ModuleHeader{
			Imports: []ImportSpec{spec},
		},
	}
	if err := seedImports(&r.ctx, mod, r.modulesByPath, r.moduleDefs, r.moduleMacros); err != nil {
		return nil, err
	}
	return loaded, nil
}

func (r *ReplCompiler) ensureModuleLoaded(path string, loading map[string]bool) ([]ReplCompiled, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve %s: %w", path, err)
	}
	if r.loadedModules[absPath] {
		return nil, nil
	}
	if loading[absPath] {
		return nil, fmt.Errorf("circular import involving %s", absPath)
	}
	loading[absPath] = true
	defer delete(loading, absPath)

	mod, err := parseModuleTokenStream(absPath, TokenizeFileToChannel(absPath))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", absPath, err)
	}
	if !mod.HasModuleHeader {
		return nil, fmt.Errorf("import %q: target %s has no module header", absPath, absPath)
	}
	r.modulesByPath[absPath] = mod

	var compiled []ReplCompiled
	for _, spec := range mod.Header.Imports {
		resolved, err := resolveImportPath(absPath, spec.Path)
		if err != nil {
			return nil, fmt.Errorf("%s: import %q: %w", absPath, spec.Path, err)
		}
		loaded, err := r.ensureModuleLoaded(resolved, loading)
		if err != nil {
			return nil, err
		}
		compiled = append(compiled, loaded...)
	}

	moduleCtx := copyCompileContext(r.ctx)
	moduleCtx.namespace = mod.Header.Namespace
	moduleCtx.moduleSymbols = map[string]string{}
	if err := seedImports(&moduleCtx, mod, r.modulesByPath, r.moduleDefs, r.moduleMacros); err != nil {
		return nil, fmt.Errorf("%s: %w", absPath, err)
	}
	result, defined, definedMacros, err := compileModuleBody(mod, &moduleCtx, false)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", absPath, err)
	}
	if err := validateExports(mod, defined, definedMacros); err != nil {
		return nil, fmt.Errorf("%s: %w", absPath, err)
	}

	compiled = append(compiled, r.replCompileResultSetups(result, r.ctx.functions, r.ctx.globals)...)
	r.moduleDefs[absPath] = defined
	r.moduleMacros[absPath] = definedMacros
	for k, v := range moduleCtx.functions {
		r.ctx.functions[k] = v
	}
	for k, v := range moduleCtx.globals {
		r.ctx.globals[k] = v
	}
	r.loadedModules[absPath] = true
	return compiled, nil
}

func (r *ReplCompiler) replCompileResultSetups(result compileResult, knownFns map[string]functionDef, knownVars map[string]exprKind) []ReplCompiled {
	out := make([]ReplCompiled, 0, len(result.typeDecls)+len(result.functions)+len(result.vars))
	for _, decl := range result.typeDecls {
		decl = strings.TrimSpace(decl)
		if decl != "" {
			out = append(out, ReplCompiled{Setup: decl})
		}
	}
	for _, def := range result.functions {
		_, exists := knownFns[def.goName]
		if def.goSignature != "" {
			out = append(out, ReplCompiled{Setup: strings.TrimSpace(renderFunctionDef(def))})
			continue
		}
		setupParts := make([]string, 0, 6)
		if !exists {
			setupParts = append(setupParts,
				fmt.Sprintf("var %s %s", def.arityName, renderDirectFunctionType(def)),
				fmt.Sprintf("var %s func(args ...flagrt.Value) flagrt.Value", def.variadicName),
			)
		}
		setupParts = append(setupParts,
			fmt.Sprintf("%s = %s", def.arityName, renderDirectFunctionLiteral(def)),
			fmt.Sprintf("%s = %s", def.variadicName, renderVariadicFunctionLiteral(def)),
		)
		out = append(out, ReplCompiled{Setup: strings.Join(setupParts, ";;")})
	}
	for _, binding := range result.vars {
		setup := fmt.Sprintf("%s = %s", binding.goName, binding.expr)
		if _, exists := knownVars[binding.goName]; !exists {
			setup = fmt.Sprintf("var %s flagrt.Value;;%s = %s", binding.goName, binding.goName, binding.expr)
		}
		out = append(out, ReplCompiled{Setup: setup})
	}
	return out
}

func replCompiledFromBody(result compileResult, definedMacros map[string]macroDef) ReplCompiled {
	stmtParts := make([]string, 0, len(result.stmts)+len(result.tests))
	var resultExpr string
	for _, stmt := range result.stmts {
		if stmt.kind == 0 {
			stmtParts = append(stmtParts, stmt.code)
			continue
		}
		if stmt.kind == exprKindValue {
			resultExpr = fmt.Sprintf("%s.ValueToAny(%s)", runtimeAlias, stmt.code)
		} else {
			resultExpr = stmt.code
		}
	}
	for _, tc := range result.tests {
		stmtParts = append(stmtParts, tc.goName+"()")
	}
	if resultExpr == "" && len(result.stmts) == 0 && len(result.tests) == 0 {
		if len(result.functions) > 0 {
			resultExpr = fmt.Sprintf("%q", result.functions[len(result.functions)-1].flagName)
		} else if len(result.vars) > 0 {
			resultExpr = fmt.Sprintf("%s.ValueToAny(%s)", runtimeAlias, result.vars[len(result.vars)-1].goName)
		} else {
			for name := range definedMacros {
				resultExpr = fmt.Sprintf("%q", name)
				break
			}
		}
	}
	return ReplCompiled{Setup: strings.Join(stmtParts, ";;"), ResultExpr: resultExpr}
}

func (r *ReplCompiler) replGoExportSetups(mod *Module) ([]ReplCompiled, error) {
	if mod == nil || !mod.HasModuleHeader || len(mod.Header.GoExports) == 0 {
		return nil, nil
	}
	out := make([]ReplCompiled, 0, len(mod.Header.GoExports))
	for localName, hostKey := range mod.Header.GoExports {
		rtIdent, ok := libraryGoBinds[hostKey]
		if !ok {
			return nil, fmt.Errorf("unknown :go-exports host bind %q for %s", hostKey, localName)
		}
		goName, err := moduleGoIdent(r.ctx.namespace, localName)
		if err != nil {
			return nil, err
		}
		if err := r.ctx.bindModuleName(localName, goName); err != nil {
			return nil, err
		}
		_, exists := r.ctx.globals[goName]
		r.ctx.globals[goName] = exprKindValue
		setup := fmt.Sprintf("%s = %s.%s", goName, runtimeAlias, rtIdent)
		if !exists {
			setup = fmt.Sprintf("var %s flagrt.Value;;%s = %s.%s", goName, goName, runtimeAlias, rtIdent)
		}
		out = append(out, ReplCompiled{Setup: setup})
	}
	return out, nil
}

func compileSource(source, path string) (compileResult, error) {
	mod, err := parseModuleFile(path, source)
	if err != nil {
		return compileResult{}, err
	}
	return compileParsedModule(mod, path)
}

func compileTokenStream(tokens <-chan SourceToken, path string) (compileResult, error) {
	mod, err := parseModuleTokenStream(path, tokens)
	if err != nil {
		return compileResult{}, err
	}
	return compileParsedModule(mod, path)
}

func compileParsedModule(mod *Module, path string) (compileResult, error) {
	if mod.HasModuleHeader && len(mod.Header.Imports) > 0 {
		if path == "" {
			return compileResult{}, fmt.Errorf("module imports require compiling from a file path (use flag-lang build <file.flag>)")
		}
		// Single-file compile with imports: load full program from this path.
		return compileProgram(path, nil)
	}

	ctx, err := newCompileContext()
	if err != nil {
		return compileResult{}, err
	}
	ctx.constants = newConstantInterner()
	if mod.HasModuleHeader {
		ctx.namespace = mod.Header.Namespace
	}

	result, defined, definedMacros, err := compileModuleBody(mod, &ctx, true)
	if err != nil {
		return compileResult{}, err
	}
	if mod.HasModuleHeader {
		if err := validateExports(mod, defined, definedMacros); err != nil {
			return compileResult{}, err
		}
		result.namespace = mod.Header.Namespace
	} else if mod.LegacyNS != "" {
		result.namespace = mod.LegacyNS
	}
	result.vars = append(result.vars, ctx.constants.decls()...)
	return withPrologue(result, ctx), nil
}

func compileProgram(entryPath string, testPaths []string) (compileResult, error) {
	prog, err := LoadProgram(entryPath)
	if err != nil {
		return compileResult{}, err
	}
	if len(testPaths) > 0 {
		if err := appendTestForms(prog.Entry, testPaths); err != nil {
			return compileResult{}, err
		}
	}

	// Shared environment across modules: functions/globals accumulate; each
	// module gets its own moduleSymbols seeded from imports.
	shared, err := newCompileContext()
	if err != nil {
		return compileResult{}, err
	}
	shared.constants = newConstantInterner()

	// path -> bare local name -> go ident for all function/var defs in that module
	moduleDefs := map[string]map[string]string{}
	// path -> exported macros defined in that module
	moduleMacros := map[string]map[string]macroDef{}
	byPath := prog.byPath

	var allTypeDecls []string
	var allFunctions []functionDef
	var allVars []varDef
	var allStmts []mainStmt
	var allTests []testCase
	needsFmt := false
	entryNS := displayNamespace(prog.Entry)

	for _, mod := range prog.Modules {
		ctx := copyCompileContext(shared)
		ctx.constants = shared.constants
		if mod.HasModuleHeader {
			ctx.namespace = mod.Header.Namespace
			if err := seedImports(&ctx, mod, byPath, moduleDefs, moduleMacros); err != nil {
				return compileResult{}, err
			}
		}

		allowTopLevel := mod == prog.Entry
		partial, defined, definedMacros, err := compileModuleBody(mod, &ctx, allowTopLevel)
		if err != nil {
			if mod.Path != "" {
				return compileResult{}, fmt.Errorf("%s: %w", mod.Path, err)
			}
			return compileResult{}, err
		}
		if mod.HasModuleHeader {
			if err := validateExports(mod, defined, definedMacros); err != nil {
				return compileResult{}, fmt.Errorf("%s: %w", mod.Path, err)
			}
		}

		defs := map[string]string{}
		for name, goName := range defined {
			defs[name] = goName
		}
		moduleDefs[mod.Path] = defs
		moduleMacros[mod.Path] = definedMacros

		// Functions/globals accumulate for cross-module Go linkage; macros do not
		// leak into shared — only seedImports installs them for importers.
		for k, v := range ctx.functions {
			shared.functions[k] = v
		}
		for k, v := range ctx.globals {
			shared.globals[k] = v
		}

		allTypeDecls = append(allTypeDecls, partial.typeDecls...)
		allFunctions = append(allFunctions, partial.functions...)
		allVars = append(allVars, partial.vars...)
		allStmts = append(allStmts, partial.stmts...)
		allTests = append(allTests, partial.tests...)
		needsFmt = needsFmt || partial.needsFmt
	}

	allVars = append(allVars, shared.constants.decls()...)
	return withPrologue(compileResult{
		namespace: entryNS,
		typeDecls: allTypeDecls,
		functions: allFunctions,
		vars:      allVars,
		stmts:     allStmts,
		tests:     allTests,
		needsFmt:  needsFmt,
	}, shared), nil
}

func seedImports(ctx *compileContext, mod *Module, byPath map[string]*Module, moduleDefs map[string]map[string]string, moduleMacros map[string]map[string]macroDef) error {
	if !mod.HasModuleHeader {
		return nil
	}
	usedPrefixes := map[string]string{} // prefix -> import path
	for _, spec := range mod.Header.Imports {
		resolved, err := resolveImportPath(mod.Path, spec.Path)
		if err != nil {
			return fmt.Errorf("import %q: %w", spec.Path, err)
		}
		provider, ok := byPath[resolved]
		if !ok {
			return fmt.Errorf("import %q: module not loaded", spec.Path)
		}
		if !provider.HasModuleHeader {
			return fmt.Errorf("import %q: target %s has no module header", spec.Path, resolved)
		}
		prefix := spec.As
		if prefix == "" {
			prefix = provider.Header.Namespace
		}
		if prev, ok := usedPrefixes[prefix]; ok && prev != resolved {
			return fmt.Errorf("import prefix %q already used by %s", prefix, prev)
		}
		usedPrefixes[prefix] = resolved

		exports := moduleExportSet(provider)
		defs := moduleDefs[resolved]
		if defs == nil {
			return fmt.Errorf("import %q: provider not compiled yet", spec.Path)
		}
		macros := moduleMacros[resolved]
		if macros == nil {
			macros = map[string]macroDef{}
		}

		for exp := range exports {
			if goName, ok := defs[exp]; ok {
				qual := prefix + "/" + exp
				if err := ctx.bindModuleName(qual, goName); err != nil {
					return err
				}
				continue
			}
			if mac, ok := macros[exp]; ok {
				// Qualified macro name so expansions can use async/go without :refer.
				ctx.macros[prefix+"/"+exp] = mac
				continue
			}
			return fmt.Errorf("import %q: exported %q has no definition", spec.Path, exp)
		}
		for _, ref := range spec.Refer {
			if !exports[ref] {
				return fmt.Errorf("import %q: cannot :refer %q (not in :exports)", spec.Path, ref)
			}
			if goName, ok := defs[ref]; ok {
				if err := ctx.bindModuleName(ref, goName); err != nil {
					return fmt.Errorf("import %q: :refer %q: %w", spec.Path, ref, err)
				}
				continue
			}
			if mac, ok := macros[ref]; ok {
				// :refer installs bare macro name; ok to overwrite same-name macro.
				ctx.macros[ref] = mac
				continue
			}
			return fmt.Errorf("import %q: referred %q has no definition", spec.Path, ref)
		}
	}
	return nil
}

func validateExports(mod *Module, defined map[string]string, definedMacros map[string]macroDef) error {
	if !mod.HasModuleHeader {
		return nil
	}
	for _, name := range mod.Header.Exports {
		if _, ok := defined[name]; ok {
			continue
		}
		if _, ok := definedMacros[name]; ok {
			continue
		}
		return fmt.Errorf("export %q is not defined in module %q", name, mod.Header.Namespace)
	}
	return nil
}

// compileModuleBody compiles the body forms of a module into ctx.
// defined maps bare FLAG local names defined in this module to Go idents.
// definedMacros maps macro names defined in this module.
func compileModuleBody(mod *Module, ctx *compileContext, allowTopLevel bool) (compileResult, map[string]string, map[string]macroDef, error) {
	typeDecls := make([]string, 0)
	functions := make([]functionDef, 0, len(mod.Forms))
	vars := make([]varDef, 0, len(mod.Forms))
	stmts := make([]mainStmt, 0, len(mod.Forms))
	tests := make([]testCase, 0, len(mod.Forms))
	needsFmt := false
	defined := map[string]string{}
	definedMacros := map[string]macroDef{}

	// Native library re-exports (:go-exports) become package-level vars bound to
	// generated flagrt.GoBind_* adapters.
	if mod.HasModuleHeader {
		for localName, hostKey := range mod.Header.GoExports {
			rtIdent, ok := libraryGoBinds[hostKey]
			if !ok {
				return compileResult{}, nil, nil, fmt.Errorf("unknown :go-exports host bind %q for %s", hostKey, localName)
			}
			goName, err := moduleGoIdent(ctx.namespace, localName)
			if err != nil {
				return compileResult{}, nil, nil, err
			}
			if _, exists := ctx.globals[goName]; exists {
				return compileResult{}, nil, nil, fmt.Errorf("symbol %q already defined", localName)
			}
			if err := ctx.bindModuleName(localName, goName); err != nil {
				return compileResult{}, nil, nil, err
			}
			ctx.globals[goName] = exprKindValue
			defined[localName] = goName
			vars = append(vars, varDef{
				flagName: localName,
				goName:   goName,
				expr:     fmt.Sprintf("%s.%s", runtimeAlias, rtIdent),
			})
		}
	}

	pendingForms := append([]Expr(nil), mod.Forms...)
	for i := 0; i < len(pendingForms); i++ {
		form := pendingForms[i]
		if list, ok := form.(ListExpr); ok && len(list.Elements) > 0 {
			if head, ok := list.Elements[0].(SymbolExpr); ok && head.Name == "defmacro" {
				name, def, err := compileDefmacro(list)
				if err != nil {
					return compileResult{}, nil, nil, err
				}
				ctx.macros[name] = def
				definedMacros[name] = def
				continue
			}
		}

		expanded, err := macroExpand(form, *ctx, 0)
		if err != nil {
			return compileResult{}, nil, nil, err
		}

		form = expanded
		list, ok := form.(ListExpr)
		if !ok || len(list.Elements) == 0 {
			if err := appendTopLevelExpr(form, *ctx, allowTopLevel, &stmts); err != nil {
				return compileResult{}, nil, nil, err
			}
			continue
		}

		head, ok := list.Elements[0].(SymbolExpr)
		if !ok {
			if err := appendTopLevelExpr(form, *ctx, allowTopLevel, &stmts); err != nil {
				return compileResult{}, nil, nil, err
			}
			continue
		}

		switch head.Name {
		case "ns":
			return compileResult{}, nil, nil, exprError(list, "ns must be the first form (or use a module header map)")
		case "defn-":
			return compileResult{}, nil, nil, exprError(list, "defn- is not supported; list public names in the module :exports instead")
		case "defn":
			def, err := compileDefn(list, *ctx)
			if err != nil {
				return compileResult{}, nil, nil, err
			}
			if _, exists := ctx.functions[def.goName]; exists && !ctx.allowRedefine {
				return compileResult{}, nil, nil, fmt.Errorf("function %q already defined", def.goName)
			}
			if err := ctx.bindModuleName(def.flagName, def.goName); err != nil {
				return compileResult{}, nil, nil, err
			}
			ctx.functions[def.goName] = def
			ctx.globals[def.goName] = exprKindValue
			defined[def.flagName] = def.goName
			functions = append(functions, def)
			vars = append(vars, varDef{
				flagName: def.flagName,
				goName:   def.goName,
				expr:     fmt.Sprintf("%s.NewFunction(%s)", runtimeAlias, def.variadicName),
			})
		case "deftest":
			def, err := compileDeftest(list, *ctx)
			if err != nil {
				return compileResult{}, nil, nil, err
			}
			if _, exists := ctx.functions[def.goName]; exists && !ctx.allowRedefine {
				return compileResult{}, nil, nil, fmt.Errorf("test %q already defined", def.goName)
			}
			ctx.functions[def.goName] = def
			functions = append(functions, def)
			testName := def.flagName
			if testName == "" {
				testName = def.goName
			}
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
			binding, kind, err := compileDef(list, *ctx)
			if err != nil {
				return compileResult{}, nil, nil, err
			}
			if err := ctx.bindModuleName(binding.flagName, binding.goName); err != nil {
				return compileResult{}, nil, nil, err
			}
			ctx.globals[binding.goName] = kind
			defined[binding.flagName] = binding.goName
			vars = append(vars, binding)
		case "defmacro":
			name, def, err := compileDefmacro(list)
			if err != nil {
				return compileResult{}, nil, nil, err
			}
			ctx.macros[name] = def
			definedMacros[name] = def
		case "println":
			if !allowTopLevel {
				return compileResult{}, nil, nil, exprError(list, "top-level expressions are only allowed in the entry module")
			}
			needsFmt = true
			arg, err := strArgExprForGoCall(list.Elements[1:], *ctx, nil)
			if err != nil {
				return compileResult{}, nil, nil, err
			}
			stmts = append(stmts, mainStmt{code: fmt.Sprintf("fmt.Println(%s)", arg)})
		case "print":
			if !allowTopLevel {
				return compileResult{}, nil, nil, exprError(list, "top-level expressions are only allowed in the entry module")
			}
			needsFmt = true
			arg, err := argumentExprForGoCall(list.Elements[1:], *ctx, nil)
			if err != nil {
				return compileResult{}, nil, nil, err
			}
			stmts = append(stmts, mainStmt{code: fmt.Sprintf("fmt.Print(%s)", arg)})
		case "do":
			if len(list.Elements) > 1 {
				inner := make([]Expr, 0, len(list.Elements)-1)
				inner = append(inner, list.Elements[1:]...)
				pendingForms = append(pendingForms[:i+1], append(inner, pendingForms[i+1:]...)...)
			}
		case "go-interface":
			wrapper, err := compileGoInterface(list, *ctx, defined)
			if err != nil {
				return compileResult{}, nil, nil, err
			}
			functions = append(functions, wrapper)
		case "defrecord", "defrecord*":
			typeDecl, ctors, recordVars, err := compileDefrecord(list, ctx)
			if err != nil {
				return compileResult{}, nil, nil, err
			}
			typeDecls = append(typeDecls, typeDecl)
			functions = append(functions, ctors...)
			vars = append(vars, recordVars...)
			for _, ctor := range ctors {
				defined[ctor.flagName] = ctor.goName
			}
		default:
			if !allowTopLevel {
				return compileResult{}, nil, nil, exprError(list, "top-level expressions are only allowed in the entry module")
			}
			if err := appendTopLevelExpr(form, *ctx, allowTopLevel, &stmts); err != nil {
				return compileResult{}, nil, nil, err
			}
		}
	}

	return compileResult{
		typeDecls: typeDecls,
		functions: functions,
		vars:      vars,
		stmts:     stmts,
		tests:     tests,
		needsFmt:  needsFmt,
	}, defined, definedMacros, nil
}

func appendTopLevelExpr(form Expr, ctx compileContext, allowTopLevel bool, stmts *[]mainStmt) error {
	if !allowTopLevel {
		return exprError(form, "top-level expressions are only allowed in the entry module")
	}
	expr, err := exprToGo(form, ctx, nil)
	if err != nil {
		return err
	}
	*stmts = append(*stmts, mainStmt{code: expr.code, kind: expr.kind})
	return nil
}

func compileDefn(form ListExpr, ctx compileContext) (functionDef, error) {
	if len(form.Elements) < 4 {
		return functionDef{}, exprError(form, "defn expects name, optional docstring, vector params, and body")
	}

	nameExpr, ok := unwrapMetaExpr(form.Elements[1]).(SymbolExpr)
	if !ok || nameExpr.Name == "" {
		return functionDef{}, exprError(form, "defn expects a function name")
	}

	doc := ""
	paramsIndex := 2
	bodyStartIndex := 3
	if docExpr, ok := form.Elements[2].(StringExpr); ok {
		doc = docExpr.Value
		paramsIndex = 3
		bodyStartIndex = 4
	}

	goName, err := moduleGoIdent(ctx.namespace, nameExpr.Name)
	if err != nil {
		return functionDef{}, err
	}

	paramsExpr, ok := form.Elements[paramsIndex].(VectorExpr)
	if !ok {
		return functionDef{}, exprError(form.Elements[paramsIndex], "defn expects a parameter vector")
	}
	params, localSymbols, localInits, hasRest, err := bindLambdaParams(paramsExpr, ctx, nil, "defn")
	if err != nil {
		return functionDef{}, err
	}
	if len(form.Elements) <= bodyStartIndex {
		return functionDef{}, exprError(form, "defn expects name, optional docstring, vector params, and body")
	}

	fnCtx := copyCompileContext(ctx)
	fnCtx.selfFunctionName = nameExpr.Name
	fnCtx.selfFunctionArity = len(params)
	fnCtx.selfArityName = fmt.Sprintf("%s_arity_%d", goName, len(params))
	if hasRest {
		fnCtx.selfArityName = goName + "_variadic"
	}
	fnCtx.globals[goName] = exprKindValue
	fnCtx.functions[goName] = functionDef{
		flagName:     nameExpr.Name,
		goName:       goName,
		variadicName: goName + "_variadic",
		arityName:    fnCtx.selfArityName,
		hasRest:      hasRest,
		doc:          doc,
		params:       params,
	}
	// Allow recursive bare / qualified references inside the body.
	_ = fnCtx.bindModuleName(nameExpr.Name, goName)

	bodyExprs := form.Elements[bodyStartIndex:]
	body, err := doExprToGo(bodyExprs, fnCtx, localSymbols)
	if err != nil {
		return functionDef{}, err
	}
	bodySource := bodyExprs[len(bodyExprs)-1]
	body, err = coerceExprToValue(body, bodySource, "defn body", ctx)
	if err != nil {
		return functionDef{}, err
	}

	return functionDef{
		flagName:     nameExpr.Name,
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
		return varDef{}, 0, false, exprError(form, "def expects name, optional docstring, and value")
	}

	nameExpr, ok := unwrapMetaExpr(form.Elements[1]).(SymbolExpr)
	if !ok || nameExpr.Name == "" {
		return varDef{}, 0, false, exprError(form, "def expects a symbol name")
	}
	doc := ""
	valueIndex := 2
	if len(form.Elements) == 4 {
		docExpr, ok := form.Elements[2].(StringExpr)
		if !ok {
			return varDef{}, 0, false, exprError(form.Elements[2], "def docstring must be a string")
		}
		doc = docExpr.Value
		valueIndex = 3
	}
	goName, err := moduleGoIdent(ctx.namespace, nameExpr.Name)
	if err != nil {
		return varDef{}, 0, false, err
	}

	valueExpr, err := exprToGo(form.Elements[valueIndex], ctx, nil)
	if err != nil {
		return varDef{}, 0, false, err
	}
	valueExpr, err = coerceExprToValue(valueExpr, form.Elements[valueIndex], "def value", ctx)
	if err != nil {
		return varDef{}, 0, false, err
	}

	_, exists := ctx.globals[goName]
	return varDef{flagName: nameExpr.Name, goName: goName, doc: doc, expr: valueExpr.code}, valueExpr.kind, !exists, nil
}

func compileDefmacro(form ListExpr) (string, macroDef, error) {
	if len(form.Elements) != 4 && len(form.Elements) != 5 {
		return "", macroDef{}, fmt.Errorf("defmacro expects name, optional docstring, vector params, and body")
	}
	nameExpr, ok := unwrapMetaExpr(form.Elements[1]).(SymbolExpr)
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
		sym, ok := unwrapMetaExpr(paramsExpr.Elements[i]).(SymbolExpr)
		if !ok || sym.Name == "" {
			return "", macroDef{}, fmt.Errorf("defmacro parameters must be symbols")
		}
		if sym.Name == "&" {
			if restParam != "" || i != len(paramsExpr.Elements)-2 {
				return "", macroDef{}, fmt.Errorf("defmacro varargs must use [& name] at end")
			}
			next, ok := unwrapMetaExpr(paramsExpr.Elements[i+1]).(SymbolExpr)
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

	nameExpr, ok := unwrapMetaExpr(form.Elements[1]).(SymbolExpr)
	if !ok || nameExpr.Name == "" {
		return functionDef{}, fmt.Errorf("deftest expects a test name")
	}

	goName, err := moduleGoIdent(ctx.namespace, nameExpr.Name)
	if err != nil {
		return functionDef{}, err
	}

	body, err := testingBodyExprToGo(form.Elements[2:], ctx, nil)
	if err != nil {
		return functionDef{}, err
	}

	return functionDef{
		flagName:     nameExpr.Name,
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
	case MetaExpr:
		meta, err := macroExpand(value.Meta, ctx, depth)
		if err != nil {
			return nil, err
		}
		target, err := macroExpand(value.Target, ctx, depth)
		if err != nil {
			return nil, err
		}
		return MetaExpr{Meta: meta, Target: target, Line: value.Line, Col: value.Col}, nil
	default:
		return expr, nil
	}
}

type macroLiteralExpr struct {
	inner Expr
}

func (macroLiteralExpr) expr() {}

func quoteMacro(expr Expr) Expr {
	if _, ok := expr.(macroLiteralExpr); ok {
		return expr
	}
	return macroLiteralExpr{inner: expr}
}

func unquoteMacro(expr Expr) Expr {
	for {
		quoted, ok := expr.(macroLiteralExpr)
		if !ok {
			return expr
		}
		expr = quoted.inner
	}
}

func unquoteMacroTree(expr Expr) Expr {
	switch value := unquoteMacro(expr).(type) {
	case ListExpr:
		out := make([]Expr, 0, len(value.Elements))
		for _, item := range value.Elements {
			out = append(out, unquoteMacroTree(item))
		}
		return ListExpr{Elements: out, Line: value.Line, Col: value.Col}
	case VectorExpr:
		out := make([]Expr, 0, len(value.Elements))
		for _, item := range value.Elements {
			out = append(out, unquoteMacroTree(item))
		}
		return VectorExpr{Elements: out, Line: value.Line, Col: value.Col}
	case MapExpr:
		out := make([]Expr, 0, len(value.Entries))
		for _, item := range value.Entries {
			out = append(out, unquoteMacroTree(item))
		}
		return MapExpr{Entries: out, Line: value.Line, Col: value.Col}
	case SetExpr:
		out := make([]Expr, 0, len(value.Elements))
		for _, item := range value.Elements {
			out = append(out, unquoteMacroTree(item))
		}
		return SetExpr{Elements: out, Line: value.Line, Col: value.Col}
	case HashFnExpr:
		return HashFnExpr{Body: unquoteMacroTree(value.Body), Line: value.Line, Col: value.Col}
	case MetaExpr:
		return MetaExpr{Meta: unquoteMacroTree(value.Meta), Target: unquoteMacroTree(value.Target), Line: value.Line, Col: value.Col}
	default:
		return value
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
	restBindings := map[string][]Expr{}
	if m.restParam != "" {
		restArgs := make([]Expr, 0, len(args)-len(m.params))
		for _, arg := range args[len(m.params):] {
			restArgs = append(restArgs, copyExpr(arg))
		}
		restBindings[m.restParam] = restArgs
	}
	expanded, err := substituteMacroExpr(m.body, values, restBindings)
	if err != nil {
		return nil, err
	}
	return unquoteMacroTree(expanded), nil
}

func substituteMacroExpr(expr Expr, values map[string]Expr, restBindings map[string][]Expr) (Expr, error) {
	if _, ok := expr.(macroLiteralExpr); ok {
		return expr, nil
	}
	switch value := expr.(type) {
	case SymbolExpr:
		if replacement, ok := values[value.Name]; ok {
			return quoteMacro(copyExpr(replacement)), nil
		}
		return value, nil
	case ListExpr:
		out := make([]Expr, 0, len(value.Elements))
		for _, item := range value.Elements {
			sym, isSym := unquoteMacro(item).(SymbolExpr)
			if isSym {
				if restArgs, ok := restBindings[sym.Name]; ok {
					for _, restArg := range restArgs {
						out = append(out, quoteMacro(copyExpr(restArg)))
					}
					continue
				}
			}
			sub, err := substituteMacroExpr(item, values, restBindings)
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
			sym, isSym := unquoteMacro(item).(SymbolExpr)
			if isSym {
				if restArgs, ok := restBindings[sym.Name]; ok {
					for _, restArg := range restArgs {
						out = append(out, quoteMacro(copyExpr(restArg)))
					}
					continue
				}
			}
			sub, err := substituteMacroExpr(item, values, restBindings)
			if err != nil {
				return nil, err
			}
			out = append(out, sub)
		}
		return VectorExpr{Elements: out}, nil
	case MapExpr:
		out := make([]Expr, 0, len(value.Entries))
		for _, item := range value.Entries {
			sub, err := substituteMacroExpr(item, values, restBindings)
			if err != nil {
				return nil, err
			}
			out = append(out, sub)
		}
		return MapExpr{Entries: out}, nil
	case SetExpr:
		out := make([]Expr, 0, len(value.Elements))
		for _, item := range value.Elements {
			sub, err := substituteMacroExpr(item, values, restBindings)
			if err != nil {
				return nil, err
			}
			out = append(out, sub)
		}
		return SetExpr{Elements: out}, nil
	case HashFnExpr:
		body, err := substituteMacroExpr(value.Body, values, restBindings)
		if err != nil {
			return nil, err
		}
		return HashFnExpr{Body: body}, nil
	case MetaExpr:
		meta, err := substituteMacroExpr(value.Meta, values, restBindings)
		if err != nil {
			return nil, err
		}
		target, err := substituteMacroExpr(value.Target, values, restBindings)
		if err != nil {
			return nil, err
		}
		return MetaExpr{Meta: meta, Target: target, Line: value.Line, Col: value.Col}, nil
	default:
		return expr, nil
	}
}

func applyMacroBuiltin(name string, args []Expr) (Expr, bool, error) {
	switch name {
	case "macro-case":
		expanded, err := expandMacroCase(args)
		return expanded, true, err
	case "macro-defrecord":
		expanded, err := expandDefrecordMacro(args)
		return expanded, true, err
	default:
		return nil, false, nil
	}
}

func expandDefrecordMacro(args []Expr) (Expr, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("macro-defrecord expects a record name and field vector")
	}
	return ListExpr{
		Elements: []Expr{
			SymbolExpr{Name: "defrecord*"},
			args[0],
			args[1],
		},
	}, nil
}

func copyExpr(expr Expr) Expr {
	switch value := expr.(type) {
	case macroLiteralExpr:
		return macroLiteralExpr{inner: copyExpr(value.inner)}
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
	case MetaExpr:
		return MetaExpr{Meta: copyExpr(value.Meta), Target: copyExpr(value.Target), Line: value.Line, Col: value.Col}
	default:
		return expr
	}
}

func expandMacroCase(args []Expr) (Expr, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("macro-case expects a target form and at least one clause")
	}

	targetList, ok := unquoteMacro(args[0]).(ListExpr)
	if !ok {
		return nil, fmt.Errorf("macro-case expects a list target form")
	}
	clauses := args[1:]
	target := make([]Expr, 0, len(targetList.Elements))
	if len(targetList.Elements) > 0 {
		target = targetList.Elements[1:]
	}

	for _, clauseExpr := range clauses {
		clause, ok := clauseExpr.(VectorExpr)
		if !ok || len(clause.Elements) != 2 {
			return nil, fmt.Errorf("macro-case clauses must be [pattern body] vectors")
		}
		bindings := map[string]Expr{}
		restBindings := map[string][]Expr{}
		matched, err := matchMacroPattern(clause.Elements[0], target, bindings, restBindings)
		if err != nil {
			return nil, err
		}
		if !matched {
			continue
		}
		return substituteMacroExpr(clause.Elements[1], bindings, restBindings)
	}

	return nil, fmt.Errorf("macro-case had no matching clause")
}

func matchMacroPattern(pattern Expr, target []Expr, bindings map[string]Expr, restBindings map[string][]Expr) (bool, error) {
	if len(target) > 0 {
		unquoted := make([]Expr, len(target))
		for i, item := range target {
			unquoted[i] = unquoteMacro(item)
		}
		target = unquoted
	}
	pattern = unquoteMacro(pattern)
	switch pat := pattern.(type) {
	case SymbolExpr:
		switch pat.Name {
		case "_":
			return len(target) == 1, nil
		case "&":
			return false, fmt.Errorf("macro-case pattern cannot use bare &")
		default:
			if len(target) != 1 {
				return false, nil
			}
			if existing, ok := bindings[pat.Name]; ok {
				return exprStructEqual(existing, target[0]), nil
			}
			bindings[pat.Name] = copyExpr(target[0])
			return true, nil
		}
	case KeywordExpr, StringExpr, CharExpr, IntExpr, BigIntExpr, FloatExpr, RatioExpr, QuotedSymbolExpr:
		if len(target) != 1 {
			return false, nil
		}
		return exprStructEqual(pat, target[0]), nil
	case ListExpr:
		if len(target) == 1 {
			if listTarget, ok := unwrapMetaExpr(target[0]).(ListExpr); ok {
				return matchMacroSequence(pat.Elements, listTarget.Elements, bindings, restBindings)
			}
		}
		return matchMacroSequence(pat.Elements, target, bindings, restBindings)
	case VectorExpr:
		if len(target) == 1 {
			if vectorTarget, ok := unwrapMetaExpr(target[0]).(VectorExpr); ok {
				return matchMacroSequence(pat.Elements, vectorTarget.Elements, bindings, restBindings)
			}
		}
		return matchMacroSequence(pat.Elements, target, bindings, restBindings)
	case MapExpr, SetExpr, HashFnExpr, MetaExpr:
		if len(target) != 1 {
			return false, nil
		}
		return exprStructEqual(pat, target[0]), nil
	default:
		if len(target) != 1 {
			return false, nil
		}
		return exprStructEqual(pat, target[0]), nil
	}
}

func exprStructEqual(a, b Expr) bool {
	switch av := a.(type) {
	case SymbolExpr:
		bv, ok := b.(SymbolExpr)
		return ok && av.Name == bv.Name
	case KeywordExpr:
		bv, ok := b.(KeywordExpr)
		return ok && av.Name == bv.Name
	case StringExpr:
		bv, ok := b.(StringExpr)
		return ok && av.Value == bv.Value
	case CharExpr:
		bv, ok := b.(CharExpr)
		return ok && av.Value == bv.Value
	case IntExpr:
		bv, ok := b.(IntExpr)
		return ok && av.Value == bv.Value
	case BigIntExpr:
		bv, ok := b.(BigIntExpr)
		return ok && av.Value == bv.Value
	case FloatExpr:
		bv, ok := b.(FloatExpr)
		return ok && av.Raw == bv.Raw && av.Value == bv.Value
	case RatioExpr:
		bv, ok := b.(RatioExpr)
		return ok && av.Numerator == bv.Numerator && av.Denominator == bv.Denominator
	case QuotedSymbolExpr:
		bv, ok := b.(QuotedSymbolExpr)
		return ok && av.Name == bv.Name
	case ListExpr:
		bv, ok := b.(ListExpr)
		if !ok || len(av.Elements) != len(bv.Elements) {
			return false
		}
		for i := range av.Elements {
			if !exprStructEqual(av.Elements[i], bv.Elements[i]) {
				return false
			}
		}
		return true
	case VectorExpr:
		bv, ok := b.(VectorExpr)
		if !ok || len(av.Elements) != len(bv.Elements) {
			return false
		}
		for i := range av.Elements {
			if !exprStructEqual(av.Elements[i], bv.Elements[i]) {
				return false
			}
		}
		return true
	case MapExpr:
		bv, ok := b.(MapExpr)
		if !ok || len(av.Entries) != len(bv.Entries) {
			return false
		}
		for i := range av.Entries {
			if !exprStructEqual(av.Entries[i], bv.Entries[i]) {
				return false
			}
		}
		return true
	case SetExpr:
		bv, ok := b.(SetExpr)
		if !ok || len(av.Elements) != len(bv.Elements) {
			return false
		}
		for i := range av.Elements {
			if !exprStructEqual(av.Elements[i], bv.Elements[i]) {
				return false
			}
		}
		return true
	case HashFnExpr:
		bv, ok := b.(HashFnExpr)
		return ok && exprStructEqual(av.Body, bv.Body)
	case MetaExpr:
		bv, ok := b.(MetaExpr)
		return ok && exprStructEqual(av.Meta, bv.Meta) && exprStructEqual(av.Target, bv.Target)
	default:
		return false
	}
}

func matchMacroSequence(patterns []Expr, target []Expr, bindings map[string]Expr, restBindings map[string][]Expr) (bool, error) {
	j := 0
	for i := 0; i < len(patterns); i++ {
		sym, ok := patterns[i].(SymbolExpr)
		if ok && sym.Name == "&" {
			if i != len(patterns)-2 {
				return false, fmt.Errorf("macro-case rest capture must be in the penultimate position")
			}
			name, ok := patterns[i+1].(SymbolExpr)
			if !ok || name.Name == "" || name.Name == "&" {
				return false, fmt.Errorf("macro-case rest capture expects a symbol name")
			}
			captured := make([]Expr, 0, len(target)-j)
			for _, expr := range target[j:] {
				captured = append(captured, copyExpr(expr))
			}
			restBindings[name.Name] = captured
			return true, nil
		}
		if j >= len(target) {
			return false, nil
		}
		pat := patterns[i]
		matched, err := matchMacroPattern(pat, []Expr{target[j]}, bindings, restBindings)
		if err != nil {
			return false, err
		}
		if !matched {
			return false, nil
		}
		j++
	}
	return j == len(target), nil
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
	case CharExpr:
		code := fmt.Sprintf("%s.NewString(%q)", runtimeAlias, string(arg.Value))
		return goExpr{code: ctx.constCode("Char", code), kind: exprKindValue}, nil
	case IntExpr:
		// Longs are scalar Values (no heap allocation), so hoisting adds clutter
		// without benefit; emit inline.
		return goExpr{code: fmt.Sprintf("%s.NewLong(%d)", runtimeAlias, arg.Value), kind: exprKindValue}, nil
	case BigIntExpr:
		code := fmt.Sprintf("%s.NewBigIntFromString(%q)", runtimeAlias, arg.Value)
		return goExpr{code: ctx.constCode("Big_"+sanitizeKeywordIdent(arg.Value), code), kind: exprKindValue}, nil
	case RatioExpr:
		code := fmt.Sprintf("%s.NewRatio(%d, %d)", runtimeAlias, arg.Numerator, arg.Denominator)
		return goExpr{code: ctx.constCode(fmt.Sprintf("Ratio_%d_%d", arg.Numerator, arg.Denominator), code), kind: exprKindValue}, nil
	case FloatExpr:
		if arg.Raw != "" {
			return goExpr{code: fmt.Sprintf("%s.NewDouble(%s)", runtimeAlias, arg.Raw), kind: exprKindValue}, nil
		}
		return goExpr{code: fmt.Sprintf("%s.NewDouble(%g)", runtimeAlias, arg.Value), kind: exprKindValue}, nil
	case KeywordExpr:
		return goExpr{code: ctx.keywordCode(arg.Name), kind: exprKindValue}, nil
	case QuotedSymbolExpr:
		code := fmt.Sprintf("%s.NewSymbol(%q)", runtimeAlias, arg.Name)
		return goExpr{code: ctx.constCode("Sym_"+sanitizeKeywordIdent(arg.Name), code), kind: exprKindValue}, nil
	case QuotedListExpr:
		return quotedListExprToGo(arg, ctx)
	case VectorExpr:
		return vectorExprToGo(arg.Elements, ctx, locals)
	case MapExpr:
		return mapExprToGo(arg.Entries, ctx, locals)
	case SetExpr:
		return setExprToGo(arg.Elements, ctx, locals)
	case HashFnExpr:
		return hashFnExprToGo(arg, ctx, locals)
	case MetaExpr:
		return exprToGo(arg.Target, ctx, locals)
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
		// Locals are always bare (params/let); never qualified.
		if !strings.Contains(arg.Name, "/") {
			if ident, err := toGoIdentifier(arg.Name); err == nil && locals != nil {
				if kind, ok := locals[ident]; ok {
					if kind == exprKindMutableValue {
						return goExpr{code: ident, kind: exprKindValue}, nil
					}
					return goExpr{code: ident, kind: kind}, nil
				}
			}
		}
		// Module table: bare locals, :refer names, and qualified imports.
		if goIdent, ok := ctx.resolveModuleSymbol(arg.Name); ok {
			if kind, ok := ctx.globals[goIdent]; ok {
				return goExpr{code: goIdent, kind: kind}, nil
			}
			if _, ok := ctx.functions[goIdent]; ok {
				return goExpr{code: fmt.Sprintf("%s.NewFunction(%s)", runtimeAlias, goIdent), kind: exprKindValue}, nil
			}
			return goExpr{code: goIdent, kind: exprKindValue}, nil
		}
		ident, err := toGoIdentifier(arg.Name)
		if err == nil {
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
		if bind, ok := goFnBindings[arg.Name]; ok {
			// Namespaced Go functions are bound statically at compile time to a
			// generated, reflection-free adapter (see internal/gobindgen); no
			// runtime name lookup occurs.
			return goExpr{code: fmt.Sprintf("%s.%s", runtimeAlias, bind), kind: exprKindValue}, nil
		}
		if strings.Contains(arg.Name, "/") {
			return goExpr{}, exprError(arg, fmt.Sprintf("unknown symbol %q", arg.Name))
		}
		if err != nil {
			return goExpr{}, err
		}
		return goExpr{}, exprError(arg, fmt.Sprintf("unknown symbol %q", arg.Name))
	case ListExpr:
		return listExprToGo(arg, ctx, locals)
	default:
		return goExpr{}, fmt.Errorf("unsupported literal")
	}
}

func quotedListExprToGo(arg QuotedListExpr, ctx compileContext) (goExpr, error) {
	parts := make([]string, 0, len(arg.Elements))
	for _, item := range arg.Elements {
		code, err := quotedLiteralToValueCode(item)
		if err != nil {
			return goExpr{}, err
		}
		parts = append(parts, code)
	}
	// A quoted list is built entirely from literals, so it is constant: hoist it.
	code := fmt.Sprintf("%s.NewList(%s)", runtimeAlias, strings.Join(parts, ", "))
	return goExpr{code: ctx.constCode("List", code), kind: exprKindValue}, nil
}

func quotedLiteralToValueCode(expr Expr) (string, error) {
	switch value := expr.(type) {
	case IntExpr:
		return fmt.Sprintf("%s.NewLong(%d)", runtimeAlias, value.Value), nil
	case BigIntExpr:
		return fmt.Sprintf("%s.NewBigIntFromString(%q)", runtimeAlias, value.Value), nil
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
	case "+", "-", "*", "/", "%", "=", "<", "<=", ">", ">=", "max", "min",
		"first", "fist", "second", "rest", "next", "last", "reverse", "cons", "take", "drop",
		"map", "concat", "group-by", "sort-by", "juxt", "partial", "apply", "max-key", "val", "zipmap", "pmap", "filter", "reduce", "range", "get", "get-in", "keys", "hash-map", "select-keys",
		"identity", "not-empty", "not-empty?", "empty?", "nil?", "not", "count", "double", "format", "keyword", "into",
		"remove", "doall", "dorun", "line-seq", "some", "some?", "seq", "seq?", "not-any?", "every?", "keep", "set", "vec", "conj", "contains?",
		"assoc", "merge", "update", "dissoc", "open-file", "file-to-strings", "rand-int", "constantly", "repeat",
		"go-fn", "go-fn-args", "re-pattern", "re-matches":
		return true
	default:
		return false
	}
}

func listExprToGo(list ListExpr, ctx compileContext, locals map[string]exprKind) (result goExpr, err error) {
	defer func() {
		if err != nil {
			err = annotateExprError(list, err)
		}
	}()
	if len(list.Elements) == 0 {
		return goExpr{}, fmt.Errorf("unsupported form")
	}

	head, ok := list.Elements[0].(SymbolExpr)
	if ok {
		if strings.HasPrefix(head.Name, ".") {
			return dotMethodExprToGo(head.Name, list.Elements[1:], ctx, locals)
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
		case "%":
			return modExprToGo(list.Elements[1:], ctx, locals)
		case "=":
			return equalityExprToGo(list.Elements[1:], ctx, locals)
		case "<":
			return comparisonExprToGo(list.Elements[1:], runtimeAlias+".Lt", "<", ctx, locals)
		case "<=":
			return comparisonExprToGo(list.Elements[1:], runtimeAlias+".Le", "<=", ctx, locals)
		case ">":
			return comparisonExprToGo(list.Elements[1:], runtimeAlias+".Gt", ">", ctx, locals)
		case ">=":
			return comparisonExprToGo(list.Elements[1:], runtimeAlias+".Ge", ">=", ctx, locals)
		case "max":
			return maxExprToGo(list.Elements[1:], ctx, locals)
		case "min":
			return minExprToGo(list.Elements[1:], ctx, locals)
		case "str":
			return strExprToGo(list.Elements[1:], ctx, locals)
		case "println":
			return printlnExprToGo(list.Elements[1:], ctx, locals)
		case "testing":
			return testingExprToGo(list.Elements[1:], ctx, locals)
		case "ex-info":
			return exInfoExprToGo(list.Elements[1:], ctx, locals)
		case "throw":
			return throwExprToGo(list.Elements[1:], ctx, locals)
		case "is":
			return isExprToGo(list, ctx, locals)
		case "expect-exception":
			return expectExceptionExprToGo(list, ctx, locals)
		case "if":
			return ifExprToGo(list.Elements[1:], ctx, locals)
		case "do":
			return doExprToGo(list.Elements[1:], ctx, locals)
		case "doto":
			return dotoExprToGo(list.Elements[1:], ctx, locals)
		case "for":
			return forExprToGo(list.Elements[1:], ctx, locals)
		case "doseq":
			return doseqExprToGo(list.Elements[1:], ctx, locals)
		case "let":
			return letExprToGo(list.Elements[1:], ctx, locals)
		case "loop":
			return loopExprToGo(list.Elements[1:], ctx, locals)
		case "recur":
			return recurExprToGo(list.Elements[1:], ctx, locals)
		case "with-open":
			return withOpenExprToGo(list.Elements[1:], ctx, locals)
		case "update!":
			return updateBangExprToGo(list.Elements[1:], ctx, locals)
		case "or":
			return orExprToGo(list.Elements[1:], ctx, locals)
		case "and":
			return andExprToGo(list.Elements[1:], ctx, locals)
		case "symbol":
			return symbolExprToGo(list.Elements[1:], ctx, locals)
		case "name":
			return nameExprToGo(list.Elements[1:], ctx, locals)
		case "keyword":
			return keywordExprToGo(list.Elements[1:], ctx, locals)
		case "first", "fist":
			return firstExprToGo(list.Elements[1:], ctx, locals)
		case "second":
			return secondExprToGo(list.Elements[1:], ctx, locals)
		case "rest":
			return restExprToGo(list.Elements[1:], ctx, locals)
		case "next":
			return nextExprToGo(list.Elements[1:], ctx, locals)
		case "last":
			return lastExprToGo(list.Elements[1:], ctx, locals)
		case "reverse":
			return reverseExprToGo(list.Elements[1:], ctx, locals)
		case "cons":
			return consExprToGo(list.Elements[1:], ctx, locals)
		case "take":
			return takeExprToGo(list.Elements[1:], ctx, locals)
		case "drop":
			return dropExprToGo(list.Elements[1:], ctx, locals)
		case "not-empty":
			return notEmptyExprToGo(list.Elements[1:], ctx, locals)
		case "seq":
			return seqExprToGo(list.Elements[1:], ctx, locals)
		case "not-empty?":
			return notEmptyPredicateExprToGo(list.Elements[1:], ctx, locals)
		case "empty?":
			return emptyPredicateExprToGo(list.Elements[1:], ctx, locals)
		case "nil?":
			return nilPredicateExprToGo(list.Elements[1:], ctx, locals)
		case "not":
			return notExprToGo(list.Elements[1:], ctx, locals)
		case "count":
			return countExprToGo(list.Elements[1:], ctx, locals)
		case "double":
			return doubleExprToGo(list.Elements[1:], ctx, locals)
		case "into":
			return intoExprToGo(list.Elements[1:], ctx, locals)
		case "format":
			return formatExprToGo(list.Elements[1:], ctx, locals)
		case "hash-map":
			return hashMapExprToGo(list.Elements[1:], ctx, locals)
		case "select-keys":
			return selectKeysExprToGo(list.Elements[1:], ctx, locals)
		case "map":
			return mapCallExprToGo(list.Elements[1:], ctx, locals)
		case "concat":
			return concatCallExprToGo(list.Elements[1:], ctx, locals)
		case "group-by":
			return groupByCallExprToGo(list.Elements[1:], ctx, locals)
		case "sort-by":
			return sortByCallExprToGo(list.Elements[1:], ctx, locals)
		case "juxt":
			return juxtCallExprToGo(list.Elements[1:], ctx, locals)
		case "partial":
			return partialCallExprToGo(list.Elements[1:], ctx, locals)
		case "apply":
			return applyCallExprToGo(list.Elements[1:], ctx, locals)
		case "max-key":
			return maxKeyCallExprToGo(list.Elements[1:], ctx, locals)
		case "val":
			return valExprToGo(list.Elements[1:], ctx, locals)
		case "zipmap":
			return zipmapCallExprToGo(list.Elements[1:], ctx, locals)
		case "pmap":
			return pmapCallExprToGo(list.Elements[1:], ctx, locals)
		case "filter":
			return filterCallExprToGo(list.Elements[1:], ctx, locals)
		case "reduce":
			return reduceCallExprToGo(list.Elements[1:], ctx, locals)
		case "remove":
			return removeCallExprToGo(list.Elements[1:], ctx, locals)
		case "doall":
			return doallExprToGo(list.Elements[1:], ctx, locals)
		case "some":
			return someCallExprToGo(list.Elements[1:], ctx, locals)
		case "some?":
			return somePredicateCallExprToGo(list.Elements[1:], ctx, locals)
		case "seq?":
			return seqPredicateExprToGo(list.Elements[1:], ctx, locals)
		case "not-any?":
			return notAnyCallExprToGo(list.Elements[1:], ctx, locals)
		case "every?":
			return everyCallExprToGo(list.Elements[1:], ctx, locals)
		case "keep":
			return keepCallExprToGo(list.Elements[1:], ctx, locals)
		case "set":
			return setCallExprToGo(list.Elements[1:], ctx, locals)
		case "vec":
			return vecCallExprToGo(list.Elements[1:], ctx, locals)
		case "conj":
			return conjExprToGo(list.Elements[1:], ctx, locals)
		case "contains?":
			return containsCallExprToGo(list.Elements[1:], ctx, locals)
		case "line-seq":
			return lineSeqExprToGo(list.Elements[1:], ctx, locals)
		case "range":
			return rangeCallExprToGo(list.Elements[1:], ctx, locals)
		case "repeat":
			return repeatCallExprToGo(list.Elements[1:], ctx, locals)
		case "rand-int":
			return randIntExprToGo(list.Elements[1:], ctx, locals)
		case "assoc":
			return assocExprToGo(list.Elements[1:], ctx, locals)
		case "merge":
			return mergeExprToGo(list.Elements[1:], ctx, locals)
		case "update":
			return updateExprToGo(list.Elements[1:], ctx, locals)
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

			valueCode, err := functionArgToValueCode(part, item, ctx)
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
		valueCode, err := functionArgToValueCode(part, item, ctx)
		if err != nil {
			return goExpr{}, err
		}
		args = append(args, valueCode)
	}
	return goExpr{code: fmt.Sprintf("%s.Call(%s)", runtimeAlias, strings.Join(append([]string{callee.code}, args...), ", ")), kind: exprKindValue}, nil
}

func dotMethodExprToGo(name string, args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	switch name {
	case ".write":
		if len(args) != 2 {
			return goExpr{}, fmt.Errorf(".write expects file and content")
		}
		fileExpr, err := exprToGo(args[0], ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if fileExpr.kind != exprKindValue {
			return goExpr{}, fmt.Errorf(".write expects a file Value")
		}
		contentExpr, err := exprToGo(args[1], ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		contentCode, err := stringishExprToStringCode(contentExpr)
		if err != nil {
			return goExpr{}, exprError(args[1], ".write content must be string-like")
		}
		return goExpr{code: fmt.Sprintf("%s.Write(%s, %s)", runtimeAlias, fileExpr.code, contentCode), kind: exprKindValue}, nil
	case ".endsWith":
		if len(args) != 2 {
			return goExpr{}, fmt.Errorf(".endsWith expects string and suffix")
		}
		left, err := exprToGo(args[0], ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		right, err := exprToGo(args[1], ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		leftCode, err := stringishExprToStringCode(left)
		if err != nil {
			return goExpr{}, exprError(args[0], ".endsWith expects a string-like value")
		}
		rightCode, err := stringishExprToStringCode(right)
		if err != nil {
			return goExpr{}, exprError(args[1], ".endsWith expects a string-like value")
		}
		return goExpr{code: fmt.Sprintf("strings.HasSuffix(%s, %s)", leftCode, rightCode), kind: exprKindBool}, nil
	default:
		return goExpr{}, fmt.Errorf("unsupported symbol %q", name)
	}
}

func stringishExprToStringCode(expr goExpr) (string, error) {
	switch expr.kind {
	case exprKindString:
		return expr.code, nil
	case exprKindValue:
		return fmt.Sprintf("%s.Str(%s)", runtimeAlias, expr.code), nil
	default:
		return "", fmt.Errorf("expected string-like value")
	}
}

func functionArgToValueCode(arg goExpr, source Expr, ctx compileContext) (string, error) {
	switch arg.kind {
	case exprKindValue:
		return arg.code, nil
	case exprKindString:
		if code, ok := ctx.stringLiteralValueCode(source, arg.code); ok {
			return code, nil
		}
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

func maxExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	return variadicNumericCallExprToGo("max", runtimeAlias+".Max", args, ctx, locals)
}

func minExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	return variadicNumericCallExprToGo("min", runtimeAlias+".Min", args, ctx, locals)
}

func variadicNumericCallExprToGo(name, goName string, args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) < 1 {
		return goExpr{}, fmt.Errorf("%s expects at least one argument", name)
	}
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("%s arguments must evaluate to Value", name)
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s(%s)", goName, strings.Join(parts, ", ")), kind: exprKindValue}, nil
}

func equalityExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) < 2 {
		return goExpr{}, fmt.Errorf("= expects at least two arguments")
	}

	parts := make([]string, 0, len(args))
	for _, item := range args {
		part, err := exprToGo(item, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		switch part.kind {
		case exprKindValue:
			parts = append(parts, part.code)
		case exprKindString:
			parts = append(parts, fmt.Sprintf("%s.NewString(%s)", runtimeAlias, part.code))
		default:
			return goExpr{}, exprError(item, "= arguments must evaluate to Value")
		}
	}

	checks := make([]string, 0, len(parts)-1)
	for i := 0; i < len(parts)-1; i++ {
		checks = append(checks, fmt.Sprintf("%s.Eq(%s, %s)", runtimeAlias, parts[i], parts[i+1]))
	}
	return goExpr{code: fmt.Sprintf("%s.NewBool(%s)", runtimeAlias, strings.Join(checks, " && ")), kind: exprKindValue}, nil
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
		switch part.kind {
		case exprKindValue:
			parts = append(parts, part.code)
		case exprKindString:
			parts = append(parts, fmt.Sprintf("%s.NewString(%s)", runtimeAlias, part.code))
		default:
			return goExpr{}, fmt.Errorf("%s expects numeric or string Value arguments", operator)
		}
	}

	checks := make([]string, 0, len(parts)-1)
	for i := 0; i < len(parts)-1; i++ {
		checks = append(checks, fmt.Sprintf("%s(%s, %s)", runtimeOp, parts[i], parts[i+1]))
	}
	return goExpr{code: strings.Join(checks, " && "), kind: exprKindBool}, nil
}

// goFormExprToGo lowers (go body...) to a goroutine that evaluates body for
// side effects and returns nil immediately to the caller.
func goFormExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) == 0 {
		return goExpr{}, fmt.Errorf("go expects at least one body expression")
	}
	body, err := doExprToGo(args, ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	body, err = coerceExprToValue(body, args[len(args)-1], "go body", ctx)
	if err != nil {
		return goExpr{}, err
	}
	// IIFE so the form is an expression; go launches async work and returns nil.
	code := fmt.Sprintf(
		"func() %s.Value {\n"+
			"\tgo func() {\n"+
			"\t\tdefer func() {\n"+
			"\t\t\tif r := recover(); r != nil {\n"+
			"\t\t\t\t%s.ReportGoPanic(r)\n"+
			"\t\t\t}\n"+
			"\t\t}()\n"+
			"\t\t_ = %s\n"+
			"\t}()\n"+
			"\treturn %s.NilValue()\n"+
			"}()",
		runtimeAlias, runtimeAlias, body.code, runtimeAlias,
	)
	return goExpr{code: code, kind: exprKindValue}, nil
}

// futureFormExprToGo lowers (future body...) to NewFuture(func() Value { body }).
// The result is a zero-arg function: call it to block for the value.
func futureFormExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) == 0 {
		return goExpr{}, fmt.Errorf("future expects at least one body expression")
	}
	body, err := doExprToGo(args, ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	body, err = coerceExprToValue(body, args[len(args)-1], "future body", ctx)
	if err != nil {
		return goExpr{}, err
	}
	code := fmt.Sprintf("%s.NewFuture(func() %s.Value { return %s })", runtimeAlias, runtimeAlias, body.code)
	return goExpr{code: code, kind: exprKindValue}, nil
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
			trueExpr, err = coerceExprToValue(trueExpr, args[1], "if true branch", ctx)
			if err != nil {
				return goExpr{}, err
			}
			falseExpr, err = coerceExprToValue(falseExpr, args[2], "if false branch", ctx)
			if err != nil {
				return goExpr{}, err
			}
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

func dotoExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) < 1 {
		return goExpr{}, fmt.Errorf("doto expects a target and at least one form")
	}
	target, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	target, err = coerceExprToValue(target, args[0], "doto target", ctx)
	if err != nil {
		return goExpr{}, err
	}

	var out strings.Builder
	fmt.Fprintf(&out, "func() %s.Value {\n", runtimeAlias)
	fmt.Fprintf(&out, "\t__doto := %s\n", target.code)
	for _, form := range args[1:] {
		step, err := exprToGo(form, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		stepCode, err := coerceExprToValue(step, form, "doto step", ctx)
		if err != nil {
			return goExpr{}, err
		}
		fmt.Fprintf(&out, "\t_ = %s.Call(%s, __doto)\n", runtimeAlias, stepCode.code)
	}
	fmt.Fprintf(&out, "\treturn __doto\n")
	out.WriteString("}()")

	return goExpr{code: out.String(), kind: exprKindValue}, nil
}

func exInfoExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 2 && len(args) != 3 {
		return goExpr{}, fmt.Errorf("ex-info expects message, data, and optional cause")
	}

	msg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	if msg.kind != exprKindString {
		return goExpr{}, fmt.Errorf("ex-info message must be a string")
	}

	data, err := exprToGo(args[1], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	dataCode, err := coerceExprToValue(data, args[1], "ex-info data", ctx)
	if err != nil {
		return goExpr{}, err
	}

	parts := []string{
		ctx.keywordCode("message"),
		fmt.Sprintf("%s.NewString(%s)", runtimeAlias, msg.code),
		ctx.keywordCode("data"),
		dataCode.code,
	}
	if len(args) == 3 {
		cause, err := exprToGo(args[2], ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		causeCode, err := coerceExprToValue(cause, args[2], "ex-info cause", ctx)
		if err != nil {
			return goExpr{}, err
		}
		parts = append(parts,
			ctx.keywordCode("cause"),
			causeCode.code,
		)
	}
	return goExpr{code: fmt.Sprintf("%s.NewMap(%s)", runtimeAlias, strings.Join(parts, ", ")), kind: exprKindValue}, nil
}

func throwExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("throw expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	valueCode, err := coerceExprToValue(arg, args[0], "throw", ctx)
	if err != nil {
		return goExpr{}, err
	}
	return goExpr{code: fmt.Sprintf("func() %s.Value {\n\tpanic(%s.ValueToString(%s))\n\treturn %s.NilValue()\n}()", runtimeAlias, runtimeAlias, valueCode.code, runtimeAlias), kind: exprKindValue}, nil
}

func forExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) < 2 {
		return goExpr{}, fmt.Errorf("for expects a binding vector and body")
	}

	bindingsExpr, ok := args[0].(VectorExpr)
	if !ok {
		return goExpr{}, fmt.Errorf("for expects a binding vector")
	}
	if len(bindingsExpr.Elements)%2 != 0 {
		return goExpr{}, fmt.Errorf("for binding vector expects name/value pairs")
	}

	return forBindingsToGo(bindingsExpr.Elements, args[1:], ctx, locals)
}

func doseqExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) < 2 {
		return goExpr{}, fmt.Errorf("doseq expects a binding vector and body")
	}

	bindingsExpr, ok := args[0].(VectorExpr)
	if !ok {
		return goExpr{}, fmt.Errorf("doseq expects a binding vector")
	}
	if len(bindingsExpr.Elements)%2 != 0 {
		return goExpr{}, fmt.Errorf("doseq binding vector expects name/value pairs")
	}

	loop, err := doseqBindingsToGo(bindingsExpr.Elements, args[1:], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}

	return goExpr{code: fmt.Sprintf("func() %s.Value {\n\t_ = %s.DoAll(%s)\n\treturn %s.NilValue()\n}()", runtimeAlias, runtimeAlias, loop.code, runtimeAlias), kind: exprKindValue}, nil
}

func forBindingsToGo(bindings []Expr, bodyExprs []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(bodyExprs) == 0 {
		return goExpr{}, fmt.Errorf("for expects at least one body form")
	}

	if len(bindings) == 0 {
		body, err := doExprToGo(bodyExprs, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		body, err = coerceExprToValue(body, bodyExprs[len(bodyExprs)-1], "for body", ctx)
		if err != nil {
			return goExpr{}, err
		}
		return goExpr{code: fmt.Sprintf("%s.NewArray(%s)", runtimeAlias, body.code), kind: exprKindValue}, nil
	}

	if len(bindings) < 2 {
		return goExpr{}, fmt.Errorf("for binding vector expects name/value pairs")
	}

	sym, ok := unwrapMetaExpr(bindings[0]).(SymbolExpr)
	if !ok {
		return goExpr{}, exprError(bindings[0], "for binding name must be a symbol")
	}
	ident, err := toGoIdentifier(sym.Name)
	if err != nil {
		return goExpr{}, err
	}

	collExpr, err := exprToGo(bindings[1], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	collCode, err := collectionArgToValueCode(collExpr)
	if err != nil {
		return goExpr{}, exprError(bindings[1], "for binding collection must evaluate to Value")
	}

	localKinds := make(map[string]exprKind, len(locals)+1)
	for name, kind := range locals {
		localKinds[name] = kind
	}
	localKinds[ident] = exprKindValue

	rest, err := forBindingsToGo(bindings[2:], bodyExprs, ctx, localKinds)
	if err != nil {
		return goExpr{}, err
	}

	var out strings.Builder
	fmt.Fprintf(&out, "func() %s.Value {\n", runtimeAlias)
	fmt.Fprintf(&out, "\treturn %s.MapCat(%s.NewFunction(func(args ...%s.Value) %s.Value {\n", runtimeAlias, runtimeAlias, runtimeAlias, runtimeAlias)
	fmt.Fprintf(&out, "\t\tif len(args) != 1 {\n")
	fmt.Fprintf(&out, "\t\t\tpanic(\"for binding expects exactly one value\")\n")
	fmt.Fprintf(&out, "\t\t}\n")
	fmt.Fprintf(&out, "\t\t%s := args[0]\n", ident)
	fmt.Fprintf(&out, "\t\treturn %s\n", rest.code)
	out.WriteString("\t}), ")
	out.WriteString(collCode)
	out.WriteString(")\n")
	out.WriteString("}()")

	return goExpr{code: out.String(), kind: exprKindValue}, nil
}

func doseqBindingsToGo(bindings []Expr, bodyExprs []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(bodyExprs) == 0 {
		return goExpr{}, fmt.Errorf("doseq expects at least one body form")
	}

	if len(bindings) == 0 {
		body, err := doExprToGo(bodyExprs, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if _, err := coerceExprToValue(body, bodyExprs[len(bodyExprs)-1], "doseq body", ctx); err != nil {
			return goExpr{}, err
		}
		return goExpr{code: fmt.Sprintf("func() %s.Value {\n\t_ = %s\n\treturn %s.NewArray()\n}()", runtimeAlias, body.code, runtimeAlias), kind: exprKindValue}, nil
	}

	if len(bindings) < 2 {
		return goExpr{}, fmt.Errorf("doseq binding vector expects name/value pairs")
	}

	sym, ok := unwrapMetaExpr(bindings[0]).(SymbolExpr)
	if !ok {
		return goExpr{}, exprError(bindings[0], "doseq binding name must be a symbol")
	}
	ident, err := toGoIdentifier(sym.Name)
	if err != nil {
		return goExpr{}, err
	}

	collExpr, err := exprToGo(bindings[1], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	collCode, err := collectionArgToValueCode(collExpr)
	if err != nil {
		return goExpr{}, exprError(bindings[1], "doseq binding collection must evaluate to Value")
	}

	localKinds := make(map[string]exprKind, len(locals)+1)
	for name, kind := range locals {
		localKinds[name] = kind
	}
	localKinds[ident] = exprKindValue

	rest, err := doseqBindingsToGo(bindings[2:], bodyExprs, ctx, localKinds)
	if err != nil {
		return goExpr{}, err
	}

	var out strings.Builder
	fmt.Fprintf(&out, "func() %s.Value {\n", runtimeAlias)
	fmt.Fprintf(&out, "\treturn %s.MapCat(%s.NewFunction(func(args ...%s.Value) %s.Value {\n", runtimeAlias, runtimeAlias, runtimeAlias, runtimeAlias)
	fmt.Fprintf(&out, "\t\tif len(args) != 1 {\n")
	fmt.Fprintf(&out, "\t\t\tpanic(\"doseq binding expects exactly one value\")\n")
	fmt.Fprintf(&out, "\t\t}\n")
	fmt.Fprintf(&out, "\t\t%s := args[0]\n", ident)
	fmt.Fprintf(&out, "\t\treturn %s\n", rest.code)
	out.WriteString("\t}), ")
	out.WriteString(collCode)
	out.WriteString(")\n")
	out.WriteString("}()")

	return goExpr{code: out.String(), kind: exprKindValue}, nil
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
		bindingPattern := bindingsExpr.Elements[i]
		valueExpr, err := exprToGo(bindingsExpr.Elements[i+1], ctx, localKinds)
		if err != nil {
			return goExpr{}, err
		}

		isVolatile, targetPattern, err := parseVolatileBindingPattern(bindingPattern)
		if err != nil {
			return goExpr{}, err
		}
		if isVolatile {
			sym, ok := targetPattern.(SymbolExpr)
			if !ok {
				return goExpr{}, exprError(bindingPattern, "volatile let binding currently supports symbol bindings only")
			}
			name, err := toGoIdentifier(sym.Name)
			if err != nil {
				return goExpr{}, err
			}
			if _, exists := declared[name]; exists {
				return goExpr{}, exprError(bindingPattern, fmt.Sprintf("duplicate binding %q", sym.Name))
			}
			declared[name] = struct{}{}
			localKinds[name] = exprKindMutableValue
			if valueExpr.kind != exprKindValue {
				return goExpr{}, exprError(bindingsExpr.Elements[i+1], "volatile let binding value must evaluate to Value")
			}
			sourceName := fmt.Sprintf("__bind%d", tempCounter)
			tempCounter++
			bindings = append(bindings, fmt.Sprintf("\tvar %s = %s\n", sourceName, valueExpr.code))
			bindings = append(bindings, fmt.Sprintf("\tvar %s = %s\n", name, sourceName))
			continue
		}

		if sym, ok := targetPattern.(SymbolExpr); ok {
			emitter := newDestructureEmitter(ctx, localKinds, &bindings, &tempCounter, declared)
			if err := emitter.bindSymbolWithKind(sym, valueExpr.code, valueExpr.kind); err != nil {
				return goExpr{}, exprError(bindingPattern, err.Error())
			}
			continue
		}

		if valueExpr.kind != exprKindValue {
			return goExpr{}, exprError(bindingsExpr.Elements[i+1], "let binding value must evaluate to Value")
		}

		sourceName := fmt.Sprintf("__bind%d", tempCounter)
		tempCounter++
		bindings = append(bindings, fmt.Sprintf("\tvar %s = %s\n", sourceName, valueExpr.code))
		emitter := newDestructureEmitter(ctx, localKinds, &bindings, &tempCounter, declared)
		if err := emitter.bind(targetPattern, sourceName); err != nil {
			return goExpr{}, err
		}
	}

	bodyExprs := args[1:]
	if len(bodyExprs) == 0 {
		if len(bindingsExpr.Elements) == 0 {
			return goExpr{}, fmt.Errorf("let without body requires at least one binding")
		}
		lastPattern := unwrapMetaExpr(bindingsExpr.Elements[len(bindingsExpr.Elements)-2])
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

func loopExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) < 1 {
		return goExpr{}, fmt.Errorf("loop expects a binding vector")
	}

	bindingsExpr, ok := args[0].(VectorExpr)
	if !ok {
		return goExpr{}, fmt.Errorf("loop expects a binding vector")
	}
	if len(bindingsExpr.Elements)%2 != 0 {
		return goExpr{}, fmt.Errorf("loop binding vector expects name/value pairs")
	}
	if len(args) < 2 {
		return goExpr{}, fmt.Errorf("loop expects at least one body form")
	}

	localKinds := make(map[string]exprKind, len(locals)+len(bindingsExpr.Elements)/2)
	for name, kind := range locals {
		localKinds[name] = kind
	}

	bindingNames := make([]string, 0, len(bindingsExpr.Elements)/2)
	initialValues := make([]string, 0, len(bindingsExpr.Elements)/2)
	declared := make(map[string]struct{}, len(bindingsExpr.Elements)/2)

	for i := 0; i < len(bindingsExpr.Elements); i += 2 {
		bindingSymbol, ok := unwrapMetaExpr(bindingsExpr.Elements[i]).(SymbolExpr)
		if !ok || bindingSymbol.Name == "" {
			return goExpr{}, exprError(bindingsExpr.Elements[i], "loop bindings must be symbols")
		}

		goName, err := toGoIdentifier(bindingSymbol.Name)
		if err != nil {
			return goExpr{}, err
		}
		if _, exists := declared[goName]; exists {
			return goExpr{}, exprError(bindingsExpr.Elements[i], fmt.Sprintf("duplicate loop binding %q", bindingSymbol.Name))
		}
		declared[goName] = struct{}{}

		valueExpr, err := exprToGo(bindingsExpr.Elements[i+1], ctx, localKinds)
		if err != nil {
			return goExpr{}, err
		}
		valueExpr, err = coerceExprToValue(valueExpr, bindingsExpr.Elements[i+1], "loop binding value", ctx)
		if err != nil {
			return goExpr{}, err
		}

		bindingNames = append(bindingNames, goName)
		initialValues = append(initialValues, valueExpr.code)
		localKinds[goName] = exprKindMutableValue
	}

	loopCtx := copyCompileContext(ctx)
	loopCtx.loopBindingNames = append([]string(nil), bindingNames...)

	bodyExpr, err := doExprToGo(args[1:], loopCtx, localKinds)
	if err != nil {
		return goExpr{}, err
	}
	bodyExpr, err = coerceExprToValue(bodyExpr, args[len(args)-1], "loop body", ctx)
	if err != nil {
		return goExpr{}, err
	}

	var out strings.Builder
	fmt.Fprintf(&out, "func() %s.Value {\n", runtimeAlias)
	for i := range bindingNames {
		fmt.Fprintf(&out, "\tvar %s = %s\n", bindingNames[i], initialValues[i])
	}
	out.WriteString("\tfor {\n")
	fmt.Fprintf(&out, "\t\t__loopResult := %s\n", bodyExpr.code)
	fmt.Fprintf(&out, "\t\tif __recurValues, __isRecur := %s.UnwrapRecur(__loopResult); __isRecur {\n", runtimeAlias)
	fmt.Fprintf(&out, "\t\t\tif len(__recurValues) != %d {\n", len(bindingNames))
	out.WriteString("\t\t\t\tpanic(\"internal error: recur arity mismatch\")\n")
	out.WriteString("\t\t\t}\n")
	for i, name := range bindingNames {
		fmt.Fprintf(&out, "\t\t\t%s = __recurValues[%d]\n", name, i)
	}
	out.WriteString("\t\t\tcontinue\n")
	out.WriteString("\t\t}\n")
	out.WriteString("\t\treturn __loopResult\n")
	out.WriteString("\t}\n")
	out.WriteString("}()")

	return goExpr{code: out.String(), kind: exprKindValue}, nil
}

func recurExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(ctx.loopBindingNames) == 0 {
		return goExpr{}, fmt.Errorf("recur can only be used within loop")
	}
	if len(args) != len(ctx.loopBindingNames) {
		return goExpr{}, fmt.Errorf("recur expects %d arguments", len(ctx.loopBindingNames))
	}

	values := make([]string, 0, len(args))
	for i, arg := range args {
		valueExpr, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		valueExpr, err = coerceExprToValue(valueExpr, arg, fmt.Sprintf("recur argument %d", i+1), ctx)
		if err != nil {
			return goExpr{}, err
		}
		values = append(values, valueExpr.code)
	}

	return goExpr{code: fmt.Sprintf("%s.NewRecur(%s)", runtimeAlias, strings.Join(values, ", ")), kind: exprKindValue}, nil
}

func unwrapMetaExpr(expr Expr) Expr {
	if meta, ok := expr.(MetaExpr); ok {
		return meta.Target
	}
	return expr
}

func parseVolatileBindingPattern(expr Expr) (bool, Expr, error) {
	meta, ok := expr.(MetaExpr)
	if !ok {
		return false, expr, nil
	}
	mapExpr, ok := meta.Meta.(MapExpr)
	if !ok {
		// Non-map metadata (e.g. ^int, ^long, ^float) is accepted as a no-op hint.
		return false, meta.Target, nil
	}
	if len(mapExpr.Entries)%2 != 0 {
		return false, meta.Target, exprError(meta.Meta, "metadata map expects key/value pairs")
	}
	volatile := false
	for i := 0; i < len(mapExpr.Entries); i += 2 {
		key, ok := mapExpr.Entries[i].(KeywordExpr)
		if !ok {
			continue
		}
		if key.Name != "volatile" {
			continue
		}
		switch value := mapExpr.Entries[i+1].(type) {
		case SymbolExpr:
			switch value.Name {
			case "true":
				volatile = true
			case "false":
				volatile = false
			default:
				return false, meta.Target, exprError(value, "metadata :volatile must be true or false")
			}
		default:
			return false, meta.Target, exprError(mapExpr.Entries[i+1], "metadata :volatile must be true or false")
		}
	}
	return volatile, meta.Target, nil
}

func withOpenExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) < 2 {
		return goExpr{}, fmt.Errorf("with-open expects a binding vector and body")
	}

	bindingsExpr, ok := args[0].(VectorExpr)
	if !ok {
		return goExpr{}, fmt.Errorf("with-open expects a binding vector")
	}
	if len(bindingsExpr.Elements)%2 != 0 {
		return goExpr{}, fmt.Errorf("with-open binding vector expects name/value pairs")
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
			return goExpr{}, fmt.Errorf("with-open binding value must evaluate to Value")
		}

		sourceName := fmt.Sprintf("__bind%d", tempCounter)
		tempCounter++
		bindings = append(bindings, fmt.Sprintf("\tvar %s = %s\n", sourceName, valueExpr.code))
		bindings = append(bindings, fmt.Sprintf("\tdefer %s.Close()\n", sourceName))

		emitter := newDestructureEmitter(ctx, localKinds, &bindings, &tempCounter, declared)
		if err := emitter.bind(bindingsExpr.Elements[i], sourceName); err != nil {
			return goExpr{}, err
		}
	}

	bodyExprs := args[1:]
	if len(bodyExprs) == 0 {
		return goExpr{}, fmt.Errorf("with-open expects at least one body form")
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
	case MetaExpr:
		return e.bind(p.Target, sourceCode)
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
	return e.bindSymbolWithKind(sym, sourceCode, exprKindValue)
}

func (e destructureEmitter) bindSymbolWithKind(sym SymbolExpr, sourceCode string, kind exprKind) error {
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
	e.locals[name] = kind
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
				entries, err := expandMapDestructureKeys(value, e.ctx)
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

		keyExpr, err := compileDestructureMapKeyToValue(key, e.ctx)
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

func expandMapDestructureKeys(expr Expr, ctx compileContext) ([]destructureKeyBinding, error) {
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
			keyExpr:     ctx.keywordCode(sym.Name),
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

func compileDestructureMapKeyToValue(expr Expr, ctx compileContext) (string, error) {
	switch value := expr.(type) {
	case KeywordExpr:
		return ctx.keywordCode(value.Name), nil
	case SymbolExpr:
		return fmt.Sprintf("%s.NewSymbol(%q)", runtimeAlias, value.Name), nil
	case QuotedSymbolExpr:
		return fmt.Sprintf("%s.NewSymbol(%q)", runtimeAlias, value.Name), nil
	case IntExpr:
		return fmt.Sprintf("%s.NewLong(%d)", runtimeAlias, value.Value), nil
	case BigIntExpr:
		return fmt.Sprintf("%s.NewBigIntFromString(%q)", runtimeAlias, value.Value), nil
	case FloatExpr:
		if value.Raw != "" {
			return fmt.Sprintf("%s.NewDouble(%s)", runtimeAlias, value.Raw), nil
		}
		return fmt.Sprintf("%s.NewDouble(%g)", runtimeAlias, value.Value), nil
	case StringExpr:
		return fmt.Sprintf("%s.NewString(%q)", runtimeAlias, value.Value), nil
	case CharExpr:
		return fmt.Sprintf("%s.NewString(%q)", runtimeAlias, string(value.Value)), nil
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

func keywordExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("keyword expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	switch arg.kind {
	case exprKindString, exprKindValue:
		return goExpr{code: fmt.Sprintf("%s.Keyword(%s)", runtimeAlias, arg.code), kind: exprKindValue}, nil
	default:
		return goExpr{}, fmt.Errorf("keyword expects a string, symbol, or keyword argument")
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

func secondExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("second expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	if arg.kind != exprKindValue {
		return goExpr{}, fmt.Errorf("second expects an argument that evaluates to Value")
	}
	return goExpr{code: fmt.Sprintf("%s.Second(%s)", runtimeAlias, arg.code), kind: exprKindValue}, nil
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

func seqPredicateExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("seq? expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	if arg.kind != exprKindValue {
		return goExpr{}, fmt.Errorf("seq? expects an argument that evaluates to Value")
	}
	return goExpr{code: fmt.Sprintf("%s.SeqPredicate(%s)", runtimeAlias, arg.code), kind: exprKindValue}, nil
}

func nextExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("next expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	if arg.kind != exprKindValue {
		return goExpr{}, fmt.Errorf("next expects an argument that evaluates to Value")
	}
	return goExpr{code: fmt.Sprintf("%s.Next(%s)", runtimeAlias, arg.code), kind: exprKindValue}, nil
}

func lastExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("last expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	if arg.kind != exprKindValue {
		return goExpr{}, fmt.Errorf("last expects an argument that evaluates to Value")
	}
	return goExpr{code: fmt.Sprintf("%s.Last(%s)", runtimeAlias, arg.code), kind: exprKindValue}, nil
}

func reverseExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("reverse expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	if arg.kind != exprKindValue {
		return goExpr{}, fmt.Errorf("reverse expects an argument that evaluates to Value")
	}
	return goExpr{code: fmt.Sprintf("%s.Reverse(%s)", runtimeAlias, arg.code), kind: exprKindValue}, nil
}

func consExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 2 {
		return goExpr{}, fmt.Errorf("cons expects exactly two arguments")
	}
	item, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	coll, err := exprToGo(args[1], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	if item.kind != exprKindValue || coll.kind != exprKindValue {
		return goExpr{}, fmt.Errorf("cons arguments must evaluate to Value")
	}
	return goExpr{code: fmt.Sprintf("%s.Cons(%s, %s)", runtimeAlias, item.code, coll.code), kind: exprKindValue}, nil
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
		parts := make([]string, 0, len(args))
		for _, arg := range args {
			parts = append(parts, exprToSourceString(arg))
		}
		return goExpr{}, fmt.Errorf("map expects function and at least one sequence, got (%s)", strings.Join(parts, " "))
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

func concatCallExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind == exprKindString {
			part = goExpr{code: fmt.Sprintf("%s.NewString(%s)", runtimeAlias, part.code), kind: exprKindValue}
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("concat arguments must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.Concat(%s)", runtimeAlias, strings.Join(parts, ", ")), kind: exprKindValue}, nil
}

func groupByCallExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 2 {
		return goExpr{}, fmt.Errorf("group-by expects function and one collection")
	}
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("group-by arguments must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.GroupBy(%s, %s)", runtimeAlias, parts[0], parts[1]), kind: exprKindValue}, nil
}

func sortByCallExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 2 && len(args) != 3 {
		return goExpr{}, fmt.Errorf("sort-by expects key function, optional comparator, and collection")
	}
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("sort-by arguments must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	if len(parts) == 2 {
		return goExpr{code: fmt.Sprintf("%s.SortBy(%s, %s)", runtimeAlias, parts[0], parts[1]), kind: exprKindValue}, nil
	}
	return goExpr{code: fmt.Sprintf("%s.SortBy(%s, %s, %s)", runtimeAlias, parts[0], parts[1], parts[2]), kind: exprKindValue}, nil
}

func juxtCallExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) < 1 {
		return goExpr{}, fmt.Errorf("juxt expects at least one function")
	}
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("juxt arguments must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.Juxt(%s)", runtimeAlias, strings.Join(parts, ", ")), kind: exprKindValue}, nil
}

func partialCallExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) < 1 {
		return goExpr{}, fmt.Errorf("partial expects function and optional bound arguments")
	}
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("partial arguments must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.Partial(%s)", runtimeAlias, strings.Join(parts, ", ")), kind: exprKindValue}, nil
}

func applyCallExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) < 2 {
		return goExpr{}, fmt.Errorf("apply expects function and at least one argument sequence")
	}
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("apply arguments must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.Apply(%s)", runtimeAlias, strings.Join(parts, ", ")), kind: exprKindValue}, nil
}

func maxKeyCallExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) < 2 {
		return goExpr{}, fmt.Errorf("max-key expects key function and at least one value")
	}
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("max-key arguments must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.MaxKey(%s)", runtimeAlias, strings.Join(parts, ", ")), kind: exprKindValue}, nil
}

func valExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("val expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	if arg.kind != exprKindValue {
		return goExpr{}, fmt.Errorf("val argument must evaluate to Value")
	}
	return goExpr{code: fmt.Sprintf("%s.Val(%s)", runtimeAlias, arg.code), kind: exprKindValue}, nil
}

func zipmapCallExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 2 {
		return goExpr{}, fmt.Errorf("zipmap expects exactly two sequence arguments")
	}
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("zipmap arguments must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.ZipMap(%s, %s)", runtimeAlias, parts[0], parts[1]), kind: exprKindValue}, nil
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

func removeCallExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 2 {
		return goExpr{}, fmt.Errorf("remove expects function and one sequence")
	}

	parts := make([]string, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("remove arguments must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.Remove(%s)", runtimeAlias, strings.Join(parts, ", ")), kind: exprKindValue}, nil
}

func doallExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("doall expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	if arg.kind != exprKindValue {
		return goExpr{}, fmt.Errorf("doall expects an argument that evaluates to Value")
	}
	return goExpr{code: fmt.Sprintf("%s.DoAll(%s)", runtimeAlias, arg.code), kind: exprKindValue}, nil
}

func someCallExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 2 {
		return goExpr{}, fmt.Errorf("some expects predicate and collection")
	}
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("some arguments must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.Some(%s)", runtimeAlias, strings.Join(parts, ", ")), kind: exprKindValue}, nil
}

func somePredicateCallExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("some? expects exactly one argument")
	}
	part, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	if part.kind != exprKindValue {
		return goExpr{}, fmt.Errorf("some? argument must evaluate to Value")
	}
	return goExpr{code: fmt.Sprintf("%s.SomePredicate(%s)", runtimeAlias, part.code), kind: exprKindValue}, nil
}

func notAnyCallExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 2 {
		return goExpr{}, fmt.Errorf("not-any? expects predicate and collection")
	}
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("not-any? arguments must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.NotAny(%s)", runtimeAlias, strings.Join(parts, ", ")), kind: exprKindValue}, nil
}

func everyCallExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 2 {
		return goExpr{}, fmt.Errorf("every? expects predicate and collection")
	}
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("every? arguments must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.Every(%s)", runtimeAlias, strings.Join(parts, ", ")), kind: exprKindValue}, nil
}

func keepCallExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 2 {
		return goExpr{}, fmt.Errorf("keep expects function and collection")
	}
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("keep arguments must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.Keep(%s)", runtimeAlias, strings.Join(parts, ", ")), kind: exprKindValue}, nil
}

func setCallExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("set expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	if arg.kind != exprKindValue {
		return goExpr{}, fmt.Errorf("set expects an argument that evaluates to Value")
	}
	return goExpr{code: fmt.Sprintf("%s.Set(%s)", runtimeAlias, arg.code), kind: exprKindValue}, nil
}

func vecCallExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("vec expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	argCode, err := collectionArgToValueCode(arg)
	if err != nil {
		return goExpr{}, fmt.Errorf("vec expects an argument that evaluates to Value")
	}
	return goExpr{code: fmt.Sprintf("%s.Vec(%s)", runtimeAlias, argCode), kind: exprKindValue}, nil
}

func conjExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) < 2 {
		return goExpr{}, fmt.Errorf("conj expects collection and at least one item")
	}
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind == exprKindString {
			part = goExpr{code: fmt.Sprintf("%s.NewString(%s)", runtimeAlias, part.code), kind: exprKindValue}
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("conj arguments must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.Conj(%s)", runtimeAlias, strings.Join(parts, ", ")), kind: exprKindValue}, nil
}

func notEmptyExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("not-empty expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	if arg.kind != exprKindValue {
		return goExpr{}, fmt.Errorf("not-empty expects an argument that evaluates to Value")
	}
	return goExpr{code: fmt.Sprintf("%s.NotEmpty(%s)", runtimeAlias, arg.code), kind: exprKindValue}, nil
}

func seqExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("seq expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	argCode, err := collectionArgToValueCode(arg)
	if err != nil {
		return goExpr{}, fmt.Errorf("seq expects an argument that evaluates to Value")
	}
	return goExpr{code: fmt.Sprintf("%s.Seq(%s)", runtimeAlias, argCode), kind: exprKindValue}, nil
}

func collectionArgToValueCode(arg goExpr) (string, error) {
	switch arg.kind {
	case exprKindValue:
		return arg.code, nil
	case exprKindString:
		return fmt.Sprintf("%s.NewString(%s)", runtimeAlias, arg.code), nil
	default:
		return "", fmt.Errorf("argument must evaluate to Value")
	}
}

func notEmptyPredicateExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("not-empty? expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	argCode, err := collectionArgToValueCode(arg)
	if err != nil {
		return goExpr{}, fmt.Errorf("not-empty? expects an argument that evaluates to Value")
	}
	return goExpr{code: fmt.Sprintf("%s.NewBool(%s.IsNotEmpty(%s))", runtimeAlias, runtimeAlias, argCode), kind: exprKindValue}, nil
}

func emptyPredicateExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("empty? expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	argCode, err := collectionArgToValueCode(arg)
	if err != nil {
		return goExpr{}, fmt.Errorf("empty? expects an argument that evaluates to Value")
	}
	return goExpr{code: fmt.Sprintf("%s.NewBool(%s.IsEmpty(%s))", runtimeAlias, runtimeAlias, argCode), kind: exprKindValue}, nil
}

func nilPredicateExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("nil? expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	argCode, err := collectionArgToValueCode(arg)
	if err != nil {
		return goExpr{}, fmt.Errorf("nil? expects an argument that evaluates to Value")
	}
	return goExpr{code: fmt.Sprintf("%s.NewBool(%s.IsNil(%s))", runtimeAlias, runtimeAlias, argCode), kind: exprKindValue}, nil
}

func notExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("not expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	argCode, err := coerceExprToValue(arg, args[0], "not", ctx)
	if err != nil {
		return goExpr{}, err
	}
	return goExpr{code: fmt.Sprintf("%s.NewBool(!%s.IsTruthy(%s))", runtimeAlias, runtimeAlias, argCode.code), kind: exprKindValue}, nil
}

func countExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("count expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	// Handle string literals specially
	if arg.kind == exprKindString {
		return goExpr{code: fmt.Sprintf("%s.NewLong(int64(%s.Count(%s.NewString(%s))))", runtimeAlias, runtimeAlias, runtimeAlias, arg.code), kind: exprKindValue}, nil
	}
	if arg.kind != exprKindValue {
		return goExpr{}, fmt.Errorf("count expects an argument that evaluates to Value")
	}
	return goExpr{code: fmt.Sprintf("%s.NewLong(int64(%s.Count(%s)))", runtimeAlias, runtimeAlias, arg.code), kind: exprKindValue}, nil
}

func doubleExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("double expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	if arg.kind != exprKindValue {
		return goExpr{}, fmt.Errorf("double expects an argument that evaluates to Value")
	}
	return goExpr{code: fmt.Sprintf("%s.Double(%s)", runtimeAlias, arg.code), kind: exprKindValue}, nil
}

func intoExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 2 {
		return goExpr{}, fmt.Errorf("into expects exactly two arguments")
	}
	target, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	if target.kind != exprKindValue {
		return goExpr{}, fmt.Errorf("into expects a collection Value as the first argument")
	}
	from, err := exprToGo(args[1], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	if from.kind != exprKindValue {
		return goExpr{}, fmt.Errorf("into expects a collection Value as the second argument")
	}
	return goExpr{code: fmt.Sprintf("%s.Into(%s, %s)", runtimeAlias, target.code, from.code), kind: exprKindValue}, nil
}

func formatExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) < 1 {
		return goExpr{}, fmt.Errorf("format expects a format string and optional values")
	}
	formatExpr, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	if formatExpr.kind != exprKindString {
		return goExpr{}, fmt.Errorf("format expects a string format argument")
	}
	parts := make([]string, 0, len(args)-1)
	for _, arg := range args[1:] {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		valueCode, err := functionArgToValueCode(part, arg, ctx)
		if err != nil {
			return goExpr{}, fmt.Errorf("format arguments must evaluate to Value")
		}
		parts = append(parts, valueCode)
	}
	return goExpr{code: fmt.Sprintf("%s.NewString(%s.Format(%s, %s))", runtimeAlias, runtimeAlias, formatExpr.code, strings.Join(parts, ", ")), kind: exprKindValue}, nil
}

func hashMapExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args)%2 != 0 {
		return goExpr{}, fmt.Errorf("hash-map expects key/value pairs")
	}
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind == exprKindString {
			part = goExpr{code: fmt.Sprintf("%s.NewString(%s)", runtimeAlias, part.code), kind: exprKindValue}
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("hash-map entries must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.NewMap(%s)", runtimeAlias, strings.Join(parts, ", ")), kind: exprKindValue}, nil
}

func updateBangExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 2 {
		return goExpr{}, fmt.Errorf("update! expects a mutable symbol and replacement value")
	}
	sym, ok := args[0].(SymbolExpr)
	if !ok {
		return goExpr{}, exprError(args[0], "update! first argument must be a mutable symbol")
	}
	ident, err := toGoIdentifier(sym.Name)
	if err != nil {
		return goExpr{}, err
	}
	if locals == nil {
		return goExpr{}, exprError(args[0], "update! first argument must be a mutable let binding")
	}
	if locals[ident] != exprKindMutableValue {
		return goExpr{}, exprError(args[0], "update! first argument must be a mutable let binding")
	}
	replacement, err := exprToGo(args[1], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	replacement, err = coerceExprToValue(replacement, args[1], "update! replacement", ctx)
	if err != nil {
		return goExpr{}, err
	}
	return goExpr{
		code: fmt.Sprintf("func() %s.Value {\n\t%s = %s\n\treturn %s\n}()", runtimeAlias, ident, replacement.code, ident),
		kind: exprKindValue,
	}, nil
}

func containsCallExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 2 {
		return goExpr{}, fmt.Errorf("contains? expects collection and key")
	}
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("contains? arguments must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.NewBool(%s.Contains(%s))", runtimeAlias, runtimeAlias, strings.Join(parts, ", ")), kind: exprKindValue}, nil
}

func lineSeqExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("line-seq expects exactly one argument")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	if arg.kind != exprKindValue {
		return goExpr{}, fmt.Errorf("line-seq expects an argument that evaluates to Value")
	}
	return goExpr{code: fmt.Sprintf("%s.LineSeq(%s)", runtimeAlias, arg.code), kind: exprKindValue}, nil
}

func orExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) == 0 {
		return goExpr{code: runtimeAlias + ".NilValue()", kind: exprKindValue}, nil
	}
	return shortCircuitExprToGo(args, ctx, locals, true)
}

func andExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) == 0 {
		return goExpr{code: runtimeAlias + ".NewBool(true)", kind: exprKindValue}, nil
	}
	return shortCircuitExprToGo(args, ctx, locals, false)
}

func shortCircuitExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind, isOr bool) (goExpr, error) {
	compiled := make([]goExpr, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		compiled = append(compiled, part)
	}
	result := compiled[len(compiled)-1]

	var out strings.Builder
	out.WriteString("func() ")
	out.WriteString(runtimeAlias)
	out.WriteString(".Value {\n")
	for i := 0; i < len(compiled)-1; i++ {
		cond, err := truthyExprToGo(compiled[i])
		if err != nil {
			return goExpr{}, err
		}
		valueCode := compiled[i].code
		if compiled[i].kind == exprKindBool {
			valueCode = fmt.Sprintf("%s.NewBool(%s)", runtimeAlias, valueCode)
		}
		if isOr {
			fmt.Fprintf(&out, "\tif %s {\n", cond)
			fmt.Fprintf(&out, "\t\treturn %s\n", valueCode)
			out.WriteString("\t}\n")
		} else {
			fmt.Fprintf(&out, "\tif !(%s) {\n", cond)
			fmt.Fprintf(&out, "\t\treturn %s\n", valueCode)
			out.WriteString("\t}\n")
		}
	}

	switch result.kind {
	case exprKindBool:
		fmt.Fprintf(&out, "\treturn %s.NewBool(%s)\n", runtimeAlias, result.code)
	case exprKindValue:
		fmt.Fprintf(&out, "\treturn %s\n", result.code)
	case exprKindString:
		fmt.Fprintf(&out, "\treturn %s.NewString(%s)\n", runtimeAlias, result.code)
	default:
		return goExpr{}, fmt.Errorf("or/and expects comparable expressions")
	}
	out.WriteString("}()")
	return goExpr{code: out.String(), kind: exprKindValue}, nil
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

func repeatCallExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 && len(args) != 2 {
		return goExpr{}, fmt.Errorf("repeat expects one or two arguments")
	}
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind == exprKindString {
			part = goExpr{code: fmt.Sprintf("%s.NewString(%s)", runtimeAlias, part.code), kind: exprKindValue}
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("repeat arguments must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.Repeat(%s)", runtimeAlias, strings.Join(parts, ", ")), kind: exprKindValue}, nil
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

func sleepExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("sleep expects exactly one argument (milliseconds)")
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	arg, err = coerceExprToValue(arg, args[0], "sleep argument", ctx)
	if err != nil {
		return goExpr{}, err
	}
	return goExpr{code: fmt.Sprintf("%s.Sleep(%s)", runtimeAlias, arg.code), kind: exprKindValue}, nil
}

func makeChannelExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) > 1 {
		return goExpr{}, fmt.Errorf("make-channel expects 0 or 1 arguments")
	}
	if len(args) == 0 {
		return goExpr{code: fmt.Sprintf("%s.MakeChannel()", runtimeAlias), kind: exprKindValue}, nil
	}
	arg, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	arg, err = coerceExprToValue(arg, args[0], "make-channel capacity", ctx)
	if err != nil {
		return goExpr{}, err
	}
	return goExpr{code: fmt.Sprintf("%s.MakeChannel(%s)", runtimeAlias, arg.code), kind: exprKindValue}, nil
}

func channelSendExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 2 {
		return goExpr{}, fmt.Errorf("channel-send expects channel and value")
	}
	ch, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	ch, err = coerceExprToValue(ch, args[0], "channel-send channel", ctx)
	if err != nil {
		return goExpr{}, err
	}
	val, err := exprToGo(args[1], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	val, err = coerceExprToValue(val, args[1], "channel-send value", ctx)
	if err != nil {
		return goExpr{}, err
	}
	return goExpr{code: fmt.Sprintf("%s.ChannelSend(%s, %s)", runtimeAlias, ch.code, val.code), kind: exprKindValue}, nil
}

func channelReceiveExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 1 {
		return goExpr{}, fmt.Errorf("channel-receive expects a channel")
	}
	ch, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}
	ch, err = coerceExprToValue(ch, args[0], "channel-receive channel", ctx)
	if err != nil {
		return goExpr{}, err
	}
	return goExpr{code: fmt.Sprintf("%s.ChannelReceive(%s)", runtimeAlias, ch.code), kind: exprKindValue}, nil
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
	case MetaExpr:
		collectHashFnPlaceholders(value.Meta, maxParam)
		collectHashFnPlaceholders(value.Target, maxParam)
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
	case MetaExpr:
		return MetaExpr{
			Meta:   replaceHashFnPlaceholders(value.Meta),
			Target: replaceHashFnPlaceholders(value.Target),
			Line:   value.Line,
			Col:    value.Col,
		}
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
	lambdaCtx := copyCompileContext(ctx)
	lambdaCtx.loopBindingNames = nil

	params, localKinds, localInits, hasRest, err := bindLambdaParams(paramsExpr, lambdaCtx, locals, label)
	if err != nil {
		return goExpr{}, err
	}

	body, err := exprToGo(bodyExpr, lambdaCtx, localKinds)
	if err != nil {
		return goExpr{}, err
	}
	body, err = coerceExprToValue(body, bodyExpr, fmt.Sprintf("%s body", label), lambdaCtx)
	if err != nil {
		return goExpr{}, err
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

func coerceExprToValue(expr goExpr, source Expr, label string, ctx compileContext) (goExpr, error) {
	switch expr.kind {
	case exprKindValue:
		return expr, nil
	case exprKindBool:
		return goExpr{code: fmt.Sprintf("%s.NewBool(%s)", runtimeAlias, expr.code), kind: exprKindValue}, nil
	case exprKindString:
		if code, ok := ctx.stringLiteralValueCode(source, expr.code); ok {
			return goExpr{code: code, kind: exprKindValue}, nil
		}
		return goExpr{code: fmt.Sprintf("%s.NewString(%s)", runtimeAlias, expr.code), kind: exprKindValue}, nil
	default:
		return goExpr{}, exprError(source, fmt.Sprintf("%s must evaluate to Value", label))
	}
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
		paramExpr := unwrapMetaExpr(paramsExpr.Elements[idx])
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
			localInits = append(localInits, fmt.Sprintf("\t_ = %s\n", restName))
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
	code := fmt.Sprintf("%s.NewMap(%s)", runtimeAlias, strings.Join(parts, ", "))
	if allConstExprs(entries) {
		code = ctx.constCode("Map", code)
	}
	return goExpr{code: code, kind: exprKindValue}, nil
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
	code := fmt.Sprintf("%s.NewArray(%s)", runtimeAlias, strings.Join(parts, ", "))
	if allConstExprs(elements) {
		code = ctx.constCode("Vec", code)
	}
	return goExpr{code: code, kind: exprKindValue}, nil
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
	code := fmt.Sprintf("%s.NewSet(%s)", runtimeAlias, strings.Join(parts, ", "))
	if allConstExprs(elements) {
		code = ctx.constCode("Set", code)
	}
	return goExpr{code: code, kind: exprKindValue}, nil
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

func mergeExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("merge arguments must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.Merge(%s)", runtimeAlias, strings.Join(parts, ", ")), kind: exprKindValue}, nil
}

func selectKeysExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) != 2 {
		return goExpr{}, fmt.Errorf("select-keys expects map and key sequence")
	}
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("select-keys arguments must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.SelectKeys(%s, %s)", runtimeAlias, parts[0], parts[1]), kind: exprKindValue}, nil
}

func updateExprToGo(args []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	if len(args) < 3 {
		return goExpr{}, fmt.Errorf("update expects collection, key, function, and optional args")
	}
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		part, err := exprToGo(arg, ctx, locals)
		if err != nil {
			return goExpr{}, err
		}
		if part.kind != exprKindValue {
			return goExpr{}, fmt.Errorf("update arguments must evaluate to Value")
		}
		parts = append(parts, part.code)
	}
	return goExpr{code: fmt.Sprintf("%s.Update(%s)", runtimeAlias, strings.Join(parts, ", ")), kind: exprKindValue}, nil
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

func expectExceptionExprToGo(form ListExpr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	args := form.Elements[1:]
	if len(args) != 1 && len(args) != 2 {
		return goExpr{}, fmt.Errorf("expect-exception expects an expression and optional message")
	}

	body, err := exprToGo(args[0], ctx, locals)
	if err != nil {
		return goExpr{}, err
	}

	expected := ""
	if len(args) == 2 {
		msg, ok := args[1].(StringExpr)
		if !ok {
			return goExpr{}, fmt.Errorf("expect-exception message must be a string")
		}
		expected = msg.Value
	}

	source := exprToSourceString(args[0])
	prefix := ""
	if line, col, ok := exprPos(form); ok {
		prefix = fmt.Sprintf("at %d:%d: ", line, col)
	}
	noPanic := prefix + "expected exception from " + source

	var out strings.Builder
	fmt.Fprintf(&out, "func() %s.Value {\n", runtimeAlias)
	out.WriteString("\tdefer func() {\n")
	out.WriteString("\t\tr := recover()\n")
	out.WriteString("\t\tif r == nil {\n")
	fmt.Fprintf(&out, "\t\t\tpanic(%q)\n", noPanic)
	out.WriteString("\t\t}\n")
	if expected != "" {
		out.WriteString("\t\tmsg := fmt.Sprint(r)\n")
		fmt.Fprintf(&out, "\t\tif !strings.Contains(msg, %q) {\n", expected)
		fmt.Fprintf(&out, "\t\t\tpanic(%q + msg)\n", prefix+"expected exception matching "+strconv.Quote(expected)+", got ")
		out.WriteString("\t\t}\n")
	}
	out.WriteString("\t}()\n")
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
	case exprKindValue, exprKindMutableValue:
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
			if ch == '-' {
				out.WriteByte('_')
				continue
			}
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
		if ch == '?' {
			out.WriteString("_q")
			continue
		}
		if ch == '!' {
			out.WriteString("_bang")
			continue
		}
		if ch == '>' {
			out.WriteString("_gt")
			continue
		}
		if ch == '<' {
			out.WriteString("_lt")
			continue
		}
		if ch != '_' && !unicode.IsLetter(ch) && !unicode.IsDigit(ch) {
			return "", fmt.Errorf("unsupported symbol %q", name)
		}
		out.WriteRune(ch)
	}
	nameOut := out.String()
	if nameOut == "main" {
		nameOut = "flag_main"
	}
	if token.IsKeyword(nameOut) {
		nameOut += "_"
	}
	return nameOut, nil
}

func renderFunctionDef(fn functionDef) string {
	// Handle go-interface wrappers with custom Go signatures
	if fn.goSignature != "" {
		return fmt.Sprintf("func %s%s {\n%s}\n", fn.goName, fn.goSignature, fn.body)
	}
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
		if param == "_" {
			fmt.Fprintf(&body, "\t_ = args[%d]\n", index)
			continue
		}
		fmt.Fprintf(&body, "\t%s := args[%d]\n", param, index)
	}
	for _, init := range fn.localInits {
		body.WriteString(init)
		trimmed := strings.TrimSpace(init)
		if strings.HasPrefix(trimmed, "var ") {
			fields := strings.Fields(trimmed)
			if len(fields) >= 2 {
				fmt.Fprintf(&body, "\t_ = %s\n", fields[1])
			}
		}
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
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "var ") {
			fields := strings.Fields(trimmed)
			if len(fields) >= 2 {
				fmt.Fprintf(&out, "\t_ = %s\n", fields[1])
			}
		}
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

func compileGoInterface(form ListExpr, ctx compileContext, defined map[string]string) (functionDef, error) {
	if len(form.Elements) != 4 {
		return functionDef{}, exprError(form, "go-interface expects: function-name return-type [param-type ...]")
	}

	// Extract target function name
	targetFuncExpr, ok := form.Elements[1].(SymbolExpr)
	if !ok {
		return functionDef{}, exprError(form.Elements[1], "go-interface expects a function name")
	}
	targetFuncName := targetFuncExpr.Name

	// Extract return type
	returnTypeExpr, ok := form.Elements[2].(SymbolExpr)
	if !ok {
		return functionDef{}, exprError(form.Elements[2], "go-interface expects a return type symbol")
	}
	returnGoType := typeHintToGoType(returnTypeExpr.Name)
	if returnGoType == "" {
		return functionDef{}, exprError(returnTypeExpr, fmt.Sprintf("unsupported return type: %s", returnTypeExpr.Name))
	}

	// Look up the target function
	targetGoName, ok := defined[targetFuncName]
	if !ok {
		return functionDef{}, exprError(form, fmt.Sprintf("function %q not defined", targetFuncName))
	}

	// Extract parameter types from vector
	paramsVec, ok := form.Elements[3].(VectorExpr)
	if !ok {
		return functionDef{}, exprError(form.Elements[3], "go-interface expects a parameter type vector")
	}

	paramTypes := make([]string, 0, len(paramsVec.Elements))
	paramGoTypes := make([]string, 0, len(paramsVec.Elements))
	for i, paramExpr := range paramsVec.Elements {
		typeExpr, ok := paramExpr.(SymbolExpr)
		if !ok {
			return functionDef{}, exprError(paramExpr, "parameter type must be a symbol")
		}
		goType := typeHintToGoType(typeExpr.Name)
		if goType == "" {
			return functionDef{}, exprError(paramExpr, fmt.Sprintf("unsupported parameter type: %s", typeExpr.Name))
		}
		paramGoTypes = append(paramGoTypes, goType)
		paramTypes = append(paramTypes, fmt.Sprintf("p%d %s", i+1, goType))
	}

	// Generate wrapper function body
	// Box parameters
	var bodyBuf bytes.Buffer
	bodyBuf.WriteString("args := make([]flagrt.Value, 0, ")
	fmt.Fprintf(&bodyBuf, "%d)\n", len(paramGoTypes))
	for i, goType := range paramGoTypes {
		boxCode := boxGoValue(fmt.Sprintf("p%d", i+1), goType)
		fmt.Fprintf(&bodyBuf, "\targs = append(args, %s)\n", boxCode)
	}

	// Call the target FLAG function
	fmt.Fprintf(&bodyBuf, "\tresult := flagrt.Call(%s, args...)\n", targetGoName)

	// Unbox result
	unboxCode := unboxGoValue("result", returnGoType)
	fmt.Fprintf(&bodyBuf, "\treturn %s\n", unboxCode)

	// Create wrapper function name (unmangled)
	wrapperName := fmt.Sprintf("%s_go", targetFuncName)
	wrapperGoName := wrapperName // Keep it unmangled for Go interop

	// Return type for the function signature
	returnTypeDecl := returnGoType
	if returnGoType == "error" {
		returnTypeDecl = "error"
	}

	return functionDef{
		flagName:     wrapperName,
		goName:       wrapperGoName,
		variadicName: "",
		arityName:    "",
		hasRest:      false,
		doc:          fmt.Sprintf("Go interop wrapper for %s", targetFuncName),
		params:       nil,
		localInits:   nil,
		body:         bodyBuf.String(),
		// Store the Go signature for special rendering
		goSignature: fmt.Sprintf("(%s) %s", strings.Join(paramTypes, ", "), returnTypeDecl),
	}, nil
}

func typeHintToGoType(hint string) string {
	switch hint {
	case "double", "float":
		return "float64"
	case "long", "int":
		return "int64"
	case "string":
		return "string"
	case "boolean", "bool":
		return "bool"
	default:
		return ""
	}
}

func boxGoValue(varName string, goType string) string {
	switch goType {
	case "float64":
		return fmt.Sprintf("flagrt.NewDouble(%s)", varName)
	case "int64":
		return fmt.Sprintf("flagrt.NewLong(%s)", varName)
	case "string":
		return fmt.Sprintf("flagrt.NewString(%s)", varName)
	case "bool":
		return fmt.Sprintf("flagrt.NewBoolean(%s)", varName)
	default:
		return varName
	}
}

func unboxGoValue(varName string, goType string) string {
	switch goType {
	case "float64":
		return fmt.Sprintf("%s.Double()", varName)
	case "int64":
		return fmt.Sprintf("%s.Long()", varName)
	case "string":
		return fmt.Sprintf("%s.String()", varName)
	case "bool":
		return fmt.Sprintf("%s.Bool()", varName)
	default:
		return varName
	}
}
