package compiler

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/token"
	"path/filepath"
	"sort"
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
	exprKindDefer
)

type goExpr struct {
	code  string
	kind  exprKind
	ir    IRExpr
	stmt  IRStmt
	stmts []IRStmt
}

func exprPos(expr Expr) (int, int, bool) {
	switch value := expr.(type) {
	case ListExpr:
		return value.Line, value.Col, value.Line > 0
	case VectorExpr:
		return value.Line, value.Col, value.Line > 0
	case PipeVectorExpr:
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
	case PipeVectorExpr:
		parts := make([]string, 0, len(value.Elements))
		for _, item := range value.Elements {
			parts = append(parts, exprToSourceString(item))
		}
		if len(parts) == 0 {
			return "| |"
		}
		return "| " + strings.Join(parts, " ") + " |"
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
	prelude string
	code    string
	kind    exprKind // 0 means a statement (println/print); otherwise an expression
}

type testCase struct {
	goName     string
	label      string
	line       int
	bodySource string
}

type arityDef struct {
	params      []string
	localInits  []string
	body        string
	bodyPrelude string
	hasRest     bool
	arityName   string
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
	bodyPrelude  string
	goSignature  string // custom Go signature for interop wrappers (empty = use standard)
	arities      []arityDef
	irDecls      []IRStmt // FLAG-assembled top-level func decls; when set, renderFunctionDef uses these
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
	macroForms        []Expr
	prologueFns       []functionDef
	prologueVars      []varDef
	selfFunctionName  string // FLAG source name for self-recursion matching
	selfFunctionArity int
	selfArityName     string
	selfVariadicName  string
	selfFunctionRest  bool
	selfArityNames    map[int]string // fixed arity count -> Go arity function
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
	// ifTemps is a shared counter for result temps (if_result_N, let_result_N).
	ifTemps *int
	// packageMode emits each FLAG module as its own Go package. Exported
	// names are capitalized; imports bind to pkg.Name selectors.
	packageMode bool
	// exportedNames is the current module :exports set (packageMode only).
	exportedNames map[string]bool
	// goImports are FLAG-module packages this file must import (packageMode).
	goImports []goImport
}

type goImport struct {
	Name string
	Path string
}

// bindModuleName records a FLAG name -> Go ident mapping for this module.
func (ctx *compileContext) bindModuleName(flagName, goName string) error {
	if ctx.moduleSymbols == nil {
		return nil
	}
	if prev, ok := ctx.moduleSymbols[flagName]; ok && prev != goName {
		if !localDefReplacesImport(ctx, prev, goName) {
			return fmt.Errorf("symbol %q already bound to %s", flagName, prev)
		}
	}
	ctx.moduleSymbols[flagName] = goName
	if ctx.namespace != "" && !strings.Contains(flagName, "/") {
		qualified := ctx.namespace + "/" + flagName
		if prev, ok := ctx.moduleSymbols[qualified]; ok && prev != goName {
			if !localDefReplacesImport(ctx, prev, goName) {
				return fmt.Errorf("symbol %q already bound to %s", qualified, prev)
			}
		}
		ctx.moduleSymbols[qualified] = goName
	}
	return nil
}

func localDefReplacesImport(ctx *compileContext, prev, goName string) bool {
	return ctx.packageMode && strings.Contains(prev, ".") && !strings.Contains(goName, ".")
}

func (ctx compileContext) defGoIdent(localName string) (string, error) {
	if ctx.packageMode {
		if ctx.exportedNames[localName] {
			return exportedGoIdent(localName)
		}
		return toGoIdentifier(localName)
	}
	return moduleGoIdent(ctx.namespace, localName)
}

func (ctx *compileContext) addGoImport(pkgName string) {
	path := GeneratedModulePath + "/" + pkgName
	for _, existing := range ctx.goImports {
		if existing.Path == path {
			return
		}
	}
	ctx.goImports = append(ctx.goImports, goImport{Name: pkgName, Path: path})
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
		macroForms:        append([]Expr(nil), ctx.macroForms...),
		prologueFns:       ctx.prologueFns,
		prologueVars:      ctx.prologueVars,
		constants:         ctx.constants,
		namespace:         ctx.namespace,
		selfFunctionName:  ctx.selfFunctionName,
		selfFunctionArity: ctx.selfFunctionArity,
		selfArityName:     ctx.selfArityName,
		selfVariadicName:  ctx.selfVariadicName,
		selfFunctionRest:  ctx.selfFunctionRest,
		allowRedefine:     ctx.allowRedefine,
		ifTemps:           ctx.ifTemps,
		packageMode:       ctx.packageMode,
		exportedNames:     ctx.exportedNames,
	}
	if ctx.selfArityNames != nil {
		out.selfArityNames = make(map[int]string, len(ctx.selfArityNames))
		for k, v := range ctx.selfArityNames {
			out.selfArityNames[k] = v
		}
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
	return renderIRExpr(ctx.internIR("Kw_"+sanitizeKeywordIdent(name), rtCall("NewKeyword", IRString{Value: name})))
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
	return renderIRExpr(ctx.internIR(hint, rtCall("NewString", IRRaw{Code: quotedCode}))), true
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
	case PipeVectorExpr:
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

//go:embed prologue.flag
var standardPrologueSource string

func newEmptyCompileContext() compileContext {
	ifTemps := 0
	return compileContext{
		functions:     make(map[string]functionDef),
		globals:       make(map[string]exprKind),
		moduleSymbols: make(map[string]string),
		recordTypes:   make(map[string]string),
		ifTemps:       &ifTemps,
	}
}

func newCompileContext() (compileContext, error) {
	ctx := newEmptyCompileContext()
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
	var seed []Expr
	var rest []Expr
	for _, form := range prologueAST.Forms {
		if _, isMap := form.(MapExpr); isMap {
			continue
		}
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
			seed = append(seed, form)
		case "defn", "def":
			rest = append(rest, form)
		default:
			return fmt.Errorf("invalid compiler prologue form %q", head.Name)
		}
	}
	ctx.macroForms = seed
	expanded, err := expandFLAGForms(seed, rest)
	if err != nil {
		return fmt.Errorf("compile compiler prologue: %w", err)
	}
	if ctx.packageMode {
		if ctx.exportedNames == nil {
			ctx.exportedNames = map[string]bool{}
		}
		for _, form := range expanded {
			if name := topLevelDefName(form); name != "" {
				ctx.exportedNames[name] = true
			}
		}
	}
	for _, form := range expanded {
		list, ok := form.(ListExpr)
		if !ok || len(list.Elements) == 0 {
			return fmt.Errorf("invalid compiler prologue form")
		}
		head, ok := list.Elements[0].(SymbolExpr)
		if !ok {
			return fmt.Errorf("invalid compiler prologue form")
		}
		switch head.Name {
		case "defn":
			def, err := compileDefn(list, *ctx)
			if err != nil {
				return fmt.Errorf("compile compiler prologue: %w", err)
			}
			ctx.functions[def.goName] = def
			ctx.globals[def.goName] = exprKindValue
			if ctx.packageMode {
				if err := ctx.bindModuleName(def.flagName, def.goName); err != nil {
					return fmt.Errorf("compile compiler prologue: %w", err)
				}
			}
			ctx.prologueFns = append(ctx.prologueFns, def)
			bindExpr, err := defnBindingExpr(def.goName, def.variadicName)
			if err != nil {
				return fmt.Errorf("compile compiler prologue: %w", err)
			}
			ctx.prologueVars = append(ctx.prologueVars, varDef{
				flagName: def.flagName,
				goName:   def.goName,
				expr:     bindExpr,
			})
		case "def":
			binding, kind, err := compileDef(list, *ctx)
			if err != nil {
				return fmt.Errorf("compile compiler prologue: %w", err)
			}
			ctx.globals[binding.goName] = kind
			if ctx.packageMode {
				if err := ctx.bindModuleName(binding.flagName, binding.goName); err != nil {
					return fmt.Errorf("compile compiler prologue: %w", err)
				}
			}
			ctx.prologueVars = append(ctx.prologueVars, binding)
		default:
			return fmt.Errorf("invalid compiler prologue form %q", head.Name)
		}
	}
	return nil
}

func topLevelDefName(form Expr) string {
	list, ok := form.(ListExpr)
	if !ok || len(list.Elements) < 2 {
		return ""
	}
	head, ok := list.Elements[0].(SymbolExpr)
	if !ok || (head.Name != "defn" && head.Name != "def") {
		return ""
	}
	name, ok := unwrapMetaExpr(list.Elements[1]).(SymbolExpr)
	if !ok {
		return ""
	}
	return name.Name
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
	namespace    string
	typeDecls    []string
	functions    []functionDef
	vars         []varDef
	stmts        []mainStmt
	tests        []testCase
	needsFmt     bool
	extraImports []goImport
}

func emitGoProgram(result compileResult) ([]byte, error) {
	return emitGoFile("main", result, true)
}

func emitGoFile(pkgName string, result compileResult, includeMain bool) ([]byte, error) {
	return flagCompileProgram(pkgName, result, includeMain)
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
	expr = flattenToExpr(expr)
	if expr.kind == exprKindValue {
		return fmt.Sprintf("%s.ValueToAny(%s)", runtimeAlias, expr.code), nil
	}
	return expr.code, nil
}

type ReplCompiler struct {
	ctx           compileContext
	loadedModules map[string]bool
	moduleDefs    map[string]map[string]string
	moduleMacros  map[string]map[string]Expr
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
		moduleMacros:  map[string]map[string]Expr{},
		modulesByPath: map[string]*Module{},
	}
}

