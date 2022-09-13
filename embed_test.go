//go:build go1.16
// +build go1.16

package goembed

import (
	"embed"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"
	"unsafe"
)

//go:embed testdata/data1.txt
var data1 string

var (
	//go:embed testdata/data2.txt
	data2 []byte

	//go:embed testdata
	fs embed.FS
)

type file struct {
	name string
	data string
	hash [16]byte
}

type myfs struct {
	files *[]file
}

func TestEmbed(t *testing.T) {
	if data1 != "hello data1" {
		t.Fail()
	}
	if string(data2) != "hello data2" {
		t.Fail()
	}
	files := *(*myfs)(unsafe.Pointer(&fs)).files
	for _, file := range files {
		t.Log(file.name, file.data, file.hash)
	}
}

func TestResolve(t *testing.T) {
	pkg, err := build.Import("github.com/visualfc/goembed", "", 0)
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
	ems, err := CheckEmbed(pkg.TestEmbedPatternPos, fset, files)
	if err != nil {
		t.Fatal(err)
	}
	r := NewResolve()
	for _, em := range ems {
		files, err := r.Load(pkg.Dir, em)
		if err != nil {
			t.Fatal("error load", em, err)
		}
		if em.Kind == EmbedFiles {
			files := BuildFS(files)
			for _, f := range files {
				t.Log(f.Name, string(f.Data), f.Hash)
			}
		}
	}
	t.Log("--------- files")
	for _, f := range r.Files() {
		t.Log(f.Name, string(f.Data), f.Hash)
	}
}

func TestBytesHex(t *testing.T) {
	data := []byte("\x68\x65\x6c\x6c\x6f\x20\x77\x6f\x72\x6c\x64")
	s := BytesToHex(data)
	if s != `\x68\x65\x6c\x6c\x6f\x20\x77\x6f\x72\x6c\x64` {
		t.Fatal(s)
	}
	if string(data) != "hello world" {
		t.Fail()
	}
}

func TestBytesList(t *testing.T) {
	data := []byte{104, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100}
	s := BytesToList(data)
	if s != `104,101,108,108,111,32,119,111,114,108,100` {
		t.Fatal(s)
	}
	if string(data) != "hello world" {
		t.Fail()
	}
}
