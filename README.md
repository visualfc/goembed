# embed
go1.16 embed util

```
package main

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"log"
	"path/filepath"

	"github.com/visualfc/goembed"
)

func main() {
	pkg, err := build.Import("github.com/visualfc/goembed", "", 0)
	if err != nil {
		log.Fatal(err)
	}
	fset := token.NewFileSet()
	var files []*ast.File
	for _, file := range pkg.TestGoFiles {
		f, err := parser.ParseFile(fset, filepath.Join(pkg.Dir, file), nil, 0)
		if err != nil {
			log.Fatal(err)
		}
		files = append(files, f)
	}
	ems := goembed.CheckEmbed(pkg.TestEmbedPatternPos, fset, files)
	r := goembed.NewResolve()
	for _, em := range ems {
		files, err := r.Load(pkg.Dir, em)
		if err != nil {
			log.Fatal("error load", em, err)
		}
		for _, f := range files {
			log.Println(f.Name, f.Data, f.Hash)
		}
	}
}
```
