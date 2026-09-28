package compiler

import (
	"fmt"
	"sync"

	flagrt "flag-lang/runtime"
)

func isLiteralExpr(expr Expr) bool {
	switch e := expr.(type) {
	case StringExpr, CharExpr, IntExpr, BigIntExpr, RatioExpr, FloatExpr, KeywordExpr, QuotedSymbolExpr:
		return true
	case SymbolExpr:
		return e.Name == "true" || e.Name == "false" || e.Name == "nil"
	default:
		return false
	}
}

func flagLiteralToIR(expr Expr) (IRExpr, error) {
	node := flagrt.Call(compiler__ast_node_to_ir, exprToFlagValue(expr))
	return flagValueToIRExpr(node)
}

func flagQuotedToIR(expr Expr) (ir IRExpr, err error) {
	defer func() {
		if r := recover(); r != nil {
			msg := flagrt.ExMessage(flagrt.PanicValue(r))
			if flagrt.IsNil(msg) {
				err = fmt.Errorf("%v", r)
				return
			}
			err = fmt.Errorf("%s", flagString(msg))
		}
	}()
	node := flagrt.Call(compiler__quoted_ast_to_ir, exprToFlagValue(expr))
	return flagValueToIRExpr(node)
}

func flagCallToIR(callee IRExpr, args []IRExpr) (ir IRExpr, err error) {
	defer func() {
		if r := recover(); r != nil {
			msg := flagrt.ExMessage(flagrt.PanicValue(r))
			if flagrt.IsNil(msg) {
				err = fmt.Errorf("%v", r)
				return
			}
			err = fmt.Errorf("%s", flagString(msg))
		}
	}()
	node := flagrt.Call(compiler__call_to_ir, irExprToFlagValue(callee), irExprsToFlagVector(args))
	return flagValueToIRExpr(node)
}

func flagCallAstToIR(expr Expr) (ir IRExpr, err error) {
	defer func() {
		if r := recover(); r != nil {
			msg := flagrt.ExMessage(flagrt.PanicValue(r))
			if flagrt.IsNil(msg) {
				err = fmt.Errorf("%v", r)
				return
			}
			err = fmt.Errorf("%s", flagString(msg))
		}
	}()
	node := flagrt.Call(compiler__call_ast_to_ir, exprToFlagValue(expr))
	return flagValueToIRExpr(node)
}

func recoverFLAG(r any) error {
	msg := flagrt.ExMessage(flagrt.PanicValue(r))
	if flagrt.IsNil(msg) {
		return fmt.Errorf("%v", r)
	}
	return fmt.Errorf("%s", flagString(msg))
}

func flagIRCall(fn flagrt.Value, args ...flagrt.Value) (ir IRExpr, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoverFLAG(r)
		}
	}()
	return flagValueToIRExpr(flagrt.Call(fn, args...))
}

func flagCtorToIR(ctor string, args []IRExpr) (IRExpr, error) {
	return flagIRCall(compiler__ctor_to_ir, flagrt.NewString(ctor), irExprsToFlagVector(args))
}

func flagRuntimeCallToIR(name string, args []IRExpr) (IRExpr, error) {
	return flagIRCall(compiler__runtime_call_to_ir, flagrt.NewString(name), irExprsToFlagVector(args))
}

func flagFoldCallToIR(name string, args []IRExpr) (IRExpr, error) {
	return flagIRCall(compiler__fold_call_to_ir, flagrt.NewString(name), irExprsToFlagVector(args))
}

func flagEvalAstToIR(expr Expr, ctx flagrt.Value) (IRExpr, error) {
	if flagrt.IsNil(ctx) {
		ctx = flagrt.NewMap()
	}
	return flagIRCall(compiler__eval_ast_to_ir, exprToFlagValue(expr), ctx)
}

func flagIfToIR(name, typeName string, cond IRExpr, thenStmts []IRStmt, thenExpr IRExpr, elseStmts []IRStmt, elseExpr IRExpr) (stmts []IRStmt, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoverFLAG(r)
		}
	}()
	node := flagrt.Call(
		compiler__if_to_ir,
		flagrt.NewString(name),
		flagrt.NewString(typeName),
		irExprToFlagValue(cond),
		irStmtsToFlagVector(thenStmts),
		irExprToFlagValue(thenExpr),
		irStmtsToFlagVector(elseStmts),
		irExprToFlagValue(elseExpr),
	)
	return flagIRStmtSeq(node)
}

