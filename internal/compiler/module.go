package compiler

import (
	"fmt"
	"strings"
)

// ModuleHeader is the first-form module map described in docs/modules.md.
type ModuleHeader struct {
	Namespace string
	Exports   []string
	Imports   []ImportSpec
	// GoExports maps a local export name to a host go-bind key (e.g. "burp/html").
	// Used by libraries/*.lib modules that re-export Go implementations.
	GoExports map[string]string
	Line      int
	Col       int
}

// ImportSpec is one :imports entry.
//
//	"chess.flag"
//	["chess.flag" :as "c"]
//	["chess.flag" :refer [move]]
//	["chess.flag" :as "c" :refer [move]]
type ImportSpec struct {
	Path  string
	As    string   // empty => use provider :namespace as prefix
	Refer []string // unqualified names in the importing module
	Line  int
	Col   int
}

// Module is a parsed source file with header and body forms.
type Module struct {
	Path   string
	Header ModuleHeader
	Forms  []Expr // body after header (or all forms if legacy)
	// LegacyNS is set when the file used (ns foo) instead of a header map.
	LegacyNS string
	// HasModuleHeader is true when the first form was a module header map.
	HasModuleHeader bool
}

// parseModuleHeader attempts to interpret expr as a module header map.
// ok is false when expr is not a map or is not a module header (no :namespace).
func parseModuleHeader(expr Expr) (ModuleHeader, bool, error) {
	m, ok := expr.(MapExpr)
	if !ok {
		return ModuleHeader{}, false, nil
	}
	if len(m.Entries)%2 != 0 {
		return ModuleHeader{}, false, exprError(expr, "module header map expects even number of forms")
	}

	header := ModuleHeader{Line: m.Line, Col: m.Col}
	seenNamespace := false

	for i := 0; i < len(m.Entries); i += 2 {
		keyExpr := m.Entries[i]
		valExpr := m.Entries[i+1]
		key, ok := keyExpr.(KeywordExpr)
		if !ok {
			return ModuleHeader{}, false, exprError(keyExpr, "module header keys must be keywords")
		}
		switch key.Name {
		case "namespace":
			seenNamespace = true
			switch v := valExpr.(type) {
			case StringExpr:
				header.Namespace = strings.TrimSpace(v.Value)
			case SymbolExpr:
				header.Namespace = strings.TrimSpace(v.Name)
			default:
				return ModuleHeader{}, false, exprError(valExpr, ":namespace must be a string or symbol")
			}
			if header.Namespace == "" {
				return ModuleHeader{}, false, exprError(valExpr, ":namespace cannot be empty")
			}
			if strings.Contains(header.Namespace, "/") {
				return ModuleHeader{}, false, exprError(valExpr, ":namespace must not contain '/'")
			}
		case "exports":
			names, err := parseSymbolVector(valExpr, ":exports")
			if err != nil {
				return ModuleHeader{}, false, err
			}
			header.Exports = names
		case "imports":
			specs, err := parseImportList(valExpr)
			if err != nil {
				return ModuleHeader{}, false, err
			}
			header.Imports = specs
		case "go-exports":
			binds, err := parseGoExportsMap(valExpr)
			if err != nil {
				return ModuleHeader{}, false, err
			}
			header.GoExports = binds
		default:
			return ModuleHeader{}, false, exprError(keyExpr, fmt.Sprintf("unknown module header key :%s", key.Name))
		}
	}

	if !seenNamespace {
		// A leading map without :namespace is ordinary program data, not a header.
		return ModuleHeader{}, false, nil
	}
	return header, true, nil
}

func parseImportList(expr Expr) ([]ImportSpec, error) {
	vec, ok := expr.(VectorExpr)
	if !ok {
		return nil, exprError(expr, ":imports must be a vector")
	}
	out := make([]ImportSpec, 0, len(vec.Elements))
	for _, item := range vec.Elements {
		spec, err := parseImportSpec(item)
		if err != nil {
			return nil, err
		}
		out = append(out, spec)
	}
	return out, nil
}

func parseImportSpec(expr Expr) (ImportSpec, error) {
	switch v := expr.(type) {
	case StringExpr:
		path := strings.TrimSpace(v.Value)
		if path == "" {
			return ImportSpec{}, exprError(expr, "import path cannot be empty")
		}
		return ImportSpec{Path: path, Line: v.Line, Col: v.Col}, nil
	case VectorExpr:
		if len(v.Elements) == 0 {
			return ImportSpec{}, exprError(expr, "import vector cannot be empty")
		}
		pathExpr, ok := v.Elements[0].(StringExpr)
		if !ok {
			return ImportSpec{}, exprError(v.Elements[0], "import vector must start with a path string")
		}
		path := strings.TrimSpace(pathExpr.Value)
		if path == "" {
			return ImportSpec{}, exprError(pathExpr, "import path cannot be empty")
		}
		spec := ImportSpec{Path: path, Line: v.Line, Col: v.Col}
		for i := 1; i < len(v.Elements); {
			kw, ok := v.Elements[i].(KeywordExpr)
			if !ok {
				return ImportSpec{}, exprError(v.Elements[i], "import options must be keyword/value pairs")
			}
			if i+1 >= len(v.Elements) {
				return ImportSpec{}, exprError(kw, fmt.Sprintf("import option :%s needs a value", kw.Name))
			}
			val := v.Elements[i+1]
			switch kw.Name {
			case "as":
				switch a := val.(type) {
				case StringExpr:
					spec.As = strings.TrimSpace(a.Value)
				case SymbolExpr:
					spec.As = strings.TrimSpace(a.Name)
				default:
					return ImportSpec{}, exprError(val, ":as must be a string or symbol")
				}
				if spec.As == "" {
					return ImportSpec{}, exprError(val, ":as cannot be empty")
				}
				if strings.Contains(spec.As, "/") {
					return ImportSpec{}, exprError(val, ":as must not contain '/'")
				}
			case "refer":
				names, err := parseSymbolVector(val, ":refer")
				if err != nil {
					return ImportSpec{}, err
				}
				spec.Refer = names
			default:
				return ImportSpec{}, exprError(kw, fmt.Sprintf("unknown import option :%s", kw.Name))
			}
			i += 2
		}
		return spec, nil
	default:
		return ImportSpec{}, exprError(expr, "import must be a string path or a vector")
	}
}

