package goembed_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
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

func load(src string) ([]*goembed.File, error) {
	fset := token.NewFileSet()
	f, err := parserFile(fset, src)
	if err != nil {
		return nil, err
	}
	ems, err := parserEmbed(fset, f)
	if err != nil {
		return nil, err
	}
	r := goembed.NewResolve()
	wd, _ := os.Getwd()
	for _, em := range ems {
		_, err := r.Load(wd, fset, em)
		if err != nil {
			return nil, err
		}
	}
	return r.Files(), nil
}

func testError(src string, want string, t *testing.T) {
	_, err := load(src)
	if err == nil {
		t.Fatalf("must have error: %v", want)
	}
	if err.Error() != want {
		t.Fatalf("\nwant %v\nhave %v", want, err)
	}
}

type File struct {
	Name string
	Data string
}

func testLoad(src string, data []*File, t *testing.T) {
	files, err := load(src)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if len(files) != len(data) {
		t.Fatalf("load files error:\n want %v\n have %v", data, files)
	}
	for i, v := range files {
		if v.Name != data[i].Name {
			t.Fatalf("\nwant %v, have %v", data[i].Name, v.Name)
		}
		if string(v.Data) != data[i].Data {
			t.Fatalf("\nwant %v, have %v", data[i].Data, string(v.Data))
		}
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

func TestErrorMultipleFiles2(t *testing.T) {
	src := `package main

import _ "embed"

//go:embed testdata
var data string

func main() {
}
`
	testError(src, `./main.go:6:5: invalid go:embed: multiple files for type string`, t)
}

func TestLoadString(t *testing.T) {
	src := `package main

import _ "embed"

//go:embed testdata/data1.txt
var data string

func main() {
}
`
	testLoad(src, []*File{{"testdata/data1.txt", "hello data1"}}, t)
}

func TestLoadString2(t *testing.T) {
	src := `package main

import _ "embed"

//go:embed testdata/data1.txt
var data T
type T string

func main() {
}
`
	testLoad(src, []*File{{"testdata/data1.txt", "hello data1"}}, t)
}

func TestLoadString3(t *testing.T) {
	src := `package main

import _ "embed"

var (
	//go:embed testdata/data1.txt
	data string
)

func main() {
}
`
	testLoad(src, []*File{{"testdata/data1.txt", "hello data1"}}, t)
}

func TestLoadBytes(t *testing.T) {
	src := `package main

import _ "embed"

//go:embed testdata/data1.txt
var data []byte

func main() {
}
`
	testLoad(src, []*File{{"testdata/data1.txt", "hello data1"}}, t)
}

func TestLoadBytes2(t *testing.T) {
	src := `package main

import _ "embed"

//go:embed testdata/data1.txt
var data T
type T []byte

func main() {
}
`
	testLoad(src, []*File{{"testdata/data1.txt", "hello data1"}}, t)
}

func TestLoadBytes3(t *testing.T) {
	src := `package main

import _ "embed"

//go:embed testdata/data1.txt
var data []T
type T byte

func main() {
}
`
	testLoad(src, []*File{{"testdata/data1.txt", "hello data1"}}, t)
}

func TestLoadFiles(t *testing.T) {
	src := `package main

import "embed"

//go:embed testdata/data1.txt
//go:embed testdata/data2.txt
var data embed.FS

func main() {
}
`
	testLoad(src, []*File{{"testdata/data1.txt", "hello data1"}, {"testdata/data2.txt", "hello data2"}}, t)
}

func TestLoadFiles2(t *testing.T) {
	src := `package main

import em "embed"

//go:embed testdata/data1.txt
//go:embed testdata/data2.txt
var data em.FS

func main() {
}
`
	testLoad(src, []*File{{"testdata/data1.txt", "hello data1"}, {"testdata/data2.txt", "hello data2"}}, t)
}

func TestLoadFiles3(t *testing.T) {
	src := `package main

import . "embed"

//go:embed testdata/data1.txt
//go:embed testdata/data2.txt
var data FS

func main() {
}
`
	testLoad(src, []*File{{"testdata/data1.txt", "hello data1"}, {"testdata/data2.txt", "hello data2"}}, t)
}

func TestLoadOne(t *testing.T) {
	src := `package main

import _ "embed"

//go:embed testdata/one
var data string

func main() {
}
`
	testLoad(src, []*File{{"testdata/one/data.txt", "hello data"}}, t)
}

func TestLoadTwo(t *testing.T) {
	src := `package main

import _ "embed"

//go:embed testdata/two/data1.txt
var data1 string

//go:embed testdata/two/data2.txt
var data1 string

func main() {
}
`
	testLoad(src, []*File{{"testdata/two/data1.txt", "sub data1"}, {"testdata/two/data2.txt", "sub data2"}}, t)
}
