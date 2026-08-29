package compiler

import (
	"fmt"
	"strconv"
	"strings"
)

type sourceTokenKind int

const (
	sourceTokenEOF sourceTokenKind = iota
	sourceTokenError
	sourceTokenListOpen
	sourceTokenListClose
	sourceTokenVectorOpen
	sourceTokenVectorClose
	sourceTokenMapOpen
	sourceTokenMapClose
	sourceTokenMetadata
	sourceTokenQuote
	sourceTokenDispatchSetOpen
	sourceTokenDispatchFnOpen
	sourceTokenString
	sourceTokenAtom
)

type parsedSourceToken struct {
	Kind    sourceTokenKind
	Lexeme  string
	String  string
	Message string
	Line    int
	Col     int
}

type sourceTokenParser struct {
	tokens    <-chan SourceToken
	lookahead *parsedSourceToken
	last      *SourceToken
}

// ParseTokenChannel parses a stream of SourceToken values into a file AST.
func ParseTokenChannel(tokens <-chan SourceToken) (FileAST, error) {
	p := sourceTokenParser{tokens: tokens}
	forms := make([]Expr, 0, 8)

	for {
		tok, err := p.peek()
		if err != nil {
			return FileAST{}, err
		}
		if tok.Kind == sourceTokenEOF {
			return FileAST{Forms: forms}, nil
		}
		form, err := p.readExpr()
		if err != nil {
			return FileAST{}, err
		}
		if _, ok := form.(CommentExpr); ok {
			continue
		}
		forms = append(forms, form)
	}
}

func (p *sourceTokenParser) peek() (parsedSourceToken, error) {
	if p.lookahead != nil {
		return *p.lookahead, nil
	}
	sourceToken, ok := <-p.tokens
	if !ok {
		eofLine, eofCol := eofPosition(p.last)
		tok := parsedSourceToken{Kind: sourceTokenEOF, Line: eofLine, Col: eofCol}
		p.lookahead = &tok
		return tok, nil
	}
	p.last = &sourceToken
	tok := classifySourceToken(sourceToken)
	p.lookahead = &tok
	if tok.Kind == sourceTokenError {
		return parsedSourceToken{}, parseTokenError(tok)
	}
	return tok, nil
}

func (p *sourceTokenParser) next() (parsedSourceToken, error) {
	tok, err := p.peek()
	if err != nil {
		return parsedSourceToken{}, err
	}
	p.lookahead = nil
	return tok, nil
}

func (p *sourceTokenParser) readExpr() (Expr, error) {
	tok, err := p.peek()
	if err != nil {
		return nil, err
	}

	switch tok.Kind {
	case sourceTokenListOpen:
		return p.readList(sourceTokenListOpen, sourceTokenListClose, false)
	case sourceTokenVectorOpen:
		return p.readList(sourceTokenVectorOpen, sourceTokenVectorClose, true)
	case sourceTokenMapOpen:
		return p.readMap()
	case sourceTokenMetadata:
		return p.readMetadataExpr()
	case sourceTokenQuote:
		return p.readQuotedExpr()
	case sourceTokenDispatchSetOpen:
		return p.readSet()
	case sourceTokenDispatchFnOpen:
		return p.readHashFn()
	case sourceTokenString:
		_, err := p.next()
		if err != nil {
			return nil, err
		}
		return StringExpr{Value: tok.String, Line: tok.Line, Col: tok.Col}, nil
	case sourceTokenAtom:
		_, err := p.next()
		if err != nil {
			return nil, err
		}
		return parseAtomToken(tok)
	case sourceTokenEOF:
		return nil, parseErrorAt(tok.Line, tok.Col, "unexpected end of input")
	default:
		return nil, parseErrorAt(tok.Line, tok.Col, "expected expression")
	}
}

func (p *sourceTokenParser) readList(openKind, closeKind sourceTokenKind, asVector bool) (Expr, error) {
	open, err := p.next()
	if err != nil {
		return nil, err
	}
	elements := make([]Expr, 0, 4)
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if tok.Kind == sourceTokenEOF {
			closer := tokenKindLabel(closeKind)
			return nil, parseErrorAt(open.Line, open.Col, fmt.Sprintf("missing closing %q", closer))
		}
		if tok.Kind == closeKind {
			_, err := p.next()
			if err != nil {
				return nil, err
			}
			if asVector {
				return VectorExpr{Elements: elements, Line: open.Line, Col: open.Col}, nil
			}
			if len(elements) > 0 {
				if sym, ok := elements[0].(SymbolExpr); ok && sym.Name == "comment" {
					return CommentExpr{}, nil
				}
			}
			return ListExpr{Elements: elements, Line: open.Line, Col: open.Col}, nil
		}

		item, err := p.readExpr()
		if err != nil {
			return nil, err
		}
		if _, ok := item.(CommentExpr); ok {
			continue
		}
		elements = append(elements, item)
	}
}

