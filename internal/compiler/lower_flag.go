package compiler

import (
	"fmt"
	"strings"
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

func flagIRStmtCall(fn flagrt.Value, args ...flagrt.Value) (stmt IRStmt, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoverFLAG(r)
		}
	}()
	return flagValueToIRStmt(flagrt.Call(fn, args...))
}

func flagIRStmtsCall(fn flagrt.Value, args ...flagrt.Value) (stmts []IRStmt, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoverFLAG(r)
		}
	}()
	return flagIRStmtSeq(flagrt.Call(fn, args...))
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

func flagUnhandled(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "not handled by FLAG") ||
		strings.Contains(msg, "unsupported symbol")
}

func flagGoIdent(name string) (ident string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoverFLAG(r)
		}
	}()
	return flagString(flagrt.Call(compiler__go_ident, flagrt.NewString(name))), nil
}

func flagBindParams(paramsExpr VectorExpr, label string, locals map[string]exprKind) (params []string, kinds map[string]exprKind, initStmts []IRStmt, hasRest bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoverFLAG(r)
		}
	}()
	node := flagrt.Call(compiler__bind_params_to_ir, exprToFlagValue(paramsExpr), flagrt.NewString(label), kindMapToFLAG(locals))
	params, err = flagIRStringSeq(flagMapGet(node, "params"))
	if err != nil {
		return nil, nil, nil, false, err
	}
	initStmts, err = flagIRStmtSeq(flagMapGet(node, "init-stmts"))
	if err != nil {
		return nil, nil, nil, false, err
	}
	kinds, err = flagStringKindMap(flagMapGet(node, "locals"))
	if err != nil {
		return nil, nil, nil, false, err
	}
	return params, kinds, initStmts, flagrt.IsTruthy(flagMapGet(node, "has-rest")), nil
}

func flagCompileForm(expr Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	return flagCompileNode(compiler__compile_form_to_ir, exprToFlagValue(expr), ctx, locals)
}

func flagCompileForms(forms []Expr, ctx compileContext, locals map[string]exprKind) (goExpr, error) {
	return flagCompileNode(compiler__compile_forms_to_ir, exprsToFlagVector(forms), ctx, locals)
}

func flagCompileNode(fn, node flagrt.Value, ctx compileContext, locals map[string]exprKind) (result goExpr, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoverFLAG(r)
		}
	}()
	out := flagrt.Call(fn, node, formCompileContext(ctx, locals))
	ir, err := flagValueToIRExpr(flagMapGet(out, "expr"))
	if err != nil {
		return goExpr{}, err
	}
	stmts, err := flagIRStmtSeq(flagMapGet(out, "stmts"))
	if err != nil {
		return goExpr{}, err
	}
	if inner := flagMapGet(out, "ctx"); !flagrt.IsNil(inner) && ctx.ifTemps != nil {
		n := flagMapGet(inner, "if-temps")
		if !flagrt.IsNil(n) {
			*ctx.ifTemps = int(n.Long())
		}
	}
	got := fromIR(ir, exprKindFromFLAG(flagMapGet(out, "expr-kind")))
	got.stmts = stmts
	return got, nil
}

func flagStringKindMap(m flagrt.Value) (map[string]exprKind, error) {
	if flagrt.IsNil(m) {
		return map[string]exprKind{}, nil
	}
	keys := flagrt.Keys(m)
	if flagrt.IsNil(keys) {
		return map[string]exprKind{}, nil
	}
	out := map[string]exprKind{}
	for _, key := range flagrt.Vec(keys).ArrayValues() {
		out[flagString(key)] = exprKindFromFLAG(flagrt.Get(m, key))
	}
	return out, nil
}

func formCompileContext(ctx compileContext, locals map[string]exprKind) flagrt.Value {
	base := loweringContext(ctx, locals)
	n := 0
	if ctx.ifTemps != nil {
		n = *ctx.ifTemps
	}
	return flagrt.Assoc(base, flagrt.NewKeyword("if-temps"), flagrt.NewLong(int64(n)))
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

func flagDeferToIR(thunk IRExpr) (stmt IRStmt, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoverFLAG(r)
		}
	}()
	return flagValueToIRStmt(flagrt.Call(compiler__defer_to_ir, irExprToFlagValue(thunk)))
}

func flagThrowToIR(stmts []IRStmt, value IRExpr) (IRExpr, error) {
	return flagIRCall(compiler__throw_to_ir, irStmtsToFlagVector(stmts), irExprToFlagValue(value))
}

func flagFutureToIR(bodyStmts []IRStmt, bodyExpr IRExpr) (IRExpr, error) {
	return flagIRCall(compiler__future_to_ir, irStmtsToFlagVector(bodyStmts), irExprToFlagValue(bodyExpr))
}

func flagDotoToIR(targetStmts []IRStmt, target IRExpr, steps []goExpr) (IRExpr, error) {
	stepVals := make([]flagrt.Value, 0, len(steps))
	for _, step := range steps {
		stepVals = append(stepVals, flagrt.NewMap(
			flagrt.NewKeyword("stmts"), irStmtsToFlagVector(step.stmts),
			flagrt.NewKeyword("expr"), irExprToFlagValue(irFromGoExpr(step)),
		))
	}
	return flagIRCall(compiler__doto_to_ir, irStmtsToFlagVector(targetStmts), irExprToFlagValue(target), flagrt.NewArray(stepVals...))
}

func flagUpdateBangToIR(name string, stmts []IRStmt, value IRExpr) (IRExpr, error) {
	return flagIRCall(compiler__update_bang_to_ir, flagrt.NewString(name), irStmtsToFlagVector(stmts), irExprToFlagValue(value))
}

