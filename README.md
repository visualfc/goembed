# goembed
goembed is Golang go:embed parse package

[![Go1.16](https://github.com/visualfc/goembed/workflows/Go1.16/badge.svg)](https://github.com/visualfc/goembed/actions/workflows/go116.yml)
[![Go1.17](https://github.com/visualfc/goembed/workflows/Go1.17/badge.svg)](https://github.com/visualfc/goembed/actions/workflows/go117.yml)
[![Go1.18](https://github.com/visualfc/goembed/workflows/Go1.18/badge.svg)](https://github.com/visualfc/goembed/actions/workflows/go118.yml)
[![Go1.19](https://github.com/visualfc/goembed/workflows/Go1.19/badge.svg)](https://github.com/visualfc/goembed/actions/workflows/go119.yml)


### demo
```
package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"path/filepath"

	"github.com/visualfc/goembed"
)

func main() {
	pkg, err := build.Import("github.com/visualfc/goembed", "", 0)
	if err != nil {
		panic(err)
	}
	fset := token.NewFileSet()
	var files []*ast.File
	for _, file := range pkg.TestGoFiles {
		f, err := parser.ParseFile(fset, filepath.Join(pkg.Dir, file), nil, 0)
		if err != nil {
			panic(err)
		}
		files = append(files, f)
	}
	ems,err := goembed.CheckEmbed(pkg.TestEmbedPatternPos, fset, files)
	if err != nil {
		panic(err)
	}
	r := goembed.NewResolve()
	for _, em := range ems {
		files, err := r.Load(pkg.Dir, fset, em)
		if err != nil {
			panic(err)
		}
		for _, f := range files {
			fmt.Println(f.Name, f.Data, f.Hash)
		}
	}
}
```