func (r *ReplCompiler) PrologueSetup() ReplCompiled {
	parts := make([]string, 0, len(r.ctx.prologueFns)*4+len(r.ctx.prologueVars)*2)
	emitted := make(map[string]bool, len(r.ctx.prologueFns))
	for _, def := range r.ctx.prologueFns {
		// Emit full function declarations (arity + variadic wrappers) so
		// self-recursive prologue fns can resolve their direct arity symbol.
		bindExpr, bindErr := defnBindingExpr(def.goName, def.variadicName)
		if bindErr != nil {
			bindExpr = fmt.Sprintf("%s.NewFunction(%s)", runtimeAlias, def.variadicName)
		}
		parts = append(parts,
			strings.TrimSpace(renderFunctionDef(def)),
			fmt.Sprintf("var %s flagrt.Value", def.goName),
			fmt.Sprintf("%s = %s", def.goName, bindExpr),
		)
		emitted[def.goName] = true
	}
	for _, binding := range r.ctx.prologueVars {
		if emitted[binding.goName] {
			continue
		}
		parts = append(parts,
			fmt.Sprintf("var %s flagrt.Value", binding.goName),
			fmt.Sprintf("%s = %s", binding.goName, binding.expr),
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
	emitVar := func(binding varDef) {
		setup := fmt.Sprintf("%s = %s", binding.goName, binding.expr)
		if _, exists := knownVars[binding.goName]; !exists {
			setup = fmt.Sprintf("var %s flagrt.Value;;%s = %s", binding.goName, binding.goName, binding.expr)
		}
		out = append(out, ReplCompiled{Setup: setup})
	}
	// Yaegi evals snippets in order. FLAG defns may call :go-exports vars, so
	// bind those before compiling function literals. Function wrapper vars
	// (NewFunction(..._variadic)) must come after the functions exist.
	var goBinds, fnVars []varDef
	for _, binding := range result.vars {
		if strings.Contains(binding.expr, "GoBind_") {
			goBinds = append(goBinds, binding)
		} else {
			fnVars = append(fnVars, binding)
		}
	}
	for _, binding := range goBinds {
		emitVar(binding)
	}
	for _, def := range result.functions {
		_, exists := knownFns[def.goName]
		if def.goSignature != "" {
			out = append(out, ReplCompiled{Setup: strings.TrimSpace(renderFunctionDef(def))})
			continue
		}
		if len(def.arities) > 0 {
			setupParts := make([]string, 0, 2+len(def.arities)*2)
			if !exists {
				for _, a := range def.arities {
					mini := arityAsFunctionDef(def, a)
					setupParts = append(setupParts, fmt.Sprintf("var %s %s", a.arityName, renderDirectFunctionType(mini)))
				}
				setupParts = append(setupParts, fmt.Sprintf("var %s func(args ...flagrt.Value) flagrt.Value", def.variadicName))
			}
			for _, a := range def.arities {
				mini := arityAsFunctionDef(def, a)
				setupParts = append(setupParts, fmt.Sprintf("%s = %s", a.arityName, renderDirectFunctionLiteral(mini)))
			}
			setupParts = append(setupParts, fmt.Sprintf("%s = %s", def.variadicName, renderMultiArityVariadicLiteral(def)))
			out = append(out, ReplCompiled{Setup: strings.Join(setupParts, ";;")})
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
	for _, binding := range fnVars {
		emitVar(binding)
	}
	return out
}

func replCompiledFromBody(result compileResult, definedMacros map[string]Expr) ReplCompiled {
	stmtParts := make([]string, 0, len(result.stmts)+len(result.tests))
	var resultExpr string
	for _, stmt := range result.stmts {
		if stmt.prelude != "" {
			stmtParts = append(stmtParts, strings.TrimSpace(stmt.prelude))
		}
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
	if path != "" {
		return compileProgram(path, nil)
	}
	if mod.HasModuleHeader && len(mod.Header.Imports) > 0 {
		return compileResult{}, fmt.Errorf("module imports require compiling from a file path (use flag-lang build <file.flag>)")
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
	return flagCompileInlineProgram(entryPath, testPaths)
}

func compileProgramGo(entryPath string, testPaths []string) (compileResult, error) {
	prog, err := loadProgramGo(entryPath)
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
	moduleMacros := map[string]map[string]Expr{}
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

func seedImports(ctx *compileContext, mod *Module, byPath map[string]*Module, moduleDefs map[string]map[string]string, moduleMacros map[string]map[string]Expr) error {
	return flagSeedImports(ctx, mod, byPath, moduleDefs, moduleMacros)
}

func seedImportsGo(ctx *compileContext, mod *Module, byPath map[string]*Module, moduleDefs map[string]map[string]string, moduleMacros map[string]map[string]Expr) error {
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
			macros = map[string]Expr{}
		}

		pkgName := ""
		if ctx.packageMode {
			var err error
			pkgName, err = goPackageName(provider.Header.Namespace)
			if err != nil {
				return err
			}
			for exp := range exports {
				if _, ok := defs[exp]; ok {
					ctx.addGoImport(pkgName)
					break
				}
			}
		}

		for exp := range exports {
			if goName, ok := defs[exp]; ok {
				qual := prefix + "/" + exp
				ref := goName
				if ctx.packageMode {
					ref = pkgName + "." + goName
				}
				if err := ctx.bindModuleName(qual, ref); err != nil {
					return err
				}
				continue
			}
			if mac, ok := macros[exp]; ok {
				// Qualified macro name so expansions can use async/go without :refer.
				ctx.macroForms = append(ctx.macroForms, renameDefmacro(mac, prefix+"/"+exp))
				continue
			}
			return fmt.Errorf("import %q: exported %q has no definition", spec.Path, exp)
		}
		for _, ref := range spec.Refer {
			if !exports[ref] {
				return fmt.Errorf("import %q: cannot :refer %q (not in :exports)", spec.Path, ref)
			}
			if goName, ok := defs[ref]; ok {
				referName := goName
				if ctx.packageMode {
					referName = pkgName + "." + goName
				}
				if err := ctx.bindModuleName(ref, referName); err != nil {
					return fmt.Errorf("import %q: :refer %q: %w", spec.Path, ref, err)
				}
				continue
			}
			if mac, ok := macros[ref]; ok {
				// :refer installs bare macro name; ok to overwrite same-name macro.
				ctx.macroForms = append(ctx.macroForms, renameDefmacro(mac, ref))
				continue
			}
			return fmt.Errorf("import %q: referred %q has no definition", spec.Path, ref)
		}
	}
	return nil
}

func validateExports(mod *Module, defined map[string]string, definedMacros map[string]Expr) error {
	return flagValidateExports(mod, defined, definedMacros)
}

func validateExportsGo(mod *Module, defined map[string]string, definedMacros map[string]Expr) error {
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
func compileModuleBody(mod *Module, ctx *compileContext, allowTopLevel bool) (compileResult, map[string]string, map[string]Expr, error) {
	vars := make([]varDef, 0, len(mod.Forms))
	defined := map[string]string{}
	definedMacros := collectDefmacros(mod.Forms)
	expandedForms, err := expandFLAGForms(ctx.macroForms, mod.Forms)
	if err != nil {
		return compileResult{}, nil, nil, err
	}
	for _, form := range mod.Forms {
		if _, ok := defmacroFormName(form); ok {
			ctx.macroForms = append(ctx.macroForms, form)
		}
	}

	// Native library re-exports (:go-exports) become package-level vars bound to
	// generated flagrt.GoBind_* adapters.
	if mod.HasModuleHeader {
		for localName, hostKey := range mod.Header.GoExports {
			rtIdent, ok := libraryGoBinds[hostKey]
			if !ok {
				return compileResult{}, nil, nil, fmt.Errorf("unknown :go-exports host bind %q for %s", hostKey, localName)
			}
			goName, err := ctx.defGoIdent(localName)
			if err != nil {
				return compileResult{}, nil, nil, err
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

	flagResult, flagDefined, flagErr := flagCompileModuleBody(expandedForms, ctx, allowTopLevel)
	if flagErr != nil {
		return compileResult{}, nil, nil, flagErr
	}
	for name, goName := range flagDefined {
		defined[name] = goName
	}
	flagResult.vars = append(vars, flagResult.vars...)
	return flagResult, defined, definedMacros, nil
}

func compileDefn(form ListExpr, ctx compileContext) (functionDef, error) {
	if len(form.Elements) < 3 {
		return functionDef{}, exprError(form, "defn expects name, optional docstring, vector params, and body")
	}

	nameExpr, ok := unwrapMetaExpr(form.Elements[1]).(SymbolExpr)
	if !ok || nameExpr.Name == "" {
		return functionDef{}, exprError(form, "defn expects a function name")
	}

	doc := ""
	paramsIndex := 2
	bodyStartIndex := 3
	if len(form.Elements) > 2 {
		if docExpr, ok := form.Elements[2].(StringExpr); ok {
			doc = docExpr.Value
			paramsIndex = 3
			bodyStartIndex = 4
		}
	}

	goName, err := ctx.defGoIdent(nameExpr.Name)
	if err != nil {
		return functionDef{}, err
	}

	if paramsIndex >= len(form.Elements) {
		return functionDef{}, exprError(form, "defn expects name, optional docstring, vector params, and body")
	}
	if _, ok := form.Elements[paramsIndex].(ListExpr); ok {
		return compileMultiArityDefn(form, nameExpr, goName, doc, paramsIndex, ctx)
	}

	paramsExpr, ok := form.Elements[paramsIndex].(VectorExpr)
	if !ok {
		return functionDef{}, exprError(form.Elements[paramsIndex], "defn expects a parameter vector")
	}
	params, localSymbols, initStmts, hasRest, err := bindLambdaParamStmts(paramsExpr, ctx, nil, "defn")
	if err != nil {
		return functionDef{}, err
	}
	if len(form.Elements) <= bodyStartIndex {
		return functionDef{}, exprError(form, "defn expects name, optional docstring, vector params, and body")
	}

	fnCtx := copyCompileContext(ctx)
	fnCtx.selfFunctionName = nameExpr.Name
	fnCtx.selfFunctionArity = len(params)
	fnCtx.selfVariadicName = goName + "_variadic"
	fnCtx.selfFunctionRest = hasRest
	fnCtx.selfArityName = fmt.Sprintf("%s_arity_%d", goName, len(params))
	if hasRest {
		fnCtx.selfArityName = fnCtx.selfVariadicName
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
	body, err := flagCompileForms(bodyExprs, fnCtx, localSymbols)
	if err != nil {
		return functionDef{}, err
	}
	bodySource := bodyExprs[len(bodyExprs)-1]
	body, err = coerceExprToValue(body, bodySource, "defn body", ctx)
	if err != nil {
		return functionDef{}, err
	}

	decls, err := flagDefnToIR(fnCtx.selfArityName, fnCtx.selfVariadicName, goName, params, hasRest, initStmts, body.stmts, irFromGoExpr(body))
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
		localInits:   irStmtsToLines(initStmts),
		body:         body.code,
		bodyPrelude:  goExprPrelude(body),
		irDecls:      decls,
	}, nil
}

func compileMultiArityDefn(form ListExpr, nameExpr SymbolExpr, goName, doc string, start int, ctx compileContext) (functionDef, error) {
	arityForms := form.Elements[start:]
	if len(arityForms) == 0 {
		return functionDef{}, exprError(form, "defn expects at least one arity")
	}

	type pendingArity struct {
		form       ListExpr
		params     []string
		initStmts  []IRStmt
		localKinds map[string]exprKind
		arityName  string
	}

	pending := make([]pendingArity, 0, len(arityForms))
	seen := make(map[int]struct{}, len(arityForms))
	arityNames := make(map[int]string, len(arityForms))
	for _, raw := range arityForms {
		list, ok := raw.(ListExpr)
		if !ok || len(list.Elements) < 2 {
			return functionDef{}, exprError(raw, "defn arity expects ([params] body...)")
		}
		paramsExpr, ok := list.Elements[0].(VectorExpr)
		if !ok {
			return functionDef{}, exprError(list.Elements[0], "defn arity expects a parameter vector")
		}
		params, localKinds, initStmts, hasRest, err := bindLambdaParamStmts(paramsExpr, ctx, nil, "defn")
		if err != nil {
			return functionDef{}, err
		}
		if hasRest {
			return functionDef{}, exprError(list, "multi-arity defn does not yet support & rest")
		}
		n := len(params)
		if _, dup := seen[n]; dup {
			return functionDef{}, exprError(list, fmt.Sprintf("duplicate arity with %d arguments", n))
		}
		seen[n] = struct{}{}
		arityName := fmt.Sprintf("%s_arity_%d", goName, n)
		arityNames[n] = arityName
		pending = append(pending, pendingArity{
			form:       list,
			params:     params,
			initStmts:  initStmts,
			localKinds: localKinds,
			arityName:  arityName,
		})
	}

	variadicName := goName + "_variadic"
	fnBase := copyCompileContext(ctx)
	fnBase.selfFunctionName = nameExpr.Name
	fnBase.selfVariadicName = variadicName
	fnBase.selfArityNames = arityNames
	fnBase.globals[goName] = exprKindValue
	fnBase.functions[goName] = functionDef{
		flagName:     nameExpr.Name,
		goName:       goName,
		variadicName: variadicName,
		doc:          doc,
		arities:      nil,
	}
	_ = fnBase.bindModuleName(nameExpr.Name, goName)

	flagArities := make([]defnArityIR, 0, len(pending))
	arities := make([]arityDef, 0, len(pending))
	counts := make([]int, 0, len(pending))
	for _, item := range pending {
		fnCtx := copyCompileContext(fnBase)
		fnCtx.selfFunctionArity = len(item.params)
		fnCtx.selfArityName = item.arityName
		fnCtx.selfFunctionRest = false
		bodyExprs := item.form.Elements[1:]
		body, err := flagCompileForms(bodyExprs, fnCtx, item.localKinds)
		if err != nil {
			return functionDef{}, err
		}
		body, err = coerceExprToValue(body, bodyExprs[len(bodyExprs)-1], "defn body", ctx)
		if err != nil {
			return functionDef{}, err
		}
		flagArities = append(flagArities, defnArityIR{
			Name:      item.arityName,
			Params:    item.params,
			InitStmts: item.initStmts,
			BodyStmts: body.stmts,
			BodyExpr:  irFromGoExpr(body),
		})
		arities = append(arities, arityDef{
			params:      item.params,
			localInits:  irStmtsToLines(item.initStmts),
			body:        body.code,
			bodyPrelude: goExprPrelude(body),
			arityName:   item.arityName,
		})
		counts = append(counts, len(item.params))
	}

	decls, err := flagDefnMultiToIR(variadicName, multiArityExpectsMessage(nameExpr.Name, counts), flagArities)
	if err != nil {
		return functionDef{}, err
	}
	return functionDef{
		flagName:     nameExpr.Name,
		goName:       goName,
		variadicName: variadicName,
		doc:          doc,
		arities:      arities,
		irDecls:      decls,
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
	goName, err := ctx.defGoIdent(nameExpr.Name)
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
	valueExpr = flattenToExpr(valueExpr)

	defIR, err := flagDefToIR(goName, irFromGoExpr(valueExpr))
	if err != nil {
		return varDef{}, 0, false, err
	}
	exprCode, err := irVarRHS(defIR)
	if err != nil {
		return varDef{}, 0, false, err
	}

	_, exists := ctx.globals[goName]
	return varDef{flagName: nameExpr.Name, goName: goName, doc: doc, expr: exprCode}, valueExpr.kind, !exists, nil
}

func irVarRHS(stmt IRStmt) (string, error) {
	v, ok := stmt.(IRVar)
	if !ok || v.Expr == nil {
		return "", fmt.Errorf("expected IR var declaration")
	}
	return renderIRExpr(v.Expr), nil
}

func defnBindingExpr(goName, variadicName string) (string, error) {
	stmt, err := flagDefnBindingToIR(goName, variadicName)
	if err != nil {
		return "", err
	}
	return irVarRHS(stmt)
}

func exprToGo(expr Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	got, err := flagCompileForm(expr, ctx, locals)
	if err != nil {
		return goExpr{}, annotateExprError(expr, err)
	}
	return got, nil
}

func quotedLiteralToIR(expr Expr) (IRExpr, error) {
	return flagQuotedToIR(expr)
}

// goFormExprToGo lowers (go body...) to a goroutine that evaluates body for
// side effects and returns nil immediately to the caller.

// futureFormExprToGo lowers (future body...) to NewFuture(func() Value { body }).
// The result is a zero-arg function: call it to block for the value.

// deferExprToGo lowers (defer f) to a Go defer statement that calls zero-arg f.
// The thunk expression is evaluated immediately; the call runs when the enclosing
// compiled Go function (typically a do/let/defn IIFE) returns.

func unwrapMetaExpr(expr Expr) Expr {
	if meta, ok := expr.(MetaExpr); ok {
		return meta.Target
	}
	return expr
}

// Non-map metadata (e.g. ^int, ^long, ^float) is accepted as a no-op hint.

// Nested #() has its own placeholder scope.

func coerceExprToValue(expr goExpr, source Expr, label string, ctx compileContext) (goExpr, error) {
	switch expr.kind {
	case exprKindValue:
		return expr, nil
	case exprKindBool:
		return keepStmts(expr, fromIR(rtCall("NewBool", irFromGoExpr(expr)), exprKindValue)), nil
	case exprKindString:
		if code, ok := ctx.stringLiteralValueCode(source, expr.code); ok {
			return keepStmts(expr, fromIR(IRIdent{Name: code}, exprKindValue)), nil
		}
		return keepStmts(expr, fromIR(rtCall("NewString", irFromGoExpr(expr)), exprKindValue)), nil
	case exprKindDefer:
		return goExpr{}, exprError(source, fmt.Sprintf("%s cannot be a defer form (defer is a statement)", label))
	default:
		return goExpr{}, exprError(source, fmt.Sprintf("%s must evaluate to Value", label))
	}
}

func bindLambdaParamStmts(
	paramsExpr VectorExpr,
	ctx compileContext,
	locals map[string]exprKind,
	label string,
) ([]string, map[string]exprKind, []IRStmt, bool, error) {
	return flagBindParams(paramsExpr, label, locals)
}

func truthyIR(expr goExpr) (IRExpr, error) {
	switch expr.kind {
	case exprKindBool:
		return irFromGoExpr(expr), nil
	case exprKindValue:
		return rtCall("IsTruthy", irFromGoExpr(expr)), nil
	case exprKindString:
		return IRRaw{Code: "true"}, nil
	default:
		return nil, fmt.Errorf("unsupported if test expression")
	}
}

func irStmtsToLines(stmts []IRStmt) []string {
	lines := make([]string, 0, len(stmts))
	for _, stmt := range stmts {
		lines = append(lines, renderIRStmt(stmt, "\t"))
	}
	return lines
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
	if src, ok := flagRenderFunctionDef(fn); ok {
		return src
	}
	return renderFunctionDefGo(fn)
}

func renderFunctionDefGo(fn functionDef) string {
	// Handle go-interface wrappers with custom Go signatures
	if fn.goSignature != "" {
		return fmt.Sprintf("func %s%s {\n%s}\n", fn.goName, fn.goSignature, fn.body)
	}
	if len(fn.irDecls) > 0 {
		return renderIRStmtsGo(fn.irDecls, "")
	}
	if len(fn.arities) > 0 {
		var out strings.Builder
		for _, a := range fn.arities {
			mini := arityAsFunctionDef(fn, a)
			fmt.Fprintf(&out, "func %s(%s) flagrt.Value {\n%s%s\treturn %s\n}\n\n",
				a.arityName,
				renderDirectParamList(mini),
				renderLocalInits(a.localInits),
				a.bodyPrelude,
				a.body,
			)
		}
		fmt.Fprintf(&out, "func %s(args ...flagrt.Value) flagrt.Value {\n%s}\n",
			fn.variadicName,
			renderMultiArityDispatch(fn),
		)
		return out.String()
	}
	if fn.hasRest {
		return fmt.Sprintf("func %s(args ...flagrt.Value) flagrt.Value {\n%s}\n",
			fn.variadicName,
			renderStandaloneVariadicBody(fn),
		)
	}
	return fmt.Sprintf("func %s(%s) flagrt.Value {\n%s%s\treturn %s\n}\n\nfunc %s(args ...flagrt.Value) flagrt.Value {\n%s}\n",
		fn.arityName,
		renderDirectParamList(fn),
		renderLocalInits(fn.localInits),
		fn.bodyPrelude,
		fn.body,
		fn.variadicName,
		renderVariadicFunctionBody(fn),
	)
}

func arityAsFunctionDef(parent functionDef, a arityDef) functionDef {
	return functionDef{
		goName:       parent.goName,
		variadicName: parent.variadicName,
		arityName:    a.arityName,
		hasRest:      a.hasRest,
		params:       a.params,
		localInits:   a.localInits,
		body:         a.body,
		bodyPrelude:  a.bodyPrelude,
	}
}

func renderMultiArityVariadicLiteral(fn functionDef) string {
	return fmt.Sprintf("func(args ...flagrt.Value) flagrt.Value {\n%s}", renderMultiArityDispatch(fn))
}

func renderMultiArityDispatch(fn functionDef) string {
	counts := make([]int, 0, len(fn.arities))
	var body strings.Builder
	body.WriteString("\tswitch len(args) {\n")
	for _, a := range fn.arities {
		n := len(a.params)
		counts = append(counts, n)
		fmt.Fprintf(&body, "\tcase %d:\n", n)
		if n == 0 {
			fmt.Fprintf(&body, "\t\treturn %s()\n", a.arityName)
			continue
		}
		callArgs := make([]string, 0, n)
		for i := 0; i < n; i++ {
			callArgs = append(callArgs, fmt.Sprintf("args[%d]", i))
		}
		fmt.Fprintf(&body, "\t\treturn %s(%s)\n", a.arityName, strings.Join(callArgs, ", "))
	}
	fmt.Fprintf(&body, "\tdefault:\n\t\tpanic(%q)\n\t}\n", multiArityExpectsMessage(fn.flagName, counts))
	return body.String()
}

func multiArityExpectsMessage(name string, counts []int) string {
	if len(counts) == 0 {
		return name + " expects arguments"
	}
	sorted := append([]int(nil), counts...)
	sort.Ints(sorted)
	parts := make([]string, len(sorted))
	for i, n := range sorted {
		parts[i] = strconv.Itoa(n)
	}
	switch len(parts) {
	case 1:
		return fmt.Sprintf("%s expects exactly %s arguments", name, parts[0])
	case 2:
		return fmt.Sprintf("%s expects %s or %s arguments", name, parts[0], parts[1])
	default:
		return fmt.Sprintf("%s expects %s, or %s arguments", name, strings.Join(parts[:len(parts)-1], ", "), parts[len(parts)-1])
	}
}

func renderFunctionLiteral(fn functionDef) string {
	return fmt.Sprintf("func(args ...flagrt.Value) flagrt.Value {\n%s}", renderStandaloneVariadicBody(fn))
}

func renderDirectFunctionType(fn functionDef) string {
	return fmt.Sprintf("func(%s) flagrt.Value", renderDirectParamTypes(fn))
}

func renderDirectFunctionLiteral(fn functionDef) string {
	return fmt.Sprintf("func(%s) flagrt.Value {\n%s%s\treturn %s\n}", renderDirectParamList(fn), renderLocalInits(fn.localInits), fn.bodyPrelude, fn.body)
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
	body.WriteString(fn.bodyPrelude)
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

// Extract target function name

// Extract return type

// Look up the target function

// Extract parameter types from vector

// Generate wrapper function body
// Box parameters

// Call the target FLAG function

// Unbox result

// Create wrapper function name (unmangled)

// Keep it unmangled for Go interop

// Return type for the function signature

// Store the Go signature for special rendering
