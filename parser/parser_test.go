package parser_test

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"testing"

	embedparser "github.com/visualfc/goembed/parser"
)

func TestParser(t *testing.T) {
	bp, err := build.Import("github.com/visualfc/goembed", "", 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("import test embed", bp.TestEmbedPatterns, bp.TestEmbedPatternPos)
	fset := token.NewFileSet()
	var files []*ast.File
	for _, filename := range bp.TestGoFiles {
		f, err := parser.ParseFile(fset, filepath.Join(bp.Dir, filename), nil, parser.ParseComments)
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, f)
	}
	embed, err := embedparser.ParseEmbed(fset, files)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("parser test embed", embed)
	if !reflect.DeepEqual(bp.TestEmbedPatterns, embed.EmbedPatterns) {
		t.Fatal("EmbedPatterns error")
	}
	if len(bp.TestEmbedPatternPos) != len(embed.EmbedPatternPos) {
		t.Fatal("EmbedPatternPos len error")
	}
	for k, v := range bp.TestEmbedPatternPos {
		v2, ok := embed.EmbedPatternPos[k]
		if !ok {
			t.Fatal("not found", k)
		}
		if !reflect.DeepEqual(v, v2) {
			t.Fatal("not equal", v, v2)
		}
	}
}
