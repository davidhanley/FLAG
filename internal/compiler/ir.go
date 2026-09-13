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
	case IRRaw:
		return e.Code
	default:
		return ""
	}
}