func flagLetToIR(name, typeName string, bindingStmts, bodyStmts []IRStmt, bodyExpr IRExpr) (stmts []IRStmt, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoverFLAG(r)
		}
	}()
	node := flagrt.Call(
		compiler__let_to_ir,
		flagrt.NewString(name),
		flagrt.NewString(typeName),
		irStmtsToFlagVector(bindingStmts),
		irStmtsToFlagVector(bodyStmts),
		irExprToFlagValue(bodyExpr),
	)
	return flagIRStmtSeq(node)
}

func goExprToFLAGDoForm(e goExpr) flagrt.Value {
	pairs := []flagrt.Value{
		flagrt.NewKeyword("stmts"), irStmtsToFlagVector(e.stmts),
		flagrt.NewKeyword("expr"), irExprToFlagValue(irFromGoExpr(e)),
		flagrt.NewKeyword("defer?"), flagrt.NewBool(e.kind == exprKindDefer),
	}
	if e.stmt != nil {
		pairs = append(pairs, flagrt.NewKeyword("stmt"), irStmtToFlagValue(e.stmt))
	}
	return flagrt.NewMap(pairs...)
}

func goExprsToFLAGDoForms(compiled []goExpr) flagrt.Value {
	forms := make([]flagrt.Value, 0, len(compiled))
	for _, part := range compiled {
		forms = append(forms, goExprToFLAGDoForm(part))
	}
	return flagrt.NewArray(forms...)
}

func flagDoPreludeStmts(compiled []goExpr) (stmts []IRStmt, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoverFLAG(r)
		}
	}()
	return flagIRStmtSeq(flagrt.Call(compiler__do_prelude_stmts, goExprsToFLAGDoForms(compiled)))
}

func flagDoDeferToIR(typeName string, compiled []goExpr) (ir IRExpr, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoverFLAG(r)
		}
	}()
	return flagValueToIRExpr(flagrt.Call(compiler__do_defer_to_ir, flagrt.NewString(typeName), goExprsToFLAGDoForms(compiled)))
}

func flagLoopToIR(initStmts []IRStmt, bindings []loopBindingIR, bodyStmts []IRStmt, bodyExpr IRExpr) (ir IRExpr, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoverFLAG(r)
		}
	}()
	bindVals := make([]flagrt.Value, 0, len(bindings))
	for _, b := range bindings {
		bindVals = append(bindVals, flagrt.NewMap(
			flagrt.NewKeyword("name"), flagrt.NewString(b.Name),
			flagrt.NewKeyword("init"), irExprToFlagValue(b.Init),
			flagrt.NewKeyword("unused"), flagrt.NewBool(b.Unused),
		))
	}
	return flagValueToIRExpr(flagrt.Call(
		compiler__loop_to_ir,
		irStmtsToFlagVector(initStmts),
		flagrt.NewArray(bindVals...),
		irStmtsToFlagVector(bodyStmts),
		irExprToFlagValue(bodyExpr),
	))
}

func flagRecurToIR(values []IRExpr) (IRExpr, error) {
	return flagIRCall(compiler__recur_to_ir, irExprsToFlagVector(values))
}

type loopBindingIR struct {
	Name   string
	Init   IRExpr
	Unused bool
}

func runtimeGoName(goName string) string {
	if _, name, ok := splitPkgName(goName); ok {
		return name
	}
	return goName
}

func splitPkgName(goName string) (string, string, bool) {
	for i := 0; i < len(goName); i++ {
		if goName[i] == '.' {
			return goName[:i], goName[i+1:], goName[i+1:] != ""
		}
	}
	return "", goName, false
}

var (
	flagGoFnsOnce sync.Once
	flagGoFns     flagrt.Value
)

func goFnBindingsFLAG() flagrt.Value {
	flagGoFnsOnce.Do(func() {
		pairs := make([]flagrt.Value, 0, len(goFnBindings)*2)
		for name, bind := range goFnBindings {
			pairs = append(pairs, flagrt.NewString(name), flagrt.NewString(bind))
		}
		flagGoFns = flagrt.NewMap(pairs...)
	})
	return flagGoFns
}

