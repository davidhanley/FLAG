package compiler

import (
	"fmt"
	"strconv"
	"strings"
)

// IRExpr is a Go-backend-neutral expression node. Lowering builds these;
// renderIRExpr is the only place that turns them into Go source. Unmigrated
// paths still use IRRaw so the rest of the compiler can keep emitting strings.
type IRExpr interface {
	irExpr()
}

type IRIdent struct{ Name string }

func (IRIdent) irExpr() {}

type IRString struct{ Value string }

func (IRString) irExpr() {}

type IRInt struct{ Value int64 }

func (IRInt) irExpr() {}

type IRSelector struct {
	Pkg  string
	Name string
}

func (IRSelector) irExpr() {}

type IRCall struct {
	Fun  IRExpr
	Args []IRExpr
}

func (IRCall) irExpr() {}

// IRRaw is an escape hatch for Go source not yet represented as IR.
type IRRaw struct{ Code string }

func (IRRaw) irExpr() {}

// IRFuncLit is an anonymous function. An IIFE is IRCall{Fun: lit} with no args.
type IRFuncLit struct {
	Result string
	Body   []IRStmt
}

func (IRFuncLit) irExpr() {}

// IRStmt is a Go-backend-neutral statement. Special forms lower to IIFEs
// whose bodies are these nodes.
type IRStmt interface {
	irStmt()
}

type IRExprStmt struct {
	Expr    IRExpr
	Discard bool
}

func (IRExprStmt) irStmt() {}

type IRReturn struct{ Expr IRExpr }

func (IRReturn) irStmt() {}

type IRDefer struct{ Expr IRExpr }

func (IRDefer) irStmt() {}

type IRGoStmt struct{ Expr IRExpr }

func (IRGoStmt) irStmt() {}

type IRVar struct {
	Name string
	Type string
	Expr IRExpr
}

func (IRVar) irStmt() {}

type IRAssign struct {
	Name string
	Expr IRExpr
}

func (IRAssign) irStmt() {}

type IRDefine struct {
	Names []string
	Expr  IRExpr
}

func (IRDefine) irStmt() {}

type IRIfStmt struct {
	Init string
	Cond IRExpr
	Then []IRStmt
	Else []IRStmt
}

func (IRIfStmt) irStmt() {}

type IRForStmt struct{ Body []IRStmt }

func (IRForStmt) irStmt() {}

// IRRawStmt is preformatted Go, including indent and trailing newline.
type IRRawStmt struct{ Code string }

func (IRRawStmt) irStmt() {}

func rtCall(name string, args ...IRExpr) IRExpr {
	return IRCall{Fun: IRSelector{Pkg: runtimeAlias, Name: name}, Args: args}
}

func identCall(name string, args ...IRExpr) IRExpr {
	return IRCall{Fun: IRIdent{Name: name}, Args: args}
}

func selectorCall(pkg, name string, args ...IRExpr) IRExpr {
	return IRCall{Fun: IRSelector{Pkg: pkg, Name: name}, Args: args}
}

func parseGoFun(goName string) IRExpr {
	pkg, name, ok := strings.Cut(goName, ".")
	if ok && pkg != "" && name != "" && !strings.Contains(name, ".") {
		return IRSelector{Pkg: pkg, Name: name}
	}
	return IRIdent{Name: goName}
}

func irFromGoExpr(e goExpr) IRExpr {
	if e.ir != nil {
		return e.ir
	}
	if e.code == "" {
		return nil
	}
	return IRRaw{Code: e.code}
}

func fromIR(ir IRExpr, kind exprKind) goExpr {
	return goExpr{code: renderIRExpr(ir), kind: kind, ir: ir}
}

func fromStmt(stmt IRStmt, kind exprKind) goExpr {
	code := strings.TrimSuffix(renderIRStmt(stmt, ""), "\n")
	return goExpr{code: code, kind: kind, stmt: stmt}
}

