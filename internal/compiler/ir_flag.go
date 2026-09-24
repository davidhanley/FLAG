package compiler

import (
	"fmt"

	flagrt "flag-lang/runtime"
)

// flagValueToIRExpr converts a FLAG IR map from libraries/compiler/ir.lib
// into the Go IRExpr used by renderIRExpr.
func flagValueToIRExpr(node flagrt.Value) (IRExpr, error) {
	if flagrt.IsNil(node) {
		return nil, nil
	}
	kind := flagKeywordName(flagMapGet(node, "kind"))
	switch kind {
	case "ident":
		return IRIdent{Name: flagString(flagMapGet(node, "name"))}, nil
	case "string":
		return IRString{Value: flagString(flagMapGet(node, "value"))}, nil
	case "int":
		return IRInt{Value: flagMapGet(node, "value").Long()}, nil
	case "selector":
		return IRSelector{
			Pkg:  flagString(flagMapGet(node, "pkg")),
			Name: flagString(flagMapGet(node, "name")),
		}, nil
	case "call":
		args, err := flagIRExprSeq(flagMapGet(node, "args"))
		if err != nil {
			return nil, err
		}
		fun, err := flagValueToIRExpr(flagMapGet(node, "fun"))
		if err != nil {
			return nil, err
		}
		return IRCall{Fun: fun, Args: args}, nil
	case "raw":
		return IRRaw{Code: flagString(flagMapGet(node, "code"))}, nil
	case "index":
		x, err := flagValueToIRExpr(flagMapGet(node, "x"))
		if err != nil {
			return nil, err
		}
		index, err := flagValueToIRExpr(flagMapGet(node, "index"))
		if err != nil {
			return nil, err
		}
		return IRIndex{X: x, Index: index}, nil
	case "slice":
		x, err := flagValueToIRExpr(flagMapGet(node, "x"))
		if err != nil {
			return nil, err
		}
		low, err := flagValueToIRExpr(flagMapGet(node, "low"))
		if err != nil {
			return nil, err
		}
		high, err := flagValueToIRExpr(flagMapGet(node, "high"))
		if err != nil {
			return nil, err
		}
		return IRSlice{X: x, Low: low, High: high}, nil
	case "spread":
		expr, err := flagValueToIRExpr(flagMapGet(node, "expr"))
		if err != nil {
			return nil, err
		}
		return IRSpread{Expr: expr}, nil
	case "unary":
		x, err := flagValueToIRExpr(flagMapGet(node, "x"))
		if err != nil {
			return nil, err
		}
		return IRUnary{Op: flagString(flagMapGet(node, "op")), X: x}, nil
	case "binary":
		left, err := flagValueToIRExpr(flagMapGet(node, "left"))
		if err != nil {
			return nil, err
		}
		right, err := flagValueToIRExpr(flagMapGet(node, "right"))
		if err != nil {
			return nil, err
		}
		return IRBinary{Op: flagString(flagMapGet(node, "op")), Left: left, Right: right}, nil
	case "func-lit":
		body, err := flagIRStmtSeq(flagMapGet(node, "body"))
		if err != nil {
			return nil, err
		}
		return IRFuncLit{
			Params: flagString(flagMapGet(node, "params")),
			Result: flagString(flagMapGet(node, "result")),
			Body:   body,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported IR expression kind %q", kind)
	}
}

func flagValueToIRStmt(node flagrt.Value) (IRStmt, error) {
	if flagrt.IsNil(node) {
		return nil, fmt.Errorf("unsupported IR statement: nil")
	}
	kind := flagKeywordName(flagMapGet(node, "kind"))
	switch kind {
	case "raw-stmt":
		return IRRawStmt{Code: flagString(flagMapGet(node, "code"))}, nil
	case "expr-stmt":
		expr, err := flagValueToIRExpr(flagMapGet(node, "expr"))
		if err != nil {
			return nil, err
		}
		return IRExprStmt{Expr: expr, Discard: flagIRTruthy(flagMapGet(node, "discard"))}, nil
	case "return":
		expr, err := flagValueToIRExpr(flagMapGet(node, "expr"))
		if err != nil {
			return nil, err
		}
		return IRReturn{Expr: expr}, nil
	case "defer":
		expr, err := flagValueToIRExpr(flagMapGet(node, "expr"))
		if err != nil {
			return nil, err
		}
		return IRDefer{Expr: expr}, nil
	case "go":
		expr, err := flagValueToIRExpr(flagMapGet(node, "expr"))
		if err != nil {
			return nil, err
		}
		return IRGoStmt{Expr: expr}, nil
	case "var":
		expr, err := flagValueToIRExpr(flagMapGet(node, "expr"))
		if err != nil {
			return nil, err
		}
		return IRVar{
			Name: flagString(flagMapGet(node, "name")),
			Type: flagString(flagMapGet(node, "type")),
			Expr: expr,
		}, nil
	case "assign":
		expr, err := flagValueToIRExpr(flagMapGet(node, "expr"))
		if err != nil {
			return nil, err
		}
		return IRAssign{Name: flagString(flagMapGet(node, "name")), Expr: expr}, nil
	case "define":
		names, err := flagIRStringSeq(flagMapGet(node, "names"))
		if err != nil {
			return nil, err
		}
		expr, err := flagValueToIRExpr(flagMapGet(node, "expr"))
		if err != nil {
			return nil, err
		}
		return IRDefine{Names: names, Expr: expr}, nil
	case "if":
		cond, err := flagValueToIRExpr(flagMapGet(node, "cond"))
		if err != nil {
			return nil, err
		}
		thenStmts, err := flagIRStmtSeq(flagMapGet(node, "then"))
		if err != nil {
			return nil, err
		}
		elseStmts, err := flagIRStmtSeq(flagMapGet(node, "else"))
		if err != nil {
			return nil, err
		}
		return IRIfStmt{
			Init: flagString(flagMapGet(node, "init")),
			Cond: cond,
			Then: thenStmts,
			Else: elseStmts,
		}, nil
	case "for":
		body, err := flagIRStmtSeq(flagMapGet(node, "body"))
		if err != nil {
			return nil, err
		}
		return IRForStmt{Body: body}, nil
	default:
		return nil, fmt.Errorf("unsupported IR statement kind %q", kind)
	}
}

func flagIRExprSeq(coll flagrt.Value) ([]IRExpr, error) {
	if flagrt.IsNil(coll) {
		return nil, nil
	}
	items := flagrt.Vec(coll).ArrayValues()
	out := make([]IRExpr, 0, len(items))
	for _, item := range items {
		expr, err := flagValueToIRExpr(item)
		if err != nil {
			return nil, err
		}
		out = append(out, expr)
	}
	return out, nil
}

func flagIRStmtSeq(coll flagrt.Value) ([]IRStmt, error) {
	if flagrt.IsNil(coll) {
		return nil, nil
	}
	items := flagrt.Vec(coll).ArrayValues()
	out := make([]IRStmt, 0, len(items))
	for _, item := range items {
		stmt, err := flagValueToIRStmt(item)
		if err != nil {
			return nil, err
		}
		out = append(out, stmt)
	}
	return out, nil
}

func flagIRStringSeq(coll flagrt.Value) ([]string, error) {
	if flagrt.IsNil(coll) {
		return nil, nil
	}
	items := flagrt.Vec(coll).ArrayValues()
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, flagString(item))
	}
	return out, nil
}

