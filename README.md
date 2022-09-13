# embed
go1.16 go:embed parse util

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
		files, err := r.Load(pkg.Dir, em)
		if err != nil {
			panic(err)
		}
		for _, f := range files {
			fmt.Println(f.Name, f.Data, f.Hash)
		}
	}
}
```
