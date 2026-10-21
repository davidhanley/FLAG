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

func flagEqToIR(args []IRExpr) (IRExpr, error) {
	return flagIRCall(compiler__eq_to_ir, irExprsToFlagVector(args))
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
	if inner := flagMapGet(out, "ctx"); !flagrt.IsNil(inner) {
		if ctx.ifTemps != nil {
			n := flagMapGet(inner, "if-temps")
			if !flagrt.IsNil(n) {
				*ctx.ifTemps = int(n.Long())
			}
		}
		applyConstantInterner(ctx.constants, flagMapGet(inner, "constants"))
	}
	if flagrt.IsTruthy(flagMapGet(out, "defer?")) {
		stmt, err := flagValueToIRStmt(flagMapGet(out, "stmt"))
		if err != nil {
			return goExpr{}, err
		}
		got := fromStmt(stmt, exprKindDefer)
		got.stmts = stmts
		return got, nil
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
	return flagrt.Assoc(
		base,
		flagrt.NewKeyword("if-temps"), flagrt.NewLong(int64(n)),
		flagrt.NewKeyword("constants"), constantInternerToFLAG(ctx.constants),
	)
}

func boolMapToFLAG(m map[string]bool) flagrt.Value {
	if len(m) == 0 {
		return flagrt.NewMap()
	}
	pairs := make([]flagrt.Value, 0, len(m)*2)
	for k, v := range m {
		pairs = append(pairs, flagrt.NewString(k), flagrt.NewBool(v))
	}
	return flagrt.NewMap(pairs...)
}

func moduleCompileContext(ctx compileContext, allowTopLevel bool) flagrt.Value {
	return flagrt.Assoc(
		formCompileContext(ctx, nil),
		flagrt.NewKeyword("namespace"), flagrt.NewString(ctx.namespace),
		flagrt.NewKeyword("package-mode"), flagrt.NewBool(ctx.packageMode),
		flagrt.NewKeyword("exported-names"), boolMapToFLAG(ctx.exportedNames),
		flagrt.NewKeyword("allow-redefine"), flagrt.NewBool(ctx.allowRedefine),
		flagrt.NewKeyword("allow-top-level"), flagrt.NewBool(allowTopLevel),
	)
}

func flagStringStringMap(m flagrt.Value) map[string]string {
	out := map[string]string{}
	if flagrt.IsNil(m) {
		return out
	}
	keys := flagrt.Keys(m)
	if flagrt.IsNil(keys) {
		return out
	}
	for _, key := range flagrt.Vec(keys).ArrayValues() {
		out[flagString(key)] = flagString(flagrt.Get(m, key))
	}
	return out
}

func flagModuleFunction(item flagrt.Value) (functionDef, error) {
	decls, err := flagIRStmtSeq(flagMapGet(item, "ir-decls"))
	if err != nil {
		return functionDef{}, err
	}
	params, err := flagIRStringSeq(flagMapGet(item, "params"))
	if err != nil {
		return functionDef{}, err
	}
	fn := functionDef{
		flagName:     flagString(flagMapGet(item, "flag-name")),
		goName:       flagString(flagMapGet(item, "go-name")),
		variadicName: flagString(flagMapGet(item, "variadic-name")),
		arityName:    flagString(flagMapGet(item, "arity-name")),
		hasRest:      flagrt.IsTruthy(flagMapGet(item, "has-rest")),
		doc:          flagString(flagMapGet(item, "doc")),
		params:       params,
		irDecls:      decls,
		goSignature:  flagString(flagMapGet(item, "go-signature")),
	}
	if src := flagString(flagMapGet(item, "body-source")); src != "" {
		fn.body = src
	}
	if raw := flagMapGet(item, "body-expr"); !flagrt.IsNil(raw) {
		bodyIR, convErr := flagValueToIRExpr(raw)
		if convErr != nil {
			return functionDef{}, convErr
		}
		fn.body = renderIRExpr(bodyIR)
	}
	if raw := flagMapGet(item, "body-stmts"); !flagrt.IsNil(raw) {
		stmts, convErr := flagIRStmtSeq(raw)
		if convErr != nil {
			return functionDef{}, convErr
		}
		fn.bodyPrelude = renderIRStmts(stmts, "\t")
	}
	if raw := flagMapGet(item, "init-stmts"); !flagrt.IsNil(raw) {
		inits, convErr := flagIRStmtSeq(raw)
		if convErr != nil {
			return functionDef{}, convErr
		}
		fn.localInits = irStmtsToLines(inits)
	}
	return fn, nil
}

func flagModuleVar(item flagrt.Value) (varDef, error) {
	exprIR, err := flagValueToIRExpr(flagMapGet(item, "expr"))
	if err != nil {
		return varDef{}, err
	}
	return varDef{
		flagName: flagString(flagMapGet(item, "flag-name")),
		goName:   flagString(flagMapGet(item, "go-name")),
		doc:      flagString(flagMapGet(item, "doc")),
		expr:     renderIRExpr(exprIR),
	}, nil
}

func flagMainStmts(raw flagrt.Value) ([]mainStmt, error) {
	if flagrt.IsNil(raw) {
		return nil, nil
	}
	out := make([]mainStmt, 0)
	for _, item := range flagrt.Vec(raw).ArrayValues() {
		preludeStmts, err := flagIRStmtSeq(flagMapGet(item, "prelude-stmts"))
		if err != nil {
			return nil, err
		}
		prelude := strings.TrimSuffix(renderIRStmts(preludeStmts, "\t"), "\n")
		if ds := flagMapGet(item, "defer-stmt"); !flagrt.IsNil(ds) {
			stmt, convErr := flagValueToIRStmt(ds)
			if convErr != nil {
				return nil, convErr
			}
			out = append(out, mainStmt{
				prelude: prelude,
				code:    strings.TrimSuffix(renderIRStmt(stmt, ""), "\n"),
			})
			continue
		}
		expr, err := flagValueToIRExpr(flagMapGet(item, "expr"))
		if err != nil {
			return nil, err
		}
		kind := exprKind(0)
		if k := flagMapGet(item, "kind"); !flagrt.IsNil(k) {
			kind = exprKind(k.Long())
		}
		out = append(out, mainStmt{
			prelude: prelude,
			code:    renderIRExpr(expr),
			kind:    kind,
		})
	}
	return out, nil
}

func flagCompileModuleBody(forms []Expr, ctx *compileContext, allowTopLevel bool) (result compileResult, defined map[string]string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoverFLAG(r)
		}
	}()
	out := flagrt.Call(compiler__compile_module_body, exprsToFlagVector(forms), moduleCompileContext(*ctx, allowTopLevel))
	fns := make([]functionDef, 0)
	if raw := flagMapGet(out, "functions"); !flagrt.IsNil(raw) {
		for _, item := range flagrt.Vec(raw).ArrayValues() {
			fn, convErr := flagModuleFunction(item)
			if convErr != nil {
				return compileResult{}, nil, convErr
			}
			fns = append(fns, fn)
		}
	}
	vars := make([]varDef, 0)
	if raw := flagMapGet(out, "vars"); !flagrt.IsNil(raw) {
		for _, item := range flagrt.Vec(raw).ArrayValues() {
			v, convErr := flagModuleVar(item)
			if convErr != nil {
				return compileResult{}, nil, convErr
			}
			vars = append(vars, v)
		}
	}
	tests := make([]testCase, 0)
	if raw := flagMapGet(out, "tests"); !flagrt.IsNil(raw) {
		for _, item := range flagrt.Vec(raw).ArrayValues() {
			line := 0
			if n := flagMapGet(item, "line"); !flagrt.IsNil(n) {
				line = int(n.Long())
			}
			tests = append(tests, testCase{
				goName:     flagString(flagMapGet(item, "go-name")),
				label:      flagString(flagMapGet(item, "label")),
				line:       line,
				bodySource: flagString(flagMapGet(item, "body-source")),
			})
		}
	}
	inner := flagMapGet(out, "ctx")
	if !flagrt.IsNil(inner) {
		if ctx.ifTemps != nil {
			n := flagMapGet(inner, "if-temps")
			if !flagrt.IsNil(n) {
				*ctx.ifTemps = int(n.Long())
			}
		}
		applyConstantInterner(ctx.constants, flagMapGet(inner, "constants"))
		if mod := flagMapGet(inner, "module"); !flagrt.IsNil(mod) {
			if ctx.moduleSymbols == nil {
				ctx.moduleSymbols = map[string]string{}
			}
			for k, v := range flagStringStringMap(mod) {
				ctx.moduleSymbols[k] = v
			}
		}
		if kinds, kindErr := flagStringKindMap(flagMapGet(inner, "globals")); kindErr == nil {
			if ctx.globals == nil {
				ctx.globals = map[string]exprKind{}
			}
			for k, v := range kinds {
				ctx.globals[k] = v
			}
		}
	}
	if ctx.functions == nil {
		ctx.functions = map[string]functionDef{}
	}
	for _, fn := range fns {
		ctx.functions[fn.goName] = fn
		ctx.globals[fn.goName] = exprKindValue
	}
	for _, v := range vars {
		if _, isFn := ctx.functions[v.goName]; !isFn {
			ctx.globals[v.goName] = exprKindValue
		}
	}
	stmts, err := flagMainStmts(flagMapGet(out, "stmts"))
	if err != nil {
		return compileResult{}, nil, err
	}
	needsFmt := flagrt.IsTruthy(flagMapGet(out, "needs-fmt"))
	typeDecls := []string{}
	if raw := flagMapGet(out, "type-decls"); !flagrt.IsNil(raw) {
		for _, item := range flagrt.Vec(raw).ArrayValues() {
			typeDecls = append(typeDecls, flagString(item))
		}
	}
	return compileResult{
		typeDecls: typeDecls,
		functions: fns,
		vars:      vars,
		tests:     tests,
		stmts:     stmts,
		needsFmt:  needsFmt,
	}, flagStringStringMap(flagMapGet(out, "defined")), nil
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

