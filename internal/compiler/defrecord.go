package compiler

import (
	"fmt"
	"go/token"
	"strings"
	"unicode"
)

type recordField struct {
	flagName string
	goName   string
	goType   string
	param    string
}

func compileDefrecord(form ListExpr, ctx *compileContext) (string, []functionDef, []varDef, error) {
	if len(form.Elements) != 3 {
		return "", nil, nil, exprError(form, "defrecord expects a name and field vector")
	}

	nameExpr, ok := unwrapMetaExpr(form.Elements[1]).(SymbolExpr)
	if !ok || nameExpr.Name == "" {
		return "", nil, nil, exprError(form.Elements[1], "defrecord name must be a symbol")
	}
	fieldsExpr, ok := form.Elements[2].(VectorExpr)
	if !ok {
		return "", nil, nil, exprError(form.Elements[2], "defrecord fields must be a vector")
	}

	typeName, err := toGoIdentifier(nameExpr.Name)
	if err != nil {
		return "", nil, nil, err
	}
	if ctx.recordTypes == nil {
		ctx.recordTypes = map[string]string{}
	}
	if prev, exists := ctx.recordTypes[typeName]; exists {
		return "", nil, nil, exprError(form, fmt.Sprintf("record %q already defined as %s", nameExpr.Name, prev))
	}
	ctx.recordTypes[typeName] = nameExpr.Name

	fields := make([]recordField, 0, len(fieldsExpr.Elements))
	params := make([]string, 0, len(fieldsExpr.Elements))
	for _, fieldExpr := range fieldsExpr.Elements {
		field, err := parseRecordField(fieldExpr)
		if err != nil {
			return "", nil, nil, exprError(fieldExpr, err.Error())
		}
		fields = append(fields, field)
		params = append(params, field.param)
	}

	var typeDecl strings.Builder
	fmt.Fprintf(&typeDecl, "type %s struct {\n", typeName)
	for _, field := range fields {
		fmt.Fprintf(&typeDecl, "\t%s %s `flag:%q`\n", field.goName, field.goType, field.flagName)
	}
	typeDecl.WriteString("}\n")

	ctorGoName, err := moduleGoIdent(ctx.namespace, "->"+nameExpr.Name)
	if err != nil {
		return "", nil, nil, err
	}
	mapCtorGoName, err := moduleGoIdent(ctx.namespace, "map->"+nameExpr.Name)
	if err != nil {
		return "", nil, nil, err
	}

	ctorInits := make([]string, 0, len(fields))
	mapInits := make([]string, 0, len(fields))
	for _, field := range fields {
		ctorInits = append(ctorInits, fmt.Sprintf("%s: %s", field.goName, unboxRecordField(field.param, field.goType)))
		mapValue := fmt.Sprintf("%s.Get(m, %s)", runtimeAlias, ctx.keywordCode(field.flagName))
		mapInits = append(mapInits, fmt.Sprintf("%s: %s", field.goName, unboxRecordField(mapValue, field.goType)))
	}

	ctor := functionDef{
		flagName:     "->" + nameExpr.Name,
		goName:       ctorGoName,
		variadicName: ctorGoName + "_variadic",
		arityName:    fmt.Sprintf("%s_arity_%d", ctorGoName, len(fields)),
		params:       params,
		body:         fmt.Sprintf("%s.NewRecord(%s{%s})", runtimeAlias, typeName, strings.Join(ctorInits, ", ")),
	}
	mapCtor := functionDef{
		flagName:     "map->" + nameExpr.Name,
		goName:       mapCtorGoName,
		variadicName: mapCtorGoName + "_variadic",
		arityName:    mapCtorGoName + "_arity_1",
		params:       []string{"m"},
		body:         fmt.Sprintf("%s.NewRecord(%s{%s})", runtimeAlias, typeName, strings.Join(mapInits, ", ")),
	}

	ctors := []functionDef{ctor, mapCtor}
	vars := make([]varDef, 0, 2)
	for i := range ctors {
		def := ctors[i]
		if err := ctx.bindModuleName(def.flagName, def.goName); err != nil {
			return "", nil, nil, err
		}
		ctx.functions[def.goName] = def
		ctx.globals[def.goName] = exprKindValue
		vars = append(vars, varDef{
			flagName: def.flagName,
			goName:   def.goName,
			expr:     fmt.Sprintf("%s.NewFunction(%s)", runtimeAlias, def.variadicName),
		})
	}

	return typeDecl.String(), ctors, vars, nil
}

func parseRecordField(expr Expr) (recordField, error) {
	hint := ""
	target := expr
	if meta, ok := expr.(MetaExpr); ok {
		target = meta.Target
		switch typed := meta.Meta.(type) {
		case SymbolExpr:
			hint = typed.Name
		default:
			return recordField{}, fmt.Errorf("defrecord field type hint must be a symbol")
		}
	}
	name, ok := target.(SymbolExpr)
	if !ok || name.Name == "" {
		return recordField{}, fmt.Errorf("defrecord fields must be symbols")
	}
	goType := "flagrt.Value"
	if hint != "" {
		mapped := typeHintToGoType(hint)
		if mapped == "" {
			return recordField{}, fmt.Errorf("unsupported defrecord field type %q", hint)
		}
		goType = mapped
	}
	param, err := toGoIdentifier(name.Name)
	if err != nil {
		return recordField{}, err
	}
	if isReservedGoIdent(param) {
		param += "_"
	}
	return recordField{
		flagName: name.Name,
		goName:   exportedGoName(name.Name),
		goType:   goType,
		param:    param,
	}, nil
}

func unboxRecordField(expr, goType string) string {
	switch goType {
	case "string":
		return runtimeAlias + ".RequireString(" + expr + ")"
	case "int64":
		return runtimeAlias + ".RequireLong(" + expr + ")"
	case "float64":
		return runtimeAlias + ".RequireDouble(" + expr + ")"
	case "bool":
		return runtimeAlias + ".RequireBool(" + expr + ")"
	default:
		return expr
	}
}

func isReservedGoIdent(name string) bool {
	if token.Lookup(name).IsKeyword() {
		return true
	}
	switch name {
	case "any", "append", "bool", "byte", "cap", "close", "complex", "copy", "delete",
		"error", "false", "float32", "float64", "imag", "int", "int8", "int16", "int32", "int64",
		"iota", "len", "make", "max", "min", "new", "nil", "panic", "print", "println",
		"real", "recover", "rune", "string", "true", "uint", "uint8", "uint16", "uint32", "uint64",
		"uintptr":
		return true
	default:
		return false
	}
}

func exportedGoName(flagName string) string {
	parts := strings.Split(flagName, "-")
	var b strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		if part == "eof" {
			b.WriteString("EOF")
			continue
		}
		r := []rune(part)
		b.WriteRune(unicode.ToUpper(r[0]))
		b.WriteString(string(r[1:]))
	}
	return b.String()
}
