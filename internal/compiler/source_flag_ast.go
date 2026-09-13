package compiler

import (
	"fmt"
	"unicode/utf8"

	flagrt "flag-lang/runtime"
)

// ASTForm is one top-level form from the streaming AST builder.
// On failure, Err is set, Expr is nil, and the channel is then closed.
type ASTForm struct {
	Expr Expr
	Err  error
}

func fileASTFromFLAGTokens(tokens flagrt.Value) (FileAST, error) {
	forms := make([]Expr, 0, 8)
	for form := range drainFLAGAST(flagrt.Call(compiler__build_ast_from_tokens, tokens)) {
		if form.Err != nil {
			return FileAST{}, form.Err
		}
		forms = append(forms, form.Expr)
	}
	return FileAST{Forms: forms}, nil
}

// BuildASTFromTokens parses SourceToken values into top-level AST forms
// and writes them to a channel, matching the tokenizer's streaming shape.
func BuildASTFromTokens(tokens <-chan SourceToken) <-chan ASTForm {
	return drainFLAGAST(flagrt.Call(compiler__build_ast_from_tokens, sourceTokensToFLAGChannel(tokens)))
}

// ParseTokenChannel parses a stream of SourceToken values into a file AST.
func ParseTokenChannel(tokens <-chan SourceToken) (FileAST, error) {
	forms := make([]Expr, 0, 8)
	for form := range BuildASTFromTokens(tokens) {
		if form.Err != nil {
			return FileAST{}, form.Err
		}
		forms = append(forms, form.Expr)
	}
	return FileAST{Forms: forms}, nil
}

func sourceTokensToFLAGChannel(tokens <-chan SourceToken) flagrt.Value {
	ch := flagrt.MakeChannel(flagrt.NewLong(64))
	go func() {
		defer flagrt.ChannelClose(ch)
		for tok := range tokens {
			flagrt.ChannelSend(ch, flagrt.NewRecord(tok))
		}
	}()
	return ch
}

func drainFLAGAST(nodes flagrt.Value) <-chan ASTForm {
	out := make(chan ASTForm, 8)
	go func() {
		defer close(out)
		for {
			node := flagrt.ChannelReceive(nodes)
			if flagrt.IsNil(node) {
				return
			}
			expr, err := flagValueToExpr(node)
			if err != nil {
				out <- ASTForm{Err: err}
				return
			}
			out <- ASTForm{Expr: expr}
		}
	}()
	return out
}

func flagValueToExpr(node flagrt.Value) (Expr, error) {
	if flagrt.IsNil(node) {
		return nil, fmt.Errorf("parse error: unexpected nil AST node")
	}
	kind := flagKeywordName(flagMapGet(node, "kind"))
	if kind == "error" {
		return nil, fmt.Errorf("%s", flagString(flagMapGet(node, "message")))
	}
	line := int(flagMapGet(node, "line").Long())
	col := int(flagMapGet(node, "col").Long())
	switch kind {
	case "list":
		elements, err := flagExprSeq(flagMapGet(node, "elements"))
		if err != nil {
			return nil, err
		}
		return ListExpr{Elements: elements, Line: line, Col: col}, nil
	case "vector":
		elements, err := flagExprSeq(flagMapGet(node, "elements"))
		if err != nil {
			return nil, err
		}
		return VectorExpr{Elements: elements, Line: line, Col: col}, nil
	case "pipe-vector":
		elements, err := flagExprSeq(flagMapGet(node, "elements"))
		if err != nil {
			return nil, err
		}
		return PipeVectorExpr{Elements: elements, Line: line, Col: col}, nil
	case "map":
		entries, err := flagExprSeq(flagMapGet(node, "entries"))
		if err != nil {
			return nil, err
		}
		return MapExpr{Entries: entries, Line: line, Col: col}, nil
	case "set":
		elements, err := flagExprSeq(flagMapGet(node, "elements"))
		if err != nil {
			return nil, err
		}
		return SetExpr{Elements: elements, Line: line, Col: col}, nil
	case "hash-fn":
		body, err := flagValueToExpr(flagMapGet(node, "body"))
		if err != nil {
			return nil, err
		}
		return HashFnExpr{Body: body, Line: line, Col: col}, nil
	case "meta":
		meta, err := flagValueToExpr(flagMapGet(node, "meta"))
		if err != nil {
			return nil, err
		}
		target, err := flagValueToExpr(flagMapGet(node, "target"))
		if err != nil {
			return nil, err
		}
		return MetaExpr{Meta: meta, Target: target, Line: line, Col: col}, nil
	case "symbol":
		return SymbolExpr{Name: flagString(flagMapGet(node, "name")), Line: line, Col: col}, nil
	case "keyword":
		return KeywordExpr{Name: flagString(flagMapGet(node, "name")), Line: line, Col: col}, nil
	case "quoted-symbol":
		return QuotedSymbolExpr{Name: flagString(flagMapGet(node, "name")), Line: line, Col: col}, nil
	case "quoted-list":
		elements, err := flagExprSeq(flagMapGet(node, "elements"))
		if err != nil {
			return nil, err
		}
		return QuotedListExpr{Elements: elements, Line: line, Col: col}, nil
	case "string":
		return StringExpr{Value: flagString(flagMapGet(node, "value")), Line: line, Col: col}, nil
	case "char":
		text := flagString(flagMapGet(node, "value"))
		r, size := utf8.DecodeRuneInString(text)
		if text == "" || size != len(text) {
			return nil, fmt.Errorf("parse error at %d:%d: unsupported character literal", line, col)
		}
		return CharExpr{Value: r, Line: line, Col: col}, nil
	case "int":
		return IntExpr{Value: flagMapGet(node, "value").Long(), Line: line, Col: col}, nil
	case "bigint":
		return BigIntExpr{Value: flagString(flagMapGet(node, "value")), Line: line, Col: col}, nil
	case "float":
		return FloatExpr{Value: flagMapGet(node, "value").Double(), Raw: flagString(flagMapGet(node, "raw")), Line: line, Col: col}, nil
	case "ratio":
		return RatioExpr{
			Numerator:   flagMapGet(node, "numerator").Long(),
			Denominator: flagMapGet(node, "denominator").Long(),
			Line:        line,
			Col:         col,
		}, nil
	default:
		return nil, fmt.Errorf("parse error: unsupported AST node %s", kind)
	}
}