func intStringMapToFLAG(m map[int]string) flagrt.Value {
	if len(m) == 0 {
		return flagrt.NewMap()
	}
	pairs := make([]flagrt.Value, 0, len(m)*2)
	for k, v := range m {
		pairs = append(pairs, flagrt.NewLong(int64(k)), flagrt.NewString(v))
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
	case "defer":
		return exprKindDefer
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
		flagrt.NewKeyword("self-arity-name"), flagrt.NewString(ctx.selfArityName),
		flagrt.NewKeyword("self-function-rest"), flagrt.NewBool(ctx.selfFunctionRest),
		flagrt.NewKeyword("self-function-arity"), flagrt.NewLong(int64(ctx.selfFunctionArity)),
		flagrt.NewKeyword("self-arity-names"), intStringMapToFLAG(ctx.selfArityNames),
		flagrt.NewKeyword("loop-binding-names"), stringSliceToFLAG(ctx.loopBindingNames),
	)
}

func stringSliceToFLAG(xs []string) flagrt.Value {
	if len(xs) == 0 {
		return flagrt.NewArray()
	}
	vals := make([]flagrt.Value, len(xs))
	for i, s := range xs {
		vals[i] = flagrt.NewString(s)
	}
	return flagrt.NewArray(vals...)
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

func flagKW(name string) flagrt.Value {
	return flagrt.NewKeyword(name)
}

func constantInternerToFLAG(c *constantInterner) flagrt.Value {
	if c == nil {
		return flagrt.NilValue()
	}
	byPairs := make([]flagrt.Value, 0, len(c.byCode)*2)
	for k, v := range c.byCode {
		byPairs = append(byPairs, flagrt.NewString(k), flagrt.NewString(v))
	}
	usedPairs := make([]flagrt.Value, 0, len(c.used)*2)
	for k, ok := range c.used {
		if ok {
			usedPairs = append(usedPairs, flagrt.NewString(k), flagrt.NewBool(true))
		}
	}
	order := make([]flagrt.Value, 0, len(c.order))
	for _, d := range c.order {
		order = append(order, flagrt.NewMap(flagKW("name"), flagrt.NewString(d.name), flagKW("code"), flagrt.NewString(d.code)))
	}
	return flagrt.NewMap(
		flagKW("by-code"), flagrt.NewMap(byPairs...),
		flagKW("used"), flagrt.NewMap(usedPairs...),
		flagKW("order"), flagrt.NewArray(order...),
	)
}

func applyConstantInterner(c *constantInterner, v flagrt.Value) {
	if c == nil || flagrt.IsNil(v) {
		return
	}
	c.byCode = map[string]string{}
	if by := flagrt.Get(v, flagKW("by-code")); !flagrt.IsNil(by) {
		if keys := flagrt.Keys(by); !flagrt.IsNil(keys) {
			for _, key := range flagrt.Vec(keys).ArrayValues() {
				c.byCode[flagString(key)] = flagString(flagrt.Get(by, key))
			}
		}
	}
	c.used = map[string]bool{}
	if used := flagrt.Get(v, flagKW("used")); !flagrt.IsNil(used) {
		if keys := flagrt.Keys(used); !flagrt.IsNil(keys) {
			for _, key := range flagrt.Vec(keys).ArrayValues() {
				c.used[flagString(key)] = true
			}
		}
	}
	c.order = nil
	if order := flagrt.Get(v, flagKW("order")); !flagrt.IsNil(order) {
		for _, item := range flagrt.Vec(order).ArrayValues() {
			c.order = append(c.order, constDecl{
				name: flagString(flagrt.Get(item, flagKW("name"))),
				code: flagString(flagrt.Get(item, flagKW("code"))),
			})
		}
	}
}

func flagStringValues(items []string) flagrt.Value {
	vals := make([]flagrt.Value, 0, len(items))
	for _, s := range items {
		vals = append(vals, flagrt.NewString(s))
	}
	return flagrt.NewArray(vals...)
}

func compileResultToFLAG(result compileResult) (flagrt.Value, error) {
	nsCode := ""
	if result.namespace != "" {
		nsStmt, err := flagNsToIR(result.namespace)
		if err != nil {
			return flagrt.Value{}, err
		}
		nsCode = renderIRStmt(nsStmt, "")
	}
	fns := make([]flagrt.Value, 0, len(result.functions))
	for _, fn := range result.functions {
		fns = append(fns, flagrt.NewMap(
			flagKW("doc"), flagrt.NewString(fn.doc),
			flagKW("source"), flagrt.NewString(renderFunctionDef(fn)),
			flagKW("flag-name"), flagrt.NewString(fn.flagName),
			flagKW("go-name"), flagrt.NewString(fn.goName),
			flagKW("has-rest"), flagrt.NewBool(fn.hasRest),
			flagKW("param-count"), flagrt.NewLong(int64(len(fn.params))),
		))
	}
	vars := make([]flagrt.Value, 0, len(result.vars))
	for _, v := range result.vars {
		vars = append(vars, flagrt.NewMap(
			flagKW("doc"), flagrt.NewString(v.doc),
			flagKW("go-name"), flagrt.NewString(v.goName),
			flagKW("expr"), flagrt.NewString(v.expr),
		))
	}
	stmts := make([]flagrt.Value, 0, len(result.stmts))
	for _, stmt := range result.stmts {
		stmts = append(stmts, flagrt.NewMap(
			flagKW("prelude"), flagrt.NewString(stmt.prelude),
			flagKW("code"), flagrt.NewString(stmt.code),
			flagKW("kind"), flagrt.NewLong(int64(stmt.kind)),
		))
	}
	tests := make([]flagrt.Value, 0, len(result.tests))
	for _, tc := range result.tests {
		tests = append(tests, flagrt.NewMap(
			flagKW("go-name"), flagrt.NewString(tc.goName),
			flagKW("label"), flagrt.NewString(tc.label),
			flagKW("line"), flagrt.NewLong(int64(tc.line)),
			flagKW("body-source"), flagrt.NewString(tc.bodySource),
		))
	}
	imports := make([]flagrt.Value, 0, len(result.extraImports))
	for _, imp := range result.extraImports {
		imports = append(imports, flagrt.NewMap(
			flagKW("name"), flagrt.NewString(imp.Name),
			flagKW("path"), flagrt.NewString(imp.Path),
		))
	}
	return flagrt.NewMap(
		flagKW("namespace"), flagrt.NewString(result.namespace),
		flagKW("namespace-code"), flagrt.NewString(nsCode),
		flagKW("type-decls"), flagStringValues(result.typeDecls),
		flagKW("functions"), flagrt.NewArray(fns...),
		flagKW("vars"), flagrt.NewArray(vars...),
		flagKW("stmts"), flagrt.NewArray(stmts...),
		flagKW("tests"), flagrt.NewArray(tests...),
		flagKW("needs-fmt"), flagrt.NewBool(result.needsFmt),
		flagKW("extra-imports"), flagrt.NewArray(imports...),
	), nil
}

func flagInternName(c *constantInterner, hint, code string) string {
	if c == nil {
		return code
	}
	out := flagrt.Call(compiler__intern_ir, constantInternerToFLAG(c), flagrt.NewString(hint), flagrt.NewString(code))
	applyConstantInterner(c, flagrt.Get(out, flagKW("constants")))
	name := flagString(flagrt.Get(out, flagKW("name")))
	if name == "" {
		return code
	}
	return name
}

func arityDefToFLAG(a arityDef) flagrt.Value {
	return flagrt.NewMap(
		flagKW("arity-name"), flagrt.NewString(a.arityName),
		flagKW("params"), stringsToFlagVector(a.params),
		flagKW("local-inits"), stringsToFlagVector(a.localInits),
		flagKW("body"), flagrt.NewString(a.body),
		flagKW("body-prelude"), flagrt.NewString(a.bodyPrelude),
		flagKW("has-rest"), flagrt.NewBool(a.hasRest),
	)
}

func functionDefToFLAG(fn functionDef) flagrt.Value {
	arities := make([]flagrt.Value, 0, len(fn.arities))
	for _, a := range fn.arities {
		arities = append(arities, arityDefToFLAG(a))
	}
	return flagrt.NewMap(
		flagKW("flag-name"), flagrt.NewString(fn.flagName),
		flagKW("go-name"), flagrt.NewString(fn.goName),
		flagKW("variadic-name"), flagrt.NewString(fn.variadicName),
		flagKW("arity-name"), flagrt.NewString(fn.arityName),
		flagKW("has-rest"), flagrt.NewBool(fn.hasRest),
		flagKW("doc"), flagrt.NewString(fn.doc),
		flagKW("params"), stringsToFlagVector(fn.params),
		flagKW("local-inits"), stringsToFlagVector(fn.localInits),
		flagKW("body"), flagrt.NewString(fn.body),
		flagKW("body-prelude"), flagrt.NewString(fn.bodyPrelude),
		flagKW("go-signature"), flagrt.NewString(fn.goSignature),
		flagKW("arities"), flagrt.NewArray(arities...),
		flagKW("ir-decls"), irStmtsToFlagVector(fn.irDecls),
	)
}

func flagRenderFunctionDef(fn functionDef) (out string, ok bool) {
	defer func() {
		if recover() != nil {
			out = ""
			ok = false
		}
	}()
	return flagString(flagrt.Call(compiler__render_function_def, functionDefToFLAG(fn))), true
}

func flagCompileProgram(pkgName string, result compileResult, includeMain bool) (src []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoverFLAG(r)
		}
	}()
	flagResult, err := compileResultToFLAG(result)
	if err != nil {
		return nil, err
	}
	out := flagrt.Call(compiler__compile_program, flagrt.NewString(pkgName), flagResult, flagrt.NewBool(includeMain))
	return []byte(flagString(out)), nil
}