func flagIRTruthy(v flagrt.Value) bool {
	if flagrt.IsNil(v) {
		return false
	}
	return flagrt.IsTruthy(v)
}

func irExprToFlagValue(ir IRExpr) flagrt.Value {
	if ir == nil {
		return flagrt.NilValue()
	}
	switch e := ir.(type) {
	case IRIdent:
		return flagIRNode("ident", flagrt.NewKeyword("name"), flagrt.NewString(e.Name))
	case IRString:
		return flagIRNode("string", flagrt.NewKeyword("value"), flagrt.NewString(e.Value))
	case IRInt:
		return flagIRNode("int", flagrt.NewKeyword("value"), flagrt.NewLong(e.Value))
	case IRSelector:
		return flagIRNode("selector",
			flagrt.NewKeyword("pkg"), flagrt.NewString(e.Pkg),
			flagrt.NewKeyword("name"), flagrt.NewString(e.Name))
	case IRCall:
		return flagIRNode("call",
			flagrt.NewKeyword("fun"), irExprToFlagValue(e.Fun),
			flagrt.NewKeyword("args"), irExprsToFlagVector(e.Args))
	case IRRaw:
		return flagIRNode("raw", flagrt.NewKeyword("code"), flagrt.NewString(e.Code))
	case IRIndex:
		return flagIRNode("index",
			flagrt.NewKeyword("x"), irExprToFlagValue(e.X),
			flagrt.NewKeyword("index"), irExprToFlagValue(e.Index))
	case IRSlice:
		return flagIRNode("slice",
			flagrt.NewKeyword("x"), irExprToFlagValue(e.X),
			flagrt.NewKeyword("low"), irExprToFlagValue(e.Low),
			flagrt.NewKeyword("high"), irExprToFlagValue(e.High))
	case IRSpread:
		return flagIRNode("spread", flagrt.NewKeyword("expr"), irExprToFlagValue(e.Expr))
	case IRUnary:
		return flagIRNode("unary",
			flagrt.NewKeyword("op"), flagrt.NewString(e.Op),
			flagrt.NewKeyword("x"), irExprToFlagValue(e.X))
	case IRBinary:
		return flagIRNode("binary",
			flagrt.NewKeyword("op"), flagrt.NewString(e.Op),
			flagrt.NewKeyword("left"), irExprToFlagValue(e.Left),
			flagrt.NewKeyword("right"), irExprToFlagValue(e.Right))
	case IRFuncLit:
		return flagIRNode("func-lit",
			flagrt.NewKeyword("params"), flagrt.NewString(e.Params),
			flagrt.NewKeyword("result"), flagrt.NewString(e.Result),
			flagrt.NewKeyword("body"), irStmtsToFlagVector(e.Body))
	default:
		return flagrt.NilValue()
	}
}

