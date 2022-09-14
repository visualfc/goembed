//go:build go1.16
// +build go1.16

package parser

import (
	"go/token"
	_ "unsafe"
)

// parseGoEmbed parses the text following "//go:embed" to extract the glob patterns.
// It accepts unquoted space-separated patterns as well as double-quoted and back-quoted Go strings.
// This is based on a similar function in cmd/compile/internal/gc/noder.go;
// this version calculates position information as well.
//go:linkname parseGoEmbed go/build.parseGoEmbed
func parseGoEmbed(args string, pos token.Position) ([]fileEmbed, error)