func parseGoExportsMap(expr Expr) (map[string]string, error) {
	m, ok := expr.(MapExpr)
	if !ok {
		return nil, exprError(expr, ":go-exports must be a map")
	}
	if len(m.Entries)%2 != 0 {
		return nil, exprError(expr, ":go-exports map expects even number of forms")
	}
	out := map[string]string{}
	for i := 0; i < len(m.Entries); i += 2 {
		var local string
		switch k := m.Entries[i].(type) {
		case SymbolExpr:
			local = k.Name
		case KeywordExpr:
			local = k.Name
		default:
			return nil, exprError(m.Entries[i], ":go-exports keys must be symbols or keywords")
		}
		if local == "" || strings.Contains(local, "/") {
			return nil, exprError(m.Entries[i], ":go-exports keys must be unqualified names")
		}
		val, ok := m.Entries[i+1].(StringExpr)
		if !ok {
			return nil, exprError(m.Entries[i+1], ":go-exports values must be strings (host bind keys)")
		}
		hostKey := strings.TrimSpace(val.Value)
		if hostKey == "" {
			return nil, exprError(m.Entries[i+1], ":go-exports value cannot be empty")
		}
		if _, exists := out[local]; exists {
			return nil, exprError(m.Entries[i], fmt.Sprintf("duplicate :go-exports key %q", local))
		}
		out[local] = hostKey
	}
	return out, nil
}

func parseSymbolVector(expr Expr, label string) ([]string, error) {
	vec, ok := expr.(VectorExpr)
	if !ok {
		return nil, exprError(expr, label+" must be a vector of symbols")
	}
	out := make([]string, 0, len(vec.Elements))
	seen := map[string]bool{}
	for _, item := range vec.Elements {
		sym, ok := item.(SymbolExpr)
		if !ok || sym.Name == "" {
			return nil, exprError(item, label+" entries must be symbols")
		}
		if strings.Contains(sym.Name, "/") {
			return nil, exprError(item, label+" entries must be unqualified symbols")
		}
		if seen[sym.Name] {
			return nil, exprError(item, fmt.Sprintf("duplicate %s symbol %q", label, sym.Name))
		}
		seen[sym.Name] = true
		out = append(out, sym.Name)
	}
	return out, nil
}

// splitQualifiedSymbol splits "prefix/name" into prefix and name.
// ok is false when name has no "/" or is malformed.
func splitQualifiedSymbol(name string) (prefix, local string, ok bool) {
	idx := strings.IndexByte(name, '/')
	if idx <= 0 || idx == len(name)-1 {
		return "", "", false
	}
	// Only the first slash separates namespace from local name.
	return name[:idx], name[idx+1:], true
}

// moduleGoIdent builds a unique Go identifier for a def in a module namespace.
// Legacy modules with an empty namespace keep the bare toGoIdentifier name.
func moduleGoIdent(namespace, localName string) (string, error) {
	local, err := toGoIdentifier(localName)
	if err != nil {
		return "", err
	}
	if namespace == "" {
		return local, nil
	}
	ns, err := toGoIdentifier(strings.ReplaceAll(namespace, ".", "_"))
	if err != nil {
		return "", fmt.Errorf("namespace %q: %w", namespace, err)
	}
	return ns + "__" + local, nil
}

// parseModuleFile parses source into a Module. path is stored for diagnostics
// and import resolution; it may be empty for string-only compiles.
func parseModuleFile(path, source string) (*Module, error) {
	ast, err := ParseFile(source)
	if err != nil {
		return nil, err
	}
	return parseModuleAST(path, ast)
}

func parseModuleTokenStream(path string, tokens <-chan ParseToken) (*Module, error) {
	ast, err := ParseTokenChannel(tokens)
	if err != nil {
		return nil, err
	}
	return parseModuleAST(path, ast)
}

func parseModuleAST(path string, ast FileAST) (*Module, error) {
	mod := &Module{Path: path, Forms: ast.Forms}
	if len(ast.Forms) == 0 {
		return mod, nil
	}

	header, isHeader, err := parseModuleHeader(ast.Forms[0])
	if err != nil {
		return nil, err
	}
	if isHeader {
		mod.Header = header
		mod.HasModuleHeader = true
		mod.Forms = ast.Forms[1:]
		return mod, nil
	}

	// Legacy (ns foo) as first form.
	if list, ok := ast.Forms[0].(ListExpr); ok && len(list.Elements) > 0 {
		if head, ok := list.Elements[0].(SymbolExpr); ok && head.Name == "ns" {
			if len(list.Elements) != 2 {
				return nil, exprError(list, "ns expects one namespace symbol")
			}
			name, ok := list.Elements[1].(SymbolExpr)
			if !ok || name.Name == "" {
				return nil, exprError(list.Elements[1], "namespace cannot be empty")
			}
			mod.LegacyNS = name.Name
			mod.Forms = ast.Forms[1:]
			return mod, nil
		}
	}

	return mod, nil
}