func flagResolveImportPath(importerPath, importPath string) (path string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoverFLAG(r)
			path = ""
		}
	}()
	out := flagrt.Call(compiler__resolve_import_path, flagrt.NewString(importerPath), flagrt.NewString(importPath))
	return flagString(out), nil
}

func flagLoadProgram(entryPath string) (prog *Program, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoverFLAG(r)
			prog = nil
		}
	}()
	out := flagrt.Call(compiler__load_program, flagrt.NewString(entryPath))
	entry := flagString(flagrt.Get(out, flagKW("entry")))
	mods := flagrt.Get(out, flagKW("modules"))
	return programFromFLAGPaths(entry, mods)
}

func programFromFLAGPaths(entry string, modules flagrt.Value) (*Program, error) {
	prog := &Program{byPath: map[string]*Module{}}
	order := make([]*Module, 0, 8)
	if !flagrt.IsNil(modules) {
		for _, item := range flagrt.Vec(modules).ArrayValues() {
			path := flagString(item)
			mod, err := parseModuleTokenStream(path, TokenizeFileToChannel(path))
			if err != nil {
				return nil, fmt.Errorf("%s: %w", path, err)
			}
			prog.byPath[path] = mod
			order = append(order, mod)
		}
	}
	prog.Modules = order
	prog.Entry = prog.byPath[entry]
	if prog.Entry == nil {
		return nil, fmt.Errorf("load-program: missing entry %s", entry)
	}
	return prog, nil
}