func flagExprSeq(coll flagrt.Value) ([]Expr, error) {
	if flagrt.IsNil(coll) {
		return nil, nil
	}
	items := flagrt.Vec(coll).ArrayValues()
	out := make([]Expr, 0, len(items))
	for _, item := range items {
		expr, err := flagValueToExpr(item)
		if err != nil {
			return nil, err
		}
		out = append(out, expr)
	}
	return out, nil
}

func flagMapGet(m flagrt.Value, key string) flagrt.Value {
	return flagrt.Get(m, flagrt.NewKeyword(key))
}

func flagKeywordName(v flagrt.Value) string {
	if flagrt.IsNil(v) {
		return ""
	}
	return flagrt.Name(v)
}

func flagString(v flagrt.Value) string {
	if flagrt.IsNil(v) {
		return ""
	}
	switch flagrt.Name(flagrt.TypeOf(v)) {
	case "string":
		return v.StringValue()
	default:
		return flagrt.ValueToString(v)
	}
}

func expandFLAGForms(seed, forms []Expr) ([]Expr, error) {
	in := flagrt.MakeChannel(flagrt.NewLong(64))
	out := flagrt.Call(compiler__expand_macros, in)
	go func() {
		defer flagrt.ChannelClose(in)
		for _, form := range seed {
			sendFLAGExpr(in, form)
		}
		for _, form := range forms {
			sendFLAGExpr(in, form)
		}
	}()

	expanded := make([]Expr, 0, len(forms))
	for {
		node := flagrt.ChannelReceive(out)
		if flagrt.IsNil(node) {
			return expanded, nil
		}
		expr, err := flagValueToExpr(node)
		if err != nil {
			return nil, err
		}
		expanded = append(expanded, expr)
	}
}

func sendFLAGExpr(ch flagrt.Value, form Expr) {
	if _, ok := form.(CommentExpr); ok {
		return
	}
	flagrt.ChannelSend(ch, exprToFlagValue(form))
}

