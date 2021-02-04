package embed

import (
	"embed"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"
)

//go:embed data/data1.txt
var data1 string

//go:embed data/data2.txt
var data2 []byte

//go:embed data
var fs embed.FS

func TestData(t *testing.T) {
	if data1 != "hello data1" {
		t.Fail()
	}
	if string(data2) != "hello data2" {
		t.Fail()
	}
	entrys, err := fs.ReadDir("data")
	if err != nil {
		t.Fatal(err)
	}
	if len(entrys) != 2 {
		t.Fail()
	}
}

func TestBuild(t *testing.T) {
	pkg, err := build.Import("github.com/visualfc/embed", "", 0)
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	var files []*ast.File
	for _, file := range pkg.TestGoFiles {
		f, err := parser.ParseFile(fset, filepath.Join(pkg.Dir, file), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, f)
	}
	ems := CheckEmbed(pkg.TestEmbedPatternPos, fset, files)
	if len(ems) != 3 {
		t.Fatal(ems)
	}
	r := NewResolve()
	for _, em := range ems {
		files, err := r.Load(pkg.Dir, em)
		if err != nil {
			t.Fatal("error load", em, err)
		}
		t.Log("resolve", em)
		for _, f := range files {
			t.Log(f.Name, string(f.Data), f.Hash)
		}
	}
}