func irStmtToFlagValue(stmt IRStmt) flagrt.Value {
	if stmt == nil {
		return flagrt.NilValue()
	}
	switch s := stmt.(type) {
	case IRRawStmt:
		return flagIRNode("raw-stmt", flagrt.NewKeyword("code"), flagrt.NewString(s.Code))
	case IRExprStmt:
		return flagIRNode("expr-stmt",
			flagrt.NewKeyword("expr"), irExprToFlagValue(s.Expr),
			flagrt.NewKeyword("discard"), flagrt.NewBool(s.Discard))
	case IRReturn:
		return flagIRNode("return", flagrt.NewKeyword("expr"), irExprToFlagValue(s.Expr))
	case IRDefer:
		return flagIRNode("defer", flagrt.NewKeyword("expr"), irExprToFlagValue(s.Expr))
	case IRGoStmt:
		return flagIRNode("go", flagrt.NewKeyword("expr"), irExprToFlagValue(s.Expr))
	case IRVar:
		return flagIRNode("var",
			flagrt.NewKeyword("name"), flagrt.NewString(s.Name),
			flagrt.NewKeyword("type"), flagrt.NewString(s.Type),
			flagrt.NewKeyword("expr"), irExprToFlagValue(s.Expr))
	case IRAssign:
		return flagIRNode("assign",
			flagrt.NewKeyword("name"), flagrt.NewString(s.Name),
			flagrt.NewKeyword("expr"), irExprToFlagValue(s.Expr))
	case IRDefine:
		return flagIRNode("define",
			flagrt.NewKeyword("names"), stringsToFlagVector(s.Names),
			flagrt.NewKeyword("expr"), irExprToFlagValue(s.Expr))
	case IRIfStmt:
		return flagIRNode("if",
			flagrt.NewKeyword("init"), flagrt.NewString(s.Init),
			flagrt.NewKeyword("cond"), irExprToFlagValue(s.Cond),
			flagrt.NewKeyword("then"), irStmtsToFlagVector(s.Then),
			flagrt.NewKeyword("else"), irStmtsToFlagVector(s.Else))
	case IRForStmt:
		return flagIRNode("for", flagrt.NewKeyword("body"), irStmtsToFlagVector(s.Body))
	default:
		return flagrt.NilValue()
	}
}

func flagIRNode(kind string, extra ...flagrt.Value) flagrt.Value {
	pairs := []flagrt.Value{flagrt.NewKeyword("kind"), flagrt.NewKeyword(kind)}
	pairs = append(pairs, extra...)
	return flagrt.NewMap(pairs...)
}

func irExprsToFlagVector(exprs []IRExpr) flagrt.Value {
	vals := make([]flagrt.Value, 0, len(exprs))
	for _, expr := range exprs {
		vals = append(vals, irExprToFlagValue(expr))
	}
	return flagrt.NewArray(vals...)
}

func irStmtsToFlagVector(stmts []IRStmt) flagrt.Value {
	vals := make([]flagrt.Value, 0, len(stmts))
	for _, stmt := range stmts {
		vals = append(vals, irStmtToFlagValue(stmt))
	}
	return flagrt.NewArray(vals...)
}

func stringsToFlagVector(names []string) flagrt.Value {
	vals := make([]flagrt.Value, 0, len(names))
	for _, name := range names {
		vals = append(vals, flagrt.NewString(name))
	}
	return flagrt.NewArray(vals...)
}
