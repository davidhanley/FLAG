package runtime

type SymbolObject struct {
	Name      string
	IsKeyword bool
}

type symbolImpl struct {
	defaultImpl
	obj *SymbolObject
}

func NewSymbol(name string) Value {
	return Value{impl: &symbolImpl{obj: &SymbolObject{Name: name}}}
}

func NewKeyword(name string) Value {
	return Value{impl: &symbolImpl{obj: &SymbolObject{Name: name, IsKeyword: true}}}
}

func (v Value) SymbolObject() *SymbolObject {
	return asImpl[*symbolImpl](v, "SymbolObject called on non-symbol Value").obj
}

func Symbol(arg any) Value {
	switch value := arg.(type) {
	case string:
		return NewSymbol(value)
	case Value:
		if s, ok := value.impl.asIdent(value); ok {
			return NewSymbol(s)
		}
		panic("symbol expects string or symbol")
	default:
		panic("symbol expects string or symbol")
	}
}

func Keyword(arg any) Value {
	switch value := arg.(type) {
	case string:
		return NewKeyword(value)
	case Value:
		if s, ok := value.impl.asIdent(value); ok {
			return NewKeyword(s)
		}
		panic("keyword expects string, symbol, or keyword")
	default:
		panic("keyword expects string, symbol, or keyword")
	}
}

func Name(arg any) string {
	switch value := arg.(type) {
	case string:
		return value
	case Value:
		return value.impl.name(value)
	case *SymbolObject:
		if value == nil {
			panic("name expects symbol or keyword")
		}
		return value.Name
	default:
		panic("name expects symbol or keyword")
	}
}

func (symbolImpl) hashImpl(v Value) uint64 {
	symbol := v.SymbolObject()
	seed := hashSeedSymbol
	if symbol.IsKeyword {
		seed = hashSeedKeyword
	}
	return mix(seed, hashBytes([]byte(symbol.Name)))
}
func (symbolImpl) eq(v Value, other Value) bool {
	if !isType[*symbolImpl](other) {
		return false
	}
	a := v.SymbolObject()
	b := other.SymbolObject()
	return a.IsKeyword == b.IsKeyword && a.Name == b.Name
}
func (symbolImpl) typ(v Value) Value {
	if v.SymbolObject().IsKeyword {
		return typeKeyword
	}
	return typeSymbol
}
func (symbolImpl) String(v Value) string {
	symbol := v.SymbolObject()
	if symbol.IsKeyword {
		return ":" + symbol.Name
	}
	return symbol.Name
}
func (symbolImpl) asString(v Value) string {
	return v.SymbolObject().Name
}
func (symbolImpl) asIdent(v Value) (string, bool) {
	return v.SymbolObject().Name, true
}
func (symbolImpl) name(v Value) string {
	return v.SymbolObject().Name
}
func (symbolImpl) toNative(v Value) any {
	symbol := v.SymbolObject()
	if symbol.IsKeyword {
		return ":" + symbol.Name
	}
	return symbol.Name
}
func (symbolImpl) toJSON(v Value) (any, error) {
	return symbolImpl{}.toNative(v), nil
}
func (symbolImpl) apply(v Value, args ...Value) Value {
	if !v.SymbolObject().IsKeyword {
		panic("call expects function, map, set, or keyword Value")
	}
	if len(args) != 1 && len(args) != 2 {
		panic("keyword invocation expects map and optional default")
	}
	if len(args) == 1 {
		return Get(args[0], v)
	}
	return Get(args[0], v, args[1])
}