func (p *sourceTokenParser) readMap() (Expr, error) {
	open, err := p.next()
	if err != nil {
		return nil, err
	}
	entries := make([]Expr, 0, 4)
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if tok.Kind == sourceTokenEOF {
			return nil, parseErrorAt(open.Line, open.Col, "missing closing \"}\"")
		}
		if tok.Kind == sourceTokenMapClose {
			_, err := p.next()
			if err != nil {
				return nil, err
			}
			return MapExpr{Entries: entries, Line: open.Line, Col: open.Col}, nil
		}
		item, err := p.readExpr()
		if err != nil {
			return nil, err
		}
		if _, ok := item.(CommentExpr); ok {
			continue
		}
		entries = append(entries, item)
	}
}

func (p *sourceTokenParser) readSet() (Expr, error) {
	open, err := p.next()
	if err != nil {
		return nil, err
	}
	elements := make([]Expr, 0, 4)
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if tok.Kind == sourceTokenEOF {
			return nil, parseErrorAt(open.Line, open.Col, "missing closing \"}\"")
		}
		if tok.Kind == sourceTokenMapClose {
			_, err := p.next()
			if err != nil {
				return nil, err
			}
			return SetExpr{Elements: elements, Line: open.Line, Col: open.Col}, nil
		}
		item, err := p.readExpr()
		if err != nil {
			return nil, err
		}
		if _, ok := item.(CommentExpr); ok {
			continue
		}
		elements = append(elements, item)
	}
}

func (p *sourceTokenParser) readHashFn() (Expr, error) {
	open, err := p.next()
	if err != nil {
		return nil, err
	}
	elements := make([]Expr, 0, 4)
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if tok.Kind == sourceTokenEOF {
			return nil, parseErrorAt(open.Line, open.Col, "missing closing \")\"")
		}
		if tok.Kind == sourceTokenListClose {
			_, err := p.next()
			if err != nil {
				return nil, err
			}
			return HashFnExpr{
				Body: ListExpr{Elements: elements, Line: open.Line, Col: open.Col},
				Line: open.Line,
				Col:  open.Col,
			}, nil
		}
		item, err := p.readExpr()
		if err != nil {
			return nil, err
		}
		if _, ok := item.(CommentExpr); ok {
			continue
		}
		elements = append(elements, item)
	}
}

func (p *sourceTokenParser) readMetadataExpr() (Expr, error) {
	metaTok, err := p.next()
	if err != nil {
		return nil, err
	}
	meta, err := p.readExpr()
	if err != nil {
		return nil, err
	}
	target, err := p.readExpr()
	if err != nil {
		return nil, err
	}
	return MetaExpr{Meta: meta, Target: target, Line: metaTok.Line, Col: metaTok.Col}, nil
}

func (p *sourceTokenParser) readQuotedExpr() (Expr, error) {
	quoteTok, err := p.next()
	if err != nil {
		return nil, err
	}
	quoted, err := p.readExpr()
	if err != nil {
		return nil, err
	}
	switch value := quoted.(type) {
	case SymbolExpr:
		return QuotedSymbolExpr{Name: value.Name, Line: quoteTok.Line, Col: quoteTok.Col}, nil
	case ListExpr:
		return QuotedListExpr{Elements: value.Elements, Line: quoteTok.Line, Col: quoteTok.Col}, nil
	default:
		return nil, parseErrorAt(quoteTok.Line, quoteTok.Col, "quote currently supports symbols and lists")
	}
}