func flagCatchHandlerToIR(name string, bodyStmts []IRStmt, bodyExpr IRExpr) (IRExpr, error) {
	return flagIRCall(compiler__catch_handler_to_ir, flagrt.NewString(name), irStmtsToFlagVector(bodyStmts), irExprToFlagValue(bodyExpr))
}

func flagTryToIR(bodyStmts []IRStmt, bodyExpr IRExpr, catches []tryCatchIR, finally *goExpr) (IRExpr, error) {
	catchVals := make([]flagrt.Value, 0, len(catches))
	for _, c := range catches {
		catchVals = append(catchVals, flagrt.NewMap(
			flagrt.NewKeyword("class"), flagrt.NewString(c.Class),
			flagrt.NewKeyword("stmts"), irStmtsToFlagVector(c.Stmts),
			flagrt.NewKeyword("handler"), irExprToFlagValue(c.Handler),
		))
	}
	var finallyVal flagrt.Value
	if finally != nil {
		finallyVal = flagrt.NewMap(
			flagrt.NewKeyword("stmts"), irStmtsToFlagVector(finally.stmts),
			flagrt.NewKeyword("expr"), irExprToFlagValue(irFromGoExpr(*finally)),
		)
	} else {
		finallyVal = flagrt.NilValue()
	}
	return flagIRCall(compiler__try_to_ir, irStmtsToFlagVector(bodyStmts), irExprToFlagValue(bodyExpr), flagrt.NewArray(catchVals...), finallyVal)
}

func flagDoseqToIR(loopExpr IRExpr) (IRExpr, error) {
	return flagIRCall(compiler__doseq_to_ir, irExprToFlagValue(loopExpr))
}

func flagDoseqBodyToIR(bodyStmts []IRStmt, bodyExpr IRExpr) (IRExpr, error) {
	return flagIRCall(compiler__doseq_body_to_ir, irStmtsToFlagVector(bodyStmts), irExprToFlagValue(bodyExpr))
}

func flagMapCatBindingToIR(ident string, unused bool, panicMsg string, rest IRExpr, restStmts []IRStmt, coll IRExpr) (IRExpr, error) {
	return flagIRCall(
		compiler__mapcat_binding_to_ir,
		flagrt.NewString(ident),
		flagrt.NewBool(unused),
		flagrt.NewString(panicMsg),
		irExprToFlagValue(rest),
		irStmtsToFlagVector(restStmts),
		irExprToFlagValue(coll),
	)
}

func flagNsToIR(namespace string) (IRStmt, error) {
	return flagIRStmtCall(compiler__ns_to_ir, flagrt.NewString(namespace))
}

func flagDefToIR(name string, expr IRExpr) (IRStmt, error) {
	return flagIRStmtCall(compiler__def_to_ir, flagrt.NewString(name), irExprToFlagValue(expr))
}

func flagDefnBindingToIR(name, variadicName string) (IRStmt, error) {
	return flagIRStmtCall(compiler__defn_binding_to_ir, flagrt.NewString(name), flagrt.NewString(variadicName))
}

func flagDefnToIR(arityName, variadicName, panicName string, params []string, hasRest bool, initStmts, bodyStmts []IRStmt, bodyExpr IRExpr) ([]IRStmt, error) {
	return flagIRStmtsCall(
		compiler__defn_to_ir,
		flagrt.NewString(arityName),
		flagrt.NewString(variadicName),
		flagrt.NewString(panicName),
		stringsToFlagVector(params),
		flagrt.NewBool(hasRest),
		irStmtsToFlagVector(initStmts),
		irStmtsToFlagVector(bodyStmts),
		irExprToFlagValue(bodyExpr),
	)
}

type defnArityIR struct {
	Name      string
	Params    []string
	InitStmts []IRStmt
	BodyStmts []IRStmt
	BodyExpr  IRExpr
}

func flagDefnMultiToIR(variadicName, panicMsg string, arities []defnArityIR) ([]IRStmt, error) {
	vals := make([]flagrt.Value, 0, len(arities))
	for _, a := range arities {
		vals = append(vals, flagrt.NewMap(
			flagrt.NewKeyword("name"), flagrt.NewString(a.Name),
			flagrt.NewKeyword("params"), stringsToFlagVector(a.Params),
			flagrt.NewKeyword("init-stmts"), irStmtsToFlagVector(a.InitStmts),
			flagrt.NewKeyword("body-stmts"), irStmtsToFlagVector(a.BodyStmts),
			flagrt.NewKeyword("body-expr"), irExprToFlagValue(a.BodyExpr),
			flagrt.NewKeyword("n"), flagrt.NewLong(int64(len(a.Params))),
		))
	}
	return flagIRStmtsCall(compiler__defn_multi_to_ir, flagrt.NewString(variadicName), flagrt.NewString(panicMsg), flagrt.NewArray(vals...))
}

func flagFnToIR(name string, params []string, hasRest bool, initStmts, bodyStmts []IRStmt, bodyExpr IRExpr) (IRExpr, error) {
	vals := make([]flagrt.Value, 0, len(params))
	for _, p := range params {
		vals = append(vals, flagrt.NewString(p))
	}
	return flagIRCall(
		compiler__fn_to_ir,
		flagrt.NewString(name),
		flagrt.NewArray(vals...),
		flagrt.NewBool(hasRest),
		irStmtsToFlagVector(initStmts),
		irStmtsToFlagVector(bodyStmts),
		irExprToFlagValue(bodyExpr),
	)
}

type tryCatchIR struct {
	Class   string
	Stmts   []IRStmt
	Handler IRExpr
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
	case "mutable-value":
		return exprKindMutableValue
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