func goImportsToFLAG(imports []goImport) flagrt.Value {
	vals := make([]flagrt.Value, 0, len(imports))
	for _, imp := range imports {
		vals = append(vals, flagrt.NewMap(
			flagKW("name"), flagrt.NewString(imp.Name),
			flagKW("path"), flagrt.NewString(imp.Path),
		))
	}
	return flagrt.NewArray(vals...)
}

func flagGoImports(v flagrt.Value) []goImport {
	if flagrt.IsNil(v) {
		return nil
	}
	items := flagrt.Vec(v).ArrayValues()
	out := make([]goImport, 0, len(items))
	for _, item := range items {
		out = append(out, goImport{
			Name: flagString(flagrt.Get(item, flagKW("name"))),
			Path: flagString(flagrt.Get(item, flagKW("path"))),
		})
	}
	return out
}

func seedCompileContext(ctx compileContext) flagrt.Value {
	return flagrt.Assoc(
		moduleCompileContext(ctx, false),
		flagKW("macro-forms"), exprsToFlagVector(ctx.macroForms),
		flagKW("go-imports"), goImportsToFLAG(ctx.goImports),
	)
}

func applySeededContext(ctx *compileContext, v flagrt.Value) error {
	if ctx == nil || flagrt.IsNil(v) {
		return nil
	}
	ctx.moduleSymbols = flagStringStringMap(flagMapGet(v, "module"))
	ctx.packageMode = flagrt.IsTruthy(flagMapGet(v, "package-mode"))
	ctx.namespace = flagString(flagMapGet(v, "namespace"))
	macros, err := flagExprSeq(flagMapGet(v, "macro-forms"))
	if err != nil {
		return err
	}
	ctx.macroForms = macros
	ctx.goImports = flagGoImports(flagMapGet(v, "go-imports"))
	return nil
}

