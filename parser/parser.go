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
	"unicode"
	"unicode/utf8"
)

// EmbedPatterns is go:embed patterns and pos
type EmbedPatterns struct {
	Patterns   []string                    // patterns from ast.File
	PatternPos map[string][]token.Position // line information for Patterns
}

// ParseEmbed parser go:embed patterns from files
func ParseEmbed(fset *token.FileSet, files []*ast.File) (*EmbedPatterns, error) {
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
	if len(embeds) == 0 {
		return nil, nil
	}
	embedMap := make(map[string][]token.Position)
	for _, emb := range embeds {
		embedMap[emb.pattern] = append(embedMap[emb.pattern], emb.pos)
	}
	return &EmbedPatterns{embedPatterns(embedMap), embedMap}, nil
}

func parseFile(fset *token.FileSet, file *ast.File) ([]fileEmbed, error) {
	hasEmbed, err := haveEmbedImport(file)
	if err != nil {
		return nil, err
	}
	var embeds []fileEmbed
	for _, group := range file.Comments {
		for _, comment := range group.List {
			if strings.HasPrefix(comment.Text, "//go:embed ") {
				if !hasEmbed {
					return nil, fmt.Errorf(`%v: go:embed only allowed in Go files that import "embed"`, fset.Position(comment.Slash+2))
				}
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
	name, err := FindEmbedImportName(file)
	return name != "", err
}

// FindEmbedImportName is find embed package import name
func FindEmbedImportName(file *ast.File) (string, error) {
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
				return "", fmt.Errorf("parser returned invalid quoted string: <%s>", quoted)
			}
			if path == "embed" {
				if spec.Name != nil {
					return spec.Name.Name, nil
				}
				return "embed", nil
			}
		}
	}
	return "", nil
}

type fileEmbed struct {
	pattern string
	pos     token.Position
}

func stringCut(s, sep string) (before, after string, found bool) {
	if i := strings.Index(s, sep); i >= 0 {
		return s[:i], s[i+len(sep):], true
	}
	return s, "", false
}

// parseGoEmbed parses the text following "//go:embed" to extract the glob patterns.
// It accepts unquoted space-separated patterns as well as double-quoted and back-quoted Go strings.
// This is based on a similar function in cmd/compile/internal/gc/noder.go;
// this version calculates position information as well.
func parseGoEmbed(args string, pos token.Position) ([]fileEmbed, error) {
	trimBytes := func(n int) {
		pos.Offset += n
		pos.Column += utf8.RuneCountInString(args[:n])
		args = args[n:]
	}
	trimSpace := func() {
		trim := strings.TrimLeftFunc(args, unicode.IsSpace)
		trimBytes(len(args) - len(trim))
	}

	var list []fileEmbed
	for trimSpace(); args != ""; trimSpace() {
		var path string
		pathPos := pos
	Switch:
		switch args[0] {
		default:
			i := len(args)
			for j, c := range args {
				if unicode.IsSpace(c) {
					i = j
					break
				}
			}
			path = args[:i]
			trimBytes(i)

		case '`':
			var ok bool
			path, _, ok = stringCut(args[1:], "`")
			if !ok {
				return nil, fmt.Errorf("invalid quoted string in //go:embed: %s", args)
			}
			trimBytes(1 + len(path) + 1)

		case '"':
			i := 1
			for ; i < len(args); i++ {
				if args[i] == '\\' {
					i++
					continue
				}
				if args[i] == '"' {
					q, err := strconv.Unquote(args[:i+1])
					if err != nil {
						return nil, fmt.Errorf("invalid quoted string in //go:embed: %s", args[:i+1])
					}
					path = q
					trimBytes(i + 1)
					break Switch
				}
			}
			if i >= len(args) {
				return nil, fmt.Errorf("invalid quoted string in //go:embed: %s", args)
			}
		}

		if args != "" {
			r, _ := utf8.DecodeRuneInString(args)
			if !unicode.IsSpace(r) {
				return nil, fmt.Errorf("invalid quoted string in //go:embed: %s", args)
			}
		}
		list = append(list, fileEmbed{path, pathPos})
	}
	return list, nil
}
