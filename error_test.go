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

func load(src string) error {
	fset := token.NewFileSet()
	f, err := parserFile(fset, src)
	if err != nil {
		return err
	}
	ems, err := parserEmbed(fset, f)
	if err != nil {
		return err
	}
	r := goembed.NewResolve()
	for _, em := range ems {
		_, err := r.Load(".", fset, em)
		if err != nil {
			return err
		}
	}
	return nil
}

func testError(src string, want string, t *testing.T) {
	err := load(src)
	if err == nil {
		t.Fatalf("must have error: %v", want)
	}
	if err.Error() != want {
		t.Fatalf("\nwant %v\nhave %v", want, err)
	}
}

func TestErrorMisImportEmbed(t *testing.T) {
	src := `package main

//go:embed testata/data1.txt
var data string

func main() {
}
	`
	testError(src, `./main.go:3:3: go:embed only allowed in Go files that import "embed"`, t)
}

func TestErrorMultipleVars(t *testing.T) {
	src := `package main

import _ "embed"

//go:embed testdata/data1.txt
var data, data2 string

func main() {
}
`
	testError(src, `./main.go:5:3: go:embed cannot apply to multiple vars`, t)
}

func TestErrorWithInitializer(t *testing.T) {
	src := `package main

import _ "embed"

//go:embed testdata/data1.txt
var data string = "hello"

func main() {
}
`
	testError(src, `./main.go:5:3: go:embed cannot apply to var with initializer`, t)
}

func TestErrorVarType(t *testing.T) {
	src := `package main

import _ "embed"

//go:embed testdata/data1.txt
var data [128]byte

func main() {
}
`
	testError(src, `./main.go:6:5: go:embed cannot apply to var of type [128]byte`, t)
}

func TestErrorVarType2(t *testing.T) {
	src := `package main

import _ "embed"

//go:embed testdata/data1.txt
var data [][]byte

func main() {
}
`
	testError(src, `./main.go:6:5: go:embed cannot apply to var of type [][]byte`, t)
}

func TestErrorVarType3(t *testing.T) {
	src := `package main

import _ "embed"

//go:embed testdata/data1.txt
var data map[int]int

func main() {
}
`
	testError(src, `./main.go:6:5: go:embed cannot apply to var of type map[int]int`, t)
}

func TestErrorMisplaced(t *testing.T) {
	src := `package main

import _ "embed"

//go:embed testdata/data1.txt
//var data string

func main() {
}
`
	testError(src, `./main.go:5:3: misplaced go:embed directive`, t)
}

func TestErrorMultipleFiles(t *testing.T) {
	src := `package main

import _ "embed"

//go:embed testdata/data1.txt testdata/data2.txt
var data string

func main() {
}
`
	testError(src, `./main.go:6:5: invalid go:embed: multiple files for type string`, t)
}