func moduleToFLAG(mod *Module) flagrt.Value {
	if mod == nil {
		return flagrt.NilValue()
	}
	imports := make([]flagrt.Value, 0, len(mod.Header.Imports))
	for _, spec := range mod.Header.Imports {
		refer := make([]flagrt.Value, 0, len(spec.Refer))
		for _, name := range spec.Refer {
			refer = append(refer, flagrt.NewString(name))
		}
		imports = append(imports, flagrt.NewMap(
			flagKW("path"), flagrt.NewString(spec.Path),
			flagKW("as"), flagrt.NewString(spec.As),
			flagKW("refer"), flagrt.NewArray(refer...),
		))
	}
	exports := make([]flagrt.Value, 0, len(mod.Header.Exports))
	for _, name := range mod.Header.Exports {
		exports = append(exports, flagrt.NewString(name))
	}
	return flagrt.NewMap(
		flagKW("path"), flagrt.NewString(mod.Path),
		flagKW("has-header"), flagrt.NewBool(mod.HasModuleHeader),
		flagKW("namespace"), flagrt.NewString(mod.Header.Namespace),
		flagKW("legacy-ns"), flagrt.NewString(mod.LegacyNS),
		flagKW("exports"), flagrt.NewArray(exports...),
		flagKW("imports"), flagrt.NewArray(imports...),
		flagKW("go-exports"), stringMapToFLAG(mod.Header.GoExports),
		flagKW("forms"), exprsToFlagVector(mod.Forms),
	)
}

