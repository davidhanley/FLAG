package compiler

import (
	"fmt"
	"strconv"
	"strings"
)

func joinCanonical(nodes []Expr) string {
	parts := make([]string, 0, len(nodes))
	for _, node := range nodes {
		parts = append(parts, exprCanonical(node))
	}
	return strings.Join(parts, " ")
}

func exprCanonical(expr Expr) string {
	switch value := expr.(type) {
	case ListExpr:
		return fmt.Sprintf("(:list :line %d :col %d :elements [%s])", value.Line, value.Col, joinCanonical(value.Elements))
	case VectorExpr:
		return fmt.Sprintf("(:vector :line %d :col %d :elements [%s])", value.Line, value.Col, joinCanonical(value.Elements))
	case PipeVectorExpr:
		return fmt.Sprintf("(:pipe-vector :line %d :col %d :elements [%s])", value.Line, value.Col, joinCanonical(value.Elements))
	case MapExpr:
		return fmt.Sprintf("(:map :line %d :col %d :entries [%s])", value.Line, value.Col, joinCanonical(value.Entries))
	case SetExpr:
		return fmt.Sprintf("(:set :line %d :col %d :elements [%s])", value.Line, value.Col, joinCanonical(value.Elements))
	case HashFnExpr:
		return fmt.Sprintf("(:hash-fn :line %d :col %d :body %s)", value.Line, value.Col, exprCanonical(value.Body))
	case MetaExpr:
		return fmt.Sprintf("(:meta :line %d :col %d :meta %s :target %s)", value.Line, value.Col, exprCanonical(value.Meta), exprCanonical(value.Target))
	case SymbolExpr:
		return fmt.Sprintf("(:symbol :name %s :line %d :col %d)", value.Name, value.Line, value.Col)
	case KeywordExpr:
		return fmt.Sprintf("(:keyword :name %s :line %d :col %d)", value.Name, value.Line, value.Col)
	case QuotedSymbolExpr:
		return fmt.Sprintf("(:quoted-symbol :name %s :line %d :col %d)", value.Name, value.Line, value.Col)
	case QuotedListExpr:
		return fmt.Sprintf("(:quoted-list :line %d :col %d :elements [%s])", value.Line, value.Col, joinCanonical(value.Elements))
	case StringExpr:
		return fmt.Sprintf("(:string :value %s :line %d :col %d)", strconv.Quote(value.Value), value.Line, value.Col)
	case CharExpr:
		return fmt.Sprintf("(:char :value %s :line %d :col %d)", strconv.Quote(string(value.Value)), value.Line, value.Col)
	case IntExpr:
		return fmt.Sprintf("(:int :value %d :line %d :col %d)", value.Value, value.Line, value.Col)
	case BigIntExpr:
		return fmt.Sprintf("(:bigint :value %s :line %d :col %d)", value.Value, value.Line, value.Col)
	case FloatExpr:
		return fmt.Sprintf("(:float :value %s :raw %s :line %d :col %d)", strconv.FormatFloat(value.Value, 'g', -1, 64), value.Raw, value.Line, value.Col)
	case RatioExpr:
		return fmt.Sprintf("(:ratio :numerator %d :denominator %d :line %d :col %d)", value.Numerator, value.Denominator, value.Line, value.Col)
	default:
		return fmt.Sprintf("(:unknown %T)", expr)
	}
}

func formsCanonical(forms []Expr) string {
	parts := make([]string, 0, len(forms))
	for _, form := range forms {
		parts = append(parts, exprCanonical(form))
	}
	return strings.Join(parts, "\n")
}

func astCanonical(source string) string {
	var forms []Expr
	var err error
	for form := range ParseSourceToChannel(source) {
		if form.Err != nil {
			err = form.Err
			continue
		}
		forms = append(forms, form.Expr)
	}
	if err != nil {
		return "ERROR: " + err.Error()
	}
	return formsCanonical(forms)
}