func kindMapToFLAG(m map[string]exprKind) flagrt.Value {
	if len(m) == 0 {
		return flagrt.NewMap()
	}
	pairs := make([]flagrt.Value, 0, len(m)*2)
	for name, kind := range m {
		pairs = append(pairs, flagrt.NewString(name), flagrt.NewKeyword(flagExprKindName(kind)))
	}
	return flagrt.NewMap(pairs...)
}

func stringMapToFLAG(m map[string]string) flagrt.Value {
	if len(m) == 0 {
		return flagrt.NewMap()
	}
	pairs := make([]flagrt.Value, 0, len(m)*2)
	for k, v := range m {
		pairs = append(pairs, flagrt.NewString(k), flagrt.NewString(v))
	}
	return flagrt.NewMap(pairs...)
}

func functionNamesToFLAG(m map[string]functionDef) flagrt.Value {
	if len(m) == 0 {
		return flagrt.NewMap()
	}
	pairs := make([]flagrt.Value, 0, len(m)*2)
	for name := range m {
		pairs = append(pairs, flagrt.NewString(name), flagrt.NewBool(true))
	}
	return flagrt.NewMap(pairs...)
}

func flagExprKindName(kind exprKind) string {
	switch kind {
	case exprKindString:
		return "string"
	case exprKindBool:
		return "bool"
	case exprKindMutableValue:
		return "mutable-value"
	default:
		return "value"
	}
}

func exprKindFromFLAG(v flagrt.Value) exprKind {
	switch flagKeywordName(v) {
	case "string":
		return exprKindString
	case "bool":
		return exprKindBool
	default:
		return exprKindValue
	}
}

func loweringContext(ctx compileContext, locals map[string]exprKind) flagrt.Value {
	return flagrt.NewMap(
		flagrt.NewKeyword("locals"), kindMapToFLAG(locals),
		flagrt.NewKeyword("globals"), kindMapToFLAG(ctx.globals),
		flagrt.NewKeyword("functions"), functionNamesToFLAG(ctx.functions),
		flagrt.NewKeyword("module"), stringMapToFLAG(ctx.moduleSymbols),
		flagrt.NewKeyword("go-fns"), goFnBindingsFLAG(),
		flagrt.NewKeyword("self-function-name"), flagrt.NewString(ctx.selfFunctionName),
		flagrt.NewKeyword("self-variadic-name"), flagrt.NewString(ctx.selfVariadicName),
	)
}

func flagSymbolToGoExpr(name, ident string, ctx compileContext, locals map[string]exprKind) (result goExpr, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoverFLAG(r)
		}
	}()
	node := flagrt.Call(compiler__symbol_to_ir, flagrt.NewString(name), flagrt.NewString(ident), loweringContext(ctx, locals))
	ir, err := flagValueToIRExpr(flagMapGet(node, "ir"))
	if err != nil {
		return goExpr{}, err
	}
	return fromIR(ir, exprKindFromFLAG(flagMapGet(node, "expr-kind"))), nil
}

func internLiteralIR(expr Expr, ir IRExpr, ctx compileContext) IRExpr {
	switch arg := expr.(type) {
	case CharExpr:
		return ctx.internIR("Char", ir)
	case BigIntExpr:
		return ctx.internIR("Big_"+sanitizeKeywordIdent(arg.Value), ir)
	case RatioExpr:
		return ctx.internIR(fmt.Sprintf("Ratio_%d_%d", arg.Numerator, arg.Denominator), ir)
	case KeywordExpr:
		return ctx.internIR("Kw_"+sanitizeKeywordIdent(arg.Name), ir)
	case QuotedSymbolExpr:
		return ctx.internIR("Sym_"+sanitizeKeywordIdent(arg.Name), ir)
	default:
		return ir
	}
}

func literalExprToGo(expr Expr, ctx compileContext) (goExpr, error) {
	ir, err := flagLiteralToIR(expr)
	if err != nil {
		return goExpr{}, err
	}
	kind := exprKindValue
	if _, ok := ir.(IRString); ok {
		kind = exprKindString
	}
	return fromIR(internLiteralIR(expr, ir, ctx), kind), nil
}
