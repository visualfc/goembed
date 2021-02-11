// +build !go1.16

package resolve

import (
	"errors"
)

func ResolveEmbed(dir string, patterns []string) ([]string, error) {
	return nil, errors.New("unsupport")
}