func classifySourceToken(st SourceToken) parsedSourceToken {
	line, col := int(st.Line), int(st.Offset)
	switch st.Token {
	case "(":
		return parsedSourceToken{Kind: sourceTokenListOpen, Line: line, Col: col}
	case ")":
		return parsedSourceToken{Kind: sourceTokenListClose, Line: line, Col: col}
	case "[":
		return parsedSourceToken{Kind: sourceTokenVectorOpen, Line: line, Col: col}
	case "]":
		return parsedSourceToken{Kind: sourceTokenVectorClose, Line: line, Col: col}
	case "{":
		return parsedSourceToken{Kind: sourceTokenMapOpen, Line: line, Col: col}
	case "}":
		return parsedSourceToken{Kind: sourceTokenMapClose, Line: line, Col: col}
	case "^":
		return parsedSourceToken{Kind: sourceTokenMetadata, Line: line, Col: col}
	case "'":
		return parsedSourceToken{Kind: sourceTokenQuote, Line: line, Col: col}
	case "#{":
		return parsedSourceToken{Kind: sourceTokenDispatchSetOpen, Line: line, Col: col}
	case "#(":
		return parsedSourceToken{Kind: sourceTokenDispatchFnOpen, Line: line, Col: col}
	}

	if strings.HasPrefix(st.Token, "#") {
		msg := "unsupported reader dispatch"
		if st.Token == "#" {
			msg = "unexpected end after #"
		}
		return parsedSourceToken{Kind: sourceTokenError, Message: msg, Line: line, Col: col}
	}
	if strings.HasPrefix(st.Token, "\"\"\"") {
		if strings.HasSuffix(st.Token, "\"\"\"") && len(st.Token) >= 6 {
			return parsedSourceToken{
				Kind:   sourceTokenString,
				String: st.Token[3 : len(st.Token)-3],
				Line:   line,
				Col:    col,
			}
		}
		return parsedSourceToken{Kind: sourceTokenError, Message: "unterminated string literal", Line: line, Col: col}
	}
	if strings.HasPrefix(st.Token, "\"") {
		if !strings.HasSuffix(st.Token, "\"") || len(st.Token) < 2 {
			return parsedSourceToken{Kind: sourceTokenError, Message: "unterminated string literal", Line: line, Col: col}
		}
		unquoted, err := strconv.Unquote(st.Token)
		if err != nil {
			return parsedSourceToken{Kind: sourceTokenError, Message: fmt.Sprintf("invalid string literal: %v", err), Line: line, Col: col}
		}
		return parsedSourceToken{Kind: sourceTokenString, String: unquoted, Line: line, Col: col}
	}
	return parsedSourceToken{Kind: sourceTokenAtom, Lexeme: st.Token, Line: line, Col: col}
}

func eofPosition(last *SourceToken) (int, int) {
	if last == nil {
		return 1, 1
	}
	return int(last.Line), int(last.Offset) + sourceTokenWidth(*last)
}

func sourceTokenWidth(token SourceToken) int {
	return len([]rune(token.Token))
}

func parseAtomToken(tok parsedSourceToken) (Expr, error) {
	token := tok.Lexeme
	if strings.HasPrefix(token, ":") && len(token) > 1 {
		return KeywordExpr{Name: token[1:], Line: tok.Line, Col: tok.Col}, nil
	}
	if strings.HasPrefix(token, "\\") && len(token) > 1 {
		charText := token[1:]
		switch charText {
		case "space":
			return CharExpr{Value: ' ', Line: tok.Line, Col: tok.Col}, nil
		case "newline":
			return CharExpr{Value: '\n', Line: tok.Line, Col: tok.Col}, nil
		case "tab":
			return CharExpr{Value: '\t', Line: tok.Line, Col: tok.Col}, nil
		}
		runes := []rune(charText)
		if len(runes) == 1 {
			return CharExpr{Value: runes[0], Line: tok.Line, Col: tok.Col}, nil
		}
		return nil, parseErrorAt(tok.Line, tok.Col, "unsupported character literal")
	}
	if strings.Count(token, "/") == 1 && !strings.HasPrefix(token, "/") && !strings.HasSuffix(token, "/") {
		parts := strings.SplitN(token, "/", 2)
		numerator, err := strconv.ParseInt(parts[0], 10, 64)
		if err == nil {
			denominator, err := strconv.ParseInt(parts[1], 10, 64)
			if err == nil {
				if denominator == 0 {
					return nil, parseErrorAt(tok.Line, tok.Col, "ratio denominator cannot be zero")
				}
				return RatioExpr{Numerator: numerator, Denominator: denominator, Line: tok.Line, Col: tok.Col}, nil
			}
		}
	}
	if strings.HasSuffix(token, "N") {
		candidate := strings.TrimSuffix(token, "N")
		if isSignedDecimalInteger(candidate) {
			return BigIntExpr{Value: candidate, Line: tok.Line, Col: tok.Col}, nil
		}
	}
	if i, err := strconv.ParseInt(token, 10, 64); err == nil {
		return IntExpr{Value: i, Line: tok.Line, Col: tok.Col}, nil
	}
	if strings.ContainsAny(token, ".eE") {
		if f, err := strconv.ParseFloat(token, 64); err == nil {
			return FloatExpr{Value: f, Raw: token, Line: tok.Line, Col: tok.Col}, nil
		}
	}
	return SymbolExpr{Name: token, Line: tok.Line, Col: tok.Col}, nil
}

func parseTokenError(tok parsedSourceToken) error {
	return parseErrorAt(tok.Line, tok.Col, tok.Message)
}

func parseErrorAt(line, col int, msg string) error {
	return fmt.Errorf("parse error at %d:%d: %s", line, col, msg)
}

func tokenKindLabel(kind sourceTokenKind) string {
	switch kind {
	case sourceTokenListClose:
		return ")"
	case sourceTokenVectorClose:
		return "]"
	case sourceTokenMapClose:
		return "}"
	default:
		return "?"
	}
}