func modulesToFLAG(mods []*Module) flagrt.Value {
	vals := make([]flagrt.Value, 0, len(mods))
	for _, mod := range mods {
		vals = append(vals, moduleToFLAG(mod))
	}
	return flagrt.NewArray(vals...)
}

func modulesByPathToFLAG(byPath map[string]*Module) flagrt.Value {
	pairs := make([]flagrt.Value, 0, len(byPath)*2)
	for path, mod := range byPath {
		pairs = append(pairs, flagrt.NewString(path), moduleToFLAG(mod))
	}
	return flagrt.NewMap(pairs...)
}

func nestedStringMapsToFLAG(m map[string]map[string]string) flagrt.Value {
	pairs := make([]flagrt.Value, 0, len(m)*2)
	for path, defs := range m {
		pairs = append(pairs, flagrt.NewString(path), stringMapToFLAG(defs))
	}
	return flagrt.NewMap(pairs...)
}

func moduleMacrosToFLAG(m map[string]map[string]Expr) flagrt.Value {
	pairs := make([]flagrt.Value, 0, len(m)*2)
	for path, macros := range m {
		inner := make([]flagrt.Value, 0, len(macros)*2)
		for name, form := range macros {
			inner = append(inner, flagrt.NewString(name), exprToFlagValue(form))
		}
		pairs = append(pairs, flagrt.NewString(path), flagrt.NewMap(inner...))
	}
	return flagrt.NewMap(pairs...)
}