func iife(result string, body ...IRStmt) IRExpr {
	return IRCall{Fun: IRFuncLit{Result: result, Body: body}}
}

func valueIIFE(body ...IRStmt) IRExpr {
	return iife(runtimeAlias+".Value", body...)
}

func (ctx compileContext) internIR(hint string, ir IRExpr) IRExpr {
	code := renderIRExpr(ir)
	name := ctx.constCode(hint, code)
	if name == code {
		return ir
	}
	return IRIdent{Name: name}
}

func renderIRExpr(expr IRExpr) string {
	switch e := expr.(type) {
	case IRIdent:
		return e.Name
	case IRString:
		return strconv.Quote(e.Value)
	case IRInt:
		return strconv.FormatInt(e.Value, 10)
	case IRSelector:
		if e.Pkg == "" {
			return e.Name
		}
		return e.Pkg + "." + e.Name
	case IRCall:
		args := make([]string, 0, len(e.Args))
		for _, arg := range e.Args {
			args = append(args, renderIRExpr(arg))
		}
		return fmt.Sprintf("%s(%s)", renderIRExpr(e.Fun), strings.Join(args, ", "))
	case IRFuncLit:
		var b strings.Builder
		if e.Result == "" {
			b.WriteString("func() {\n")
		} else {
			fmt.Fprintf(&b, "func() %s {\n", e.Result)
		}
		b.WriteString(renderIRStmts(e.Body, "\t"))
		b.WriteString("}")
		return b.String()
	case IRRaw:
		return e.Code
	default:
		return ""
	}
}

func renderIRStmts(stmts []IRStmt, indent string) string {
	var b strings.Builder
	for _, stmt := range stmts {
		b.WriteString(renderIRStmt(stmt, indent))
	}
	return b.String()
}

func renderIRStmt(stmt IRStmt, indent string) string {
	switch s := stmt.(type) {
	case IRRawStmt:
		return s.Code
	case IRExprStmt:
		if s.Discard {
			return indent + "_ = " + renderIRExpr(s.Expr) + "\n"
		}
		return indent + renderIRExpr(s.Expr) + "\n"
	case IRReturn:
		if s.Expr == nil {
			return indent + "return\n"
		}
		return indent + "return " + renderIRExpr(s.Expr) + "\n"
	case IRDefer:
		return indent + "defer " + renderIRExpr(s.Expr) + "\n"
	case IRGoStmt:
		return indent + "go " + renderIRExpr(s.Expr) + "\n"
	case IRVar:
		switch {
		case s.Type != "" && s.Expr == nil:
			return indent + "var " + s.Name + " " + s.Type + "\n"
		case s.Type != "":
			return indent + "var " + s.Name + " " + s.Type + " = " + renderIRExpr(s.Expr) + "\n"
		default:
			return indent + "var " + s.Name + " = " + renderIRExpr(s.Expr) + "\n"
		}
	case IRAssign:
		return indent + s.Name + " = " + renderIRExpr(s.Expr) + "\n"
	case IRDefine:
		return indent + strings.Join(s.Names, ", ") + " := " + renderIRExpr(s.Expr) + "\n"
	case IRIfStmt:
		var b strings.Builder
		if s.Init != "" {
			fmt.Fprintf(&b, "%sif %s; %s {\n", indent, s.Init, renderIRExpr(s.Cond))
		} else {
			fmt.Fprintf(&b, "%sif %s {\n", indent, renderIRExpr(s.Cond))
		}
		b.WriteString(renderIRStmts(s.Then, indent+"\t"))
		if len(s.Else) > 0 {
			b.WriteString(indent + "} else {\n")
			b.WriteString(renderIRStmts(s.Else, indent+"\t"))
		}
		b.WriteString(indent + "}\n")
		return b.String()
	case IRForStmt:
		var b strings.Builder
		b.WriteString(indent + "for {\n")
		b.WriteString(renderIRStmts(s.Body, indent+"\t"))
		b.WriteString(indent + "}\n")
		return b.String()
	default:
		return ""
	}
}
