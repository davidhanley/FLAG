package runtime

import "unsafe"

type SymbolObject struct {
	Name      string
	IsKeyword bool
}

func NewSymbol(name string) Value {
	return Value{p: unsafe.Pointer(&SymbolObject{Name: name}), tag: TagSymbol}
}

func NewKeyword(name string) Value {
	return Value{p: unsafe.Pointer(&SymbolObject{Name: name, IsKeyword: true}), tag: TagSymbol}
}

func (v Value) SymbolObject() *SymbolObject {
	if v.tag != TagSymbol {
		panic("SymbolObject called on non-symbol Value")
	}
	if v.p == nil {
		panic("symbol Value does not contain symbol pointer")
	}
	return (*SymbolObject)(v.p)
}

func Symbol(arg any) Value {
	switch value := arg.(type) {
	case string:
		return NewSymbol(value)
	case Value:
		switch value.tag {
		case TagString:
			return NewSymbol(value.StringValue())
		case TagSymbol:
			return NewSymbol(value.SymbolObject().Name)
		default:
			panic("symbol expects string or symbol")
		}
	default:
		panic("symbol expects string or symbol")
	}
}

func Keyword(arg any) Value {
	switch value := arg.(type) {
	case string:
		return NewKeyword(value)
	case Value:
		switch value.tag {
		case TagString:
			return NewKeyword(value.StringValue())
		case TagSymbol:
			return NewKeyword(value.SymbolObject().Name)
		default:
			panic("keyword expects string, symbol, or keyword")
		}
	default:
		panic("keyword expects string, symbol, or keyword")
	}
}

func Name(arg any) string {
	switch value := arg.(type) {
	case string:
		return value
	case Value:
		if value.tag != TagSymbol {
			panic("name expects symbol or keyword")
		}
		return value.SymbolObject().Name
	case *SymbolObject:
		if value == nil {
			panic("name expects symbol or keyword")
		}
		return value.Name
	default:
		panic("name expects symbol or keyword")
	}
}