func macrosToFLAG(m map[string]Expr) flagrt.Value {
	pairs := make([]flagrt.Value, 0, len(m)*2)
	for name, form := range m {
		pairs = append(pairs, flagrt.NewString(name), exprToFlagValue(form))
	}
	return flagrt.NewMap(pairs...)
}

func prologueNamesToFLAG(ctx compileContext) flagrt.Value {
	vals := make([]flagrt.Value, 0, len(ctx.prologueFns)+len(ctx.prologueVars))
	for _, fn := range ctx.prologueFns {
		vals = append(vals, flagrt.NewMap(
			flagKW("flag-name"), flagrt.NewString(fn.flagName),
			flagKW("go-name"), flagrt.NewString(fn.goName),
		))
	}
	for _, v := range ctx.prologueVars {
		vals = append(vals, flagrt.NewMap(
			flagKW("flag-name"), flagrt.NewString(v.flagName),
			flagKW("go-name"), flagrt.NewString(v.goName),
		))
	}
	return flagrt.NewArray(vals...)
}

func flagGoPackageName(namespace string) (name string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoverFLAG(r)
			name = ""
		}
	}()
	return flagString(flagrt.Call(compiler__go_package_name, flagrt.NewString(namespace))), nil
}

func flagValidateExports(mod *Module, defined map[string]string, definedMacros map[string]Expr) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoverFLAG(r)
		}
	}()
	flagrt.Call(compiler__validate_exports, moduleToFLAG(mod), stringMapToFLAG(defined), macrosToFLAG(definedMacros))
	return nil
}

func flagSeedImports(ctx *compileContext, mod *Module, byPath map[string]*Module, moduleDefs map[string]map[string]string, moduleMacros map[string]map[string]Expr) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoverFLAG(r)
		}
	}()
	out := flagrt.Call(
		compiler__seed_imports,
		seedCompileContext(*ctx),
		moduleToFLAG(mod),
		modulesByPathToFLAG(byPath),
		nestedStringMapsToFLAG(moduleDefs),
		moduleMacrosToFLAG(moduleMacros),
	)
	return applySeededContext(ctx, out)
}

func flagEmitFlagI(ns string, exports []string, goExports map[string]string, macros map[string]Expr, defined map[string]string) (out []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoverFLAG(r)
			out = nil
		}
	}()
	src := flagrt.Call(
		compiler__emit_flagi,
		flagrt.NewString(ns),
		flagStringValues(exports),
		stringMapToFLAG(goExports),
		macrosToFLAG(macros),
		stringMapToFLAG(defined),
	)
	return []byte(flagString(src)), nil
}

func flagCompiledVar(item flagrt.Value) (varDef, error) {
	expr := flagMapGet(item, "expr")
	if !flagrt.IsNil(expr) && flagrt.Name(flagrt.TypeOf(expr)) == "string" {
		return varDef{
			flagName: flagString(flagMapGet(item, "flag-name")),
			goName:   flagString(flagMapGet(item, "go-name")),
			doc:      flagString(flagMapGet(item, "doc")),
			expr:     flagString(expr),
		}, nil
	}
	return flagModuleVar(item)
}

func compileResultFromFLAG(out flagrt.Value) (compileResult, error) {
	fns := make([]functionDef, 0)
	if raw := flagMapGet(out, "functions"); !flagrt.IsNil(raw) {
		for _, item := range flagrt.Vec(raw).ArrayValues() {
			fn, err := flagModuleFunction(item)
			if err != nil {
				return compileResult{}, err
			}
			fns = append(fns, fn)
		}
	}
	vars := make([]varDef, 0)
	if raw := flagMapGet(out, "vars"); !flagrt.IsNil(raw) {
		for _, item := range flagrt.Vec(raw).ArrayValues() {
			v, err := flagCompiledVar(item)
			if err != nil {
				return compileResult{}, err
			}
			vars = append(vars, v)
		}
	}
	tests := make([]testCase, 0)
	if raw := flagMapGet(out, "tests"); !flagrt.IsNil(raw) {
		for _, item := range flagrt.Vec(raw).ArrayValues() {
			line := 0
			if n := flagMapGet(item, "line"); !flagrt.IsNil(n) {
				line = int(n.Long())
			}
			tests = append(tests, testCase{
				goName:     flagString(flagMapGet(item, "go-name")),
				label:      flagString(flagMapGet(item, "label")),
				line:       line,
				bodySource: flagString(flagMapGet(item, "body-source")),
			})
		}
	}
	stmts, err := flagMainStmts(flagMapGet(out, "stmts"))
	if err != nil {
		return compileResult{}, err
	}
	decls, err := flagIRStringSeq(flagMapGet(out, "type-decls"))
	if err != nil {
		return compileResult{}, err
	}
	return compileResult{
		namespace: flagString(flagMapGet(out, "namespace")),
		typeDecls: decls,
		functions: fns,
		vars:      vars,
		stmts:     stmts,
		tests:     tests,
		needsFmt:  flagrt.IsTruthy(flagMapGet(out, "needs-fmt")),
	}, nil
}

