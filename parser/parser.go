//go:build go1.16
// +build go1.16

package parser

import (
	"fmt"
	"go/ast"
	"go/token"
	"sort"
	"strconv"
	"strings"
	_ "unsafe"
)

type Embed struct {
	EmbedPatterns   []string                    // patterns from ast.File
	EmbedPatternPos map[string][]token.Position // line information for EmbedPatterns
}

func ParseEmbed(fset *token.FileSet, files []*ast.File) (*Embed, error) {
	var embeds []fileEmbed
	for _, file := range files {
		ems, err := parseFile(fset, file)
		if err != nil {
			return nil, err
		}
		if len(ems) > 0 {
			embeds = append(embeds, ems...)
		}
	}
	embedMap := make(map[string][]token.Position)
	for _, emb := range embeds {
		embedMap[emb.pattern] = append(embedMap[emb.pattern], emb.pos)
	}
	return &Embed{embedPatterns(embedMap), embedMap}, nil
}

func parseFile(fset *token.FileSet, file *ast.File) ([]fileEmbed, error) {
	var embeds []fileEmbed
	for _, group := range file.Comments {
		for _, comment := range group.List {
			if strings.HasPrefix(comment.Text, "//go:embed ") {
				embs, err := parseGoEmbed(comment.Text[11:], fset.Position(comment.Slash+11))
				if err == nil {
					embeds = append(embeds, embs...)
				}
			}
		}
	}
	if len(embeds) == 0 {
		return nil, nil
	}
	hasEmbed, err := haveEmbedImport(file)
	if err != nil {
		return nil, err
	}
	if !hasEmbed {
		return nil, fmt.Errorf(`%v: go:embed only allowed in Go files that import "embed"`, embeds[0].pos)
	}
	return embeds, nil
}

func embedPatterns(m map[string][]token.Position) []string {
	all := make([]string, 0, len(m))
	for path := range m {
		all = append(all, path)
	}
	sort.Strings(all)
	return all
}

func haveEmbedImport(file *ast.File) (bool, error) {
	for _, decl := range file.Decls {
		d, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, dspec := range d.Specs {
			spec, ok := dspec.(*ast.ImportSpec)
			if !ok {
				continue
			}
			quoted := spec.Path.Value
			path, err := strconv.Unquote(quoted)
			if err != nil {
				return false, fmt.Errorf("parser returned invalid quoted string: <%s>", quoted)
			}
			if path == "embed" {
				return true, nil
			}
		}
	}
	return false, nil
}

type fileEmbed struct {
	pattern string
	pos     token.Position
}

// parseGoEmbed parses the text following "//go:embed" to extract the glob patterns.
// It accepts unquoted space-separated patterns as well as double-quoted and back-quoted Go strings.
// This is based on a similar function in cmd/compile/internal/gc/noder.go;
// this version calculates position information as well.
//go:linkname parseGoEmbed go/build.parseGoEmbed
func parseGoEmbed(args string, pos token.Position) ([]fileEmbed, error)
