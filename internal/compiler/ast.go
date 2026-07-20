package compiler

type FileAST struct {
	Forms []Expr
}

type Expr interface {
	expr()
}

type ListExpr struct {
	Elements []Expr
}

func (ListExpr) expr() {}

type VectorExpr struct {
	Elements []Expr
}

func (VectorExpr) expr() {}

type MapExpr struct {
	Entries []Expr
}

func (MapExpr) expr() {}

type SymbolExpr struct {
	Name string
}

func (SymbolExpr) expr() {}

type KeywordExpr struct {
	Name string
}

func (KeywordExpr) expr() {}

type StringExpr struct {
	Value string
}

func (StringExpr) expr() {}

type IntExpr struct {
	Value int64
}

func (IntExpr) expr() {}

type FloatExpr struct {
	Value float64
	Raw   string
}

func (FloatExpr) expr() {}