func flagCompileInlineProgram(entryPath string, testPaths []string) (result compileResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoverFLAG(r)
			result = compileResult{}
		}
	}()
	prog, err := LoadProgram(entryPath)
	if err != nil {
		return compileResult{}, err
	}
	if len(testPaths) > 0 {
		if err := appendTestForms(prog.Entry, testPaths); err != nil {
			return compileResult{}, err
		}
	}
	shared := newEmptyCompileContext()
	shared.constants = newConstantInterner()
	out := flagrt.Call(
		compiler__compile_inline_modules,
		modulesToFLAG(prog.Modules),
		seedCompileContext(shared),
		flagrt.NewString(prog.Entry.Path),
	)
	return compileResultFromFLAG(out)
}

func flagCompileProgramPackages(entryPath string) (pkgs []CompiledPackage, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoverFLAG(r)
			pkgs = nil
		}
	}()
	prog, err := LoadProgram(entryPath)
	if err != nil {
		return nil, err
	}
	shared := newEmptyCompileContext()
	shared.packageMode = true
	out := flagrt.Call(
		compiler__compile_package_modules,
		modulesToFLAG(prog.Modules),
		seedCompileContext(shared),
		stringMapToFLAG(nil),
		flagrt.NewString(prog.Entry.Path),
		prologueNamesToFLAG(shared),
	)
	if flagrt.IsNil(out) {
		return nil, nil
	}
	pkgs = []CompiledPackage{}
	for _, item := range flagrt.Vec(out).ArrayValues() {
		src := flagString(flagMapGet(item, "source"))
		flagi := flagString(flagMapGet(item, "flagi"))
		pkg := CompiledPackage{
			Dir:    flagString(flagMapGet(item, "dir")),
			Name:   flagString(flagMapGet(item, "name")),
			Source: []byte(src),
			IsMain: flagrt.IsTruthy(flagMapGet(item, "main?")),
		}
		if flagi != "" {
			pkg.FlagI = []byte(flagi)
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs, nil
}

func compiledPackagesToFLAG(pkgs []CompiledPackage) flagrt.Value {
	vals := make([]flagrt.Value, 0, len(pkgs))
	for _, pkg := range pkgs {
		flagi := ""
		if len(pkg.FlagI) > 0 {
			flagi = string(pkg.FlagI)
		}
		vals = append(vals, flagrt.NewMap(
			flagKW("dir"), flagrt.NewString(pkg.Dir),
			flagKW("name"), flagrt.NewString(pkg.Name),
			flagKW("source"), flagrt.NewString(string(pkg.Source)),
			flagKW("flagi"), flagrt.NewString(flagi),
			flagKW("main?"), flagrt.NewBool(pkg.IsMain),
		))
	}
	return flagrt.NewArray(vals...)
}

func flagFindFlagLangRoot(start string) (root string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoverFLAG(r)
			root = ""
		}
	}()
	return flagString(flagrt.Call(compiler__find_flag_lang_root, flagrt.NewString(start))), nil
}

func flagWriteCompiledPackages(dir string, pkgs []CompiledPackage) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoverFLAG(r)
		}
	}()
	flagrt.Call(compiler__write_compiled_packages, flagrt.NewString(dir), compiledPackagesToFLAG(pkgs))
	return nil
}

func flagWriteProgramPackages(dir string, pkgs []CompiledPackage, entryPath string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoverFLAG(r)
		}
	}()
	flagrt.Call(
		compiler__write_program_packages,
		flagrt.NewString(dir),
		compiledPackagesToFLAG(pkgs),
		flagrt.NewString(entryPath),
	)
	return nil
}