func exprToFlagValue(expr Expr) flagrt.Value {
	switch e := expr.(type) {
	case ListExpr:
		return flagASTNode("list", e.Line, e.Col, flagrt.NewKeyword("elements"), exprsToFlagVector(e.Elements))
	case VectorExpr:
		return flagASTNode("vector", e.Line, e.Col, flagrt.NewKeyword("elements"), exprsToFlagVector(e.Elements))
	case PipeVectorExpr:
		return flagASTNode("pipe-vector", e.Line, e.Col, flagrt.NewKeyword("elements"), exprsToFlagVector(e.Elements))
	case MapExpr:
		return flagASTNode("map", e.Line, e.Col, flagrt.NewKeyword("entries"), exprsToFlagVector(e.Entries))
	case SetExpr:
		return flagASTNode("set", e.Line, e.Col, flagrt.NewKeyword("elements"), exprsToFlagVector(e.Elements))
	case HashFnExpr:
		return flagASTNode("hash-fn", e.Line, e.Col, flagrt.NewKeyword("body"), exprToFlagValue(e.Body))
	case MetaExpr:
		return flagASTNode("meta", e.Line, e.Col,
			flagrt.NewKeyword("meta"), exprToFlagValue(e.Meta),
			flagrt.NewKeyword("target"), exprToFlagValue(e.Target))
	case SymbolExpr:
		return flagASTNode("symbol", e.Line, e.Col, flagrt.NewKeyword("name"), flagrt.NewString(e.Name))
	case KeywordExpr:
		return flagASTNode("keyword", e.Line, e.Col, flagrt.NewKeyword("name"), flagrt.NewString(e.Name))
	case QuotedSymbolExpr:
		return flagASTNode("quoted-symbol", e.Line, e.Col, flagrt.NewKeyword("name"), flagrt.NewString(e.Name))
	case QuotedListExpr:
		return flagASTNode("quoted-list", e.Line, e.Col, flagrt.NewKeyword("elements"), exprsToFlagVector(e.Elements))
	case StringExpr:
		return flagASTNode("string", e.Line, e.Col, flagrt.NewKeyword("value"), flagrt.NewString(e.Value))
	case CharExpr:
		return flagASTNode("char", e.Line, e.Col, flagrt.NewKeyword("value"), flagrt.NewString(string(e.Value)))
	case IntExpr:
		return flagASTNode("int", e.Line, e.Col, flagrt.NewKeyword("value"), flagrt.NewLong(e.Value))
	case BigIntExpr:
		return flagASTNode("bigint", e.Line, e.Col, flagrt.NewKeyword("value"), flagrt.NewString(e.Value))
	case FloatExpr:
		return flagASTNode("float", e.Line, e.Col,
			flagrt.NewKeyword("value"), flagrt.NewDouble(e.Value),
			flagrt.NewKeyword("raw"), flagrt.NewString(e.Raw))
	case RatioExpr:
		return flagASTNode("ratio", e.Line, e.Col,
			flagrt.NewKeyword("numerator"), flagrt.NewLong(e.Numerator),
			flagrt.NewKeyword("denominator"), flagrt.NewLong(e.Denominator))
	default:
		return flagrt.NilValue()
	}
}

func exprsToFlagVector(exprs []Expr) flagrt.Value {
	vals := make([]flagrt.Value, 0, len(exprs))
	for _, expr := range exprs {
		if _, ok := expr.(CommentExpr); ok {
			continue
		}
		vals = append(vals, exprToFlagValue(expr))
	}
	return flagrt.NewVector(vals...)
}

func flagASTNode(kind string, line, col int, extra ...flagrt.Value) flagrt.Value {
	pairs := []flagrt.Value{
		flagrt.NewKeyword("kind"), flagrt.NewKeyword(kind),
		flagrt.NewKeyword("line"), flagrt.NewLong(int64(line)),
		flagrt.NewKeyword("col"), flagrt.NewLong(int64(col)),
	}
	pairs = append(pairs, extra...)
	return flagrt.NewMap(pairs...)
}

func defmacroFormName(form Expr) (string, bool) {
	list, ok := form.(ListExpr)
	if !ok || len(list.Elements) < 2 {
		return "", false
	}
	head, ok := list.Elements[0].(SymbolExpr)
	if !ok || head.Name != "defmacro" {
		return "", false
	}
	name, ok := unwrapMetaExpr(list.Elements[1]).(SymbolExpr)
	if !ok || name.Name == "" {
		return "", false
	}
	return name.Name, true
}

func collectDefmacros(forms []Expr) map[string]Expr {
	out := map[string]Expr{}
	for _, form := range forms {
		if name, ok := defmacroFormName(form); ok {
			out[name] = form
		}
	}
	return out
}

func renameDefmacro(form Expr, name string) Expr {
	list, ok := form.(ListExpr)
	if !ok || len(list.Elements) < 2 {
		return form
	}
	elems := append([]Expr(nil), list.Elements...)
	elems[1] = SymbolExpr{Name: name, Line: list.Line, Col: list.Col}
	return ListExpr{Elements: elems, Line: list.Line, Col: list.Col}
}
