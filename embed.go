package goembed

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"sort"
	"strings"

	embedparser "github.com/visualfc/goembed/parser"
)

// Kind is embed var type kind
type Kind int

const (
	EmbedUnknown Kind = iota
	EmbedBytes
	EmbedString
	EmbedFiles
	EmbedMaybeAlias // may be alias string or []byte
)

// Embed describes go:embed variable
type Embed struct {
	Name     string
	Kind     Kind
	Patterns []string
	Pos      token.Position
	Spec     *ast.ValueSpec
}

// embedPos is go:embed start postion
func (e *Embed) embedPos() (pos token.Position) {
	pos = e.Pos
	pos.Column -= 9
	return
}

type embedPattern struct {
	Patterns string
	Pos      token.Position
}

// CheckEmbed lookup go:embed vars for embedPatternPos
func CheckEmbed(embedPatternPos map[string][]token.Position, fset *token.FileSet, files []*ast.File) ([]*Embed, error) {
	if len(embedPatternPos) == 0 {
		return nil, nil
	}
	fmap := make(map[string]bool)
	var ep []*embedPattern
	for k, v := range embedPatternPos {
		for _, pos := range v {
			fmap[pos.Filename] = true
			ep = append(ep, &embedPattern{k, pos})
		}
	}
	sort.SliceStable(ep, func(i, j int) bool {
		n := strings.Compare(ep[i].Pos.Filename, ep[j].Pos.Filename)
		if n == 0 {
			return ep[i].Pos.Offset < ep[j].Pos.Offset
		}
		return n < 0
	})
	var eps []*Embed
	last := &Embed{Patterns: []string{ep[0].Patterns}, Pos: ep[0].Pos}
	eps = append(eps, last)
	for i := 1; i < len(ep); i++ {
		e := ep[i]
		if e.Pos.Filename == last.Pos.Filename &&
			(e.Pos.Line == last.Pos.Line || e.Pos.Line == last.Pos.Line+1) {
			last.Patterns = append(last.Patterns, e.Patterns)
			last.Pos = e.Pos
		} else {
			last = &Embed{Patterns: []string{e.Patterns}, Pos: e.Pos}
			eps = append(eps, last)
		}
	}
	for _, file := range files {
		if fmap[fset.Position(file.Package).Filename] {
			err := findEmbed(fset, file, eps)
			if err != nil {
				return nil, err
			}
		}
	}
	for _, e := range eps {
		if e.Spec == nil {
			return nil, fmt.Errorf("%v: misplaced go:embed directive", e.embedPos())
		}
	}
	return eps, nil
}

func checkIdent(v ast.Expr, name string) bool {
	if ident, ok := v.(*ast.Ident); ok && ident.Name == name {
		return true
	}
	return false
}

func embedKind(typ ast.Expr, importName string) Kind {
	switch v := typ.(type) {
	case *ast.Ident:
		switch v.Name {
		case "string":
			return EmbedString
		case "FS":
			if importName == "." {
				return EmbedFiles
			}
		}
		return EmbedMaybeAlias
	case *ast.ArrayType:
		if v.Len != nil {
			break
		}
		if ident, ok := v.Elt.(*ast.Ident); ok {
			if ident.Name == "byte" {
				return EmbedBytes
			}
			return EmbedMaybeAlias
		}
	case *ast.SelectorExpr:
		if checkIdent(v.X, importName) && checkIdent(v.Sel, "FS") {
			return EmbedFiles
		}
	}
	return EmbedUnknown
}

func findEmbed(fset *token.FileSet, file *ast.File, eps []*Embed) error {
	importName, err := embedparser.FindEmbedImportName(file)
	if err != nil {
		return err
	}
	for _, decl := range file.Decls {
		if d, ok := decl.(*ast.GenDecl); ok && d.Tok == token.VAR {
			for _, spec := range d.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				name := vs.Names[0]
				pos := fset.Position(name.NamePos)
				for _, e := range eps {
					if pos.Filename == e.Pos.Filename &&
						pos.Line == e.Pos.Line+1 {
						if len(vs.Names) != 1 {
							return fmt.Errorf("%v: go:embed cannot apply to multiple vars", e.embedPos())
						}
						if len(vs.Values) > 0 {
							return fmt.Errorf("%v: go:embed cannot apply to var with initializer", e.embedPos())
						}
						kind := embedKind(vs.Type, importName)
						if kind == EmbedUnknown {
							var buf bytes.Buffer
							printer.Fprint(&buf, fset, vs.Type)
							return fmt.Errorf("%v: go:embed cannot apply to var of type %v", pos, buf.String())
						}
						e.Name = name.Name
						e.Kind = kind
						e.Spec = vs
					}
				}
			}
		}
	}
	return nil
}
