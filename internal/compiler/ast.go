package compiler

type FileAST struct {
	Forms []Expr
}

type Expr interface {
	expr()
}

type ListExpr struct {
	Elements []Expr
	Line     int
	Col      int
}

func (ListExpr) expr() {}

type VectorExpr struct {
	Elements []Expr
	Line     int
	Col      int
}

func (VectorExpr) expr() {}

type MapExpr struct {
	Entries []Expr
	Line    int
	Col     int
}

func (MapExpr) expr() {}

type SetExpr struct {
	Elements []Expr
	Line     int
	Col      int
}

func (SetExpr) expr() {}

type MetaExpr struct {
	Meta   Expr
	Target Expr
	Line   int
	Col    int
}

func (MetaExpr) expr() {}

type CommentExpr struct{}

func (CommentExpr) expr() {}

type HashFnExpr struct {
	Body Expr
	Line int
	Col  int
}

func (HashFnExpr) expr() {}

type SymbolExpr struct {
	Name string
	Line int
	Col  int
}

func (SymbolExpr) expr() {}

type KeywordExpr struct {
	Name string
	Line int
	Col  int
}

func (KeywordExpr) expr() {}

type QuotedSymbolExpr struct {
	Name string
	Line int
	Col  int
}

func (QuotedSymbolExpr) expr() {}

type QuotedListExpr struct {
	Elements []Expr
	Line     int
	Col      int
}

func (QuotedListExpr) expr() {}

type StringExpr struct {
	Value string
	Line  int
	Col   int
}

func (StringExpr) expr() {}

type CharExpr struct {
	Value rune
	Line  int
	Col   int
}

func (CharExpr) expr() {}

type IntExpr struct {
	Value int64
	Line  int
	Col   int
}

func (IntExpr) expr() {}

type BigIntExpr struct {
	Value string
	Line  int
	Col   int
}

func (BigIntExpr) expr() {}

type FloatExpr struct {
	Value float64
	Raw   string
	Line  int
	Col   int
}

func (FloatExpr) expr() {}

type RatioExpr struct {
	Numerator   int64
	Denominator int64
	Line        int
	Col         int
}

func (RatioExpr) expr() {}
