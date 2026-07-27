package compiler

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type parser struct {
	src []rune
	pos int
}

func ParseFile(source string) (FileAST, error) {
	p := parser{src: []rune(source)}
	forms := make([]Expr, 0, 8)

	for {
		p.skipIgnorable()
		if p.done() {
			break
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

	return FileAST{Forms: forms}, nil
}

func (p *parser) readExpr() (Expr, error) {
	if p.done() {
		return nil, p.errorf("unexpected end of input")
	}

	switch p.peek() {
	case '(':
		return p.readList('(', ')')
	case '[':
		return p.readVector()
	case '{':
		return p.readMap()
	case '#':
		return p.readDispatch()
	case '\'':
		return p.readQuotedExpr()
	case '"':
		return p.readString()
	default:
		return p.readAtom()
	}
}

func (p *parser) readList(open, close rune) (Expr, error) {
	line, col := p.lineAndColumn()
	if p.next() != open {
		return nil, p.errorf("internal parser mismatch")
	}

	elements := make([]Expr, 0, 4)
	for {
		p.skipIgnorable()
		if p.done() {
			return nil, p.errorf("missing closing %q", string(close))
		}

		if p.peek() == close {
			p.next()
			if len(elements) > 0 {
				if sym, ok := elements[0].(SymbolExpr); ok && sym.Name == "comment" {
					return CommentExpr{}, nil
				}
			}
			return ListExpr{Elements: elements, Line: line, Col: col}, nil
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

func (p *parser) readVector() (Expr, error) {
	line, col := p.lineAndColumn()
	list, err := p.readList('[', ']')
	if err != nil {
		return nil, err
	}
	return VectorExpr{Elements: list.(ListExpr).Elements, Line: line, Col: col}, nil
}

func (p *parser) readMap() (Expr, error) {
	line, col := p.lineAndColumn()
	list, err := p.readList('{', '}')
	if err != nil {
		return nil, err
	}
	return MapExpr{Entries: list.(ListExpr).Elements, Line: line, Col: col}, nil
}

func (p *parser) readDispatch() (Expr, error) {
	line, col := p.lineAndColumn()
	if p.next() != '#' {
		return nil, p.errorf("internal parser mismatch")
	}
	if p.done() {
		return nil, p.errorf("unexpected end after #")
	}
	if p.peek() != '{' {
		if p.peek() == '(' {
			list, err := p.readList('(', ')')
			if err != nil {
				return nil, err
			}
			return HashFnExpr{Body: list, Line: line, Col: col}, nil
		}
		return nil, p.errorf("unsupported reader dispatch")
	}

	list, err := p.readList('{', '}')
	if err != nil {
		return nil, err
	}
	return SetExpr{Elements: list.(ListExpr).Elements, Line: line, Col: col}, nil
}

func (p *parser) readQuotedExpr() (Expr, error) {
	line, col := p.lineAndColumn()
	if p.next() != '\'' {
		return nil, p.errorf("internal parser mismatch")
	}
	p.skipIgnorable()
	if p.done() {
		return nil, p.errorf("unexpected end after quote")
	}
	quoted, err := p.readExpr()
	if err != nil {
		return nil, err
	}
	switch value := quoted.(type) {
	case SymbolExpr:
		return QuotedSymbolExpr{Name: value.Name, Line: line, Col: col}, nil
	case ListExpr:
		return QuotedListExpr{Elements: value.Elements, Line: line, Col: col}, nil
	default:
		return nil, p.errorf("quote currently supports symbols and lists")
	}
}

func (p *parser) readString() (Expr, error) {
	line, col := p.lineAndColumn()
	if p.pos+2 < len(p.src) && p.src[p.pos] == '"' && p.src[p.pos+1] == '"' && p.src[p.pos+2] == '"' {
		return p.readMultilineString(line, col)
	}

	start := p.pos
	p.next() // opening quote
	escaped := false

	for !p.done() {
		ch := p.next()
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == '"' {
			raw := string(p.src[start:p.pos])
			value, err := strconv.Unquote(raw)
			if err != nil {
				return nil, p.errorf("invalid string literal: %v", err)
			}
			return StringExpr{Value: value, Line: line, Col: col}, nil
		}
	}

	return nil, p.errorf("unterminated string literal")
}

func (p *parser) readMultilineString(line, col int) (Expr, error) {
	p.pos += 3
	start := p.pos
	for !p.done() {
		if p.pos+2 < len(p.src) && p.src[p.pos] == '"' && p.src[p.pos+1] == '"' && p.src[p.pos+2] == '"' {
			value := string(p.src[start:p.pos])
			p.pos += 3
			return StringExpr{Value: value, Line: line, Col: col}, nil
		}
		p.next()
	}
	return nil, p.errorf("unterminated string literal")
}

func (p *parser) readAtom() (Expr, error) {
	line, col := p.lineAndColumn()
	start := p.pos
	for !p.done() {
		ch := p.peek()
		if isDelimiter(ch) {
			break
		}
		p.next()
	}

	token := string(p.src[start:p.pos])
	if token == "" {
		return nil, p.errorf("expected expression")
	}

	if strings.HasPrefix(token, ":") && len(token) > 1 {
		return KeywordExpr{Name: token[1:], Line: line, Col: col}, nil
	}
	if strings.Count(token, "/") == 1 && !strings.HasPrefix(token, "/") && !strings.HasSuffix(token, "/") {
		parts := strings.SplitN(token, "/", 2)
		numerator, err := strconv.ParseInt(parts[0], 10, 64)
		if err == nil {
			denominator, err := strconv.ParseInt(parts[1], 10, 64)
			if err == nil {
				if denominator == 0 {
					return nil, p.errorf("ratio denominator cannot be zero")
				}
				return RatioExpr{Numerator: numerator, Denominator: denominator, Line: line, Col: col}, nil
			}
		}
	}
	if i, err := strconv.ParseInt(token, 10, 64); err == nil {
		return IntExpr{Value: i, Line: line, Col: col}, nil
	}

	if strings.ContainsAny(token, ".eE") {
		if f, err := strconv.ParseFloat(token, 64); err == nil {
			return FloatExpr{Value: f, Raw: token, Line: line, Col: col}, nil
		}
	}

	return SymbolExpr{Name: token, Line: line, Col: col}, nil
}

func (p *parser) skipIgnorable() {
	for !p.done() {
		ch := p.peek()
		switch {
		case unicode.IsSpace(ch), ch == ',':
			p.next()
		case ch == ';':
			p.skipComment()
		default:
			return
		}
	}
}

func (p *parser) skipComment() {
	for !p.done() {
		ch := p.next()
		if ch == '\n' {
			return
		}
	}
}

func (p *parser) done() bool {
	return p.pos >= len(p.src)
}

func (p *parser) peek() rune {
	return p.src[p.pos]
}

func (p *parser) next() rune {
	ch := p.src[p.pos]
	p.pos++
	return ch
}

func (p *parser) errorf(format string, args ...any) error {
	line, col := p.lineAndColumn()
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("parse error at %d:%d: %s", line, col, msg)
}

func (p *parser) lineAndColumn() (int, int) {
	line := 1
	col := 1
	for i := 0; i < p.pos && i < len(p.src); i++ {
		if p.src[i] == '\n' {
			line++
			col = 1
			continue
		}
		col++
	}
	return line, col
}

func isDelimiter(ch rune) bool {
	return unicode.IsSpace(ch) ||
		ch == ',' ||
		ch == ';' ||
		ch == '\'' ||
		ch == '(' || ch == ')' ||
		ch == '[' || ch == ']' ||
		ch == '{' || ch == '}' ||
		ch == '"'
}
