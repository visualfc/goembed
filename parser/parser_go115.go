//go:build !go1.16
// +build !go1.16

package parser

import "go/token"

func parseGoEmbed(args string, pos token.Position) ([]fileEmbed, error) {
	return nil, nil
}
