package compiler

import "go/ast"

// Go predeclared identifiers that FLAG prelude functions may emit as package-level
// names. Those must be renamed in embedded *_flag_gen.go files so they do not
// shadow builtins used by the rest of package compiler (e.g. close(chan)).
var goPredeclaredNames = map[string]bool{
	"append": true, "cap": true, "clear": true, "close": true,
	"complex": true, "copy": true, "delete": true, "imag": true,
	"len": true, "make": true, "max": true, "min": true, "new": true,
	"panic": true, "print": true, "println": true, "real": true,
	"recover": true,
}

// RewriteGoPredeclaredShadows renames package-level decls whose names are Go
// predeclared identifiers (and all matching idents in the file) to prologue_<name>.
// Selector names (pkg.Name) are left unchanged.
func RewriteGoPredeclaredShadows(file *ast.File) {
	shadows := map[string]string{}
	for _, decl := range file.Decls {
		for _, name := range predeclaredDeclNames(decl) {
			if goPredeclaredNames[name] {
				shadows[name] = "prologue_" + name
			}
		}
	}
	if len(shadows) == 0 {
		return
	}
	ast.Inspect(file, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok {
			ast.Inspect(sel.X, func(n ast.Node) bool {
				id, ok := n.(*ast.Ident)
				if ok {
					if repl, hit := shadows[id.Name]; hit {
						id.Name = repl
					}
				}
				return true
			})
			return false
		}
		id, ok := n.(*ast.Ident)
		if !ok {
			return true
		}
		if repl, hit := shadows[id.Name]; hit {
			id.Name = repl
		}
		return true
	})
}

func predeclaredDeclNames(decl ast.Decl) []string {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		if d.Recv != nil {
			return nil
		}
		return []string{d.Name.Name}
	case *ast.GenDecl:
		names := make([]string, 0, len(d.Specs))
		for _, spec := range d.Specs {
			switch s := spec.(type) {
			case *ast.ValueSpec:
				for _, ident := range s.Names {
					names = append(names, ident.Name)
				}
			case *ast.TypeSpec:
				names = append(names, s.Name.Name)
			}
		}
		return names
	default:
		return nil
	}
}
