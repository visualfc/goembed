package goembed_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/visualfc/goembed"

	embedparser "github.com/visualfc/goembed/parser"
)

func parserFile(fset *token.FileSet, src string) (*ast.File, error) {
	file, err := parser.ParseFile(fset, "./main.go", src, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func parserEmbed(fset *token.FileSet, f *ast.File) ([]*goembed.Embed, error) {
	eps, err := embedparser.ParseEmbed(fset, []*ast.File{f})
	if err != nil {
		return nil, err
	}
	return goembed.CheckEmbed(eps.PatternPos, fset, []*ast.File{f})
}

func testError(src string, want string, t *testing.T) {
	fset := token.NewFileSet()
	f, err := parserFile(fset, src)
	if err != nil {
		t.Fatal(err)
	}
	_, err = parserEmbed(fset, f)
	if err == nil {
		t.Fatalf("must have error: %v", want)
	}
	if err.Error() != want {
		t.Fatalf("\nwant %v\nhave %v", want, err)
	}
}

func TestMisImportEmbed(t *testing.T) {
	src := `package main

//go:embed testata/data1.txt
var data string

func main() {
}
	`
	testError(src, `./main.go:3:3: go:embed only allowed in Go files that import "embed"`, t)
}
