package fs

import (
	"os"
	"path/filepath"
)

type (
	FileInfo  = os.FileInfo
	PathError = os.PathError
	FileMode  = os.FileMode
)

const (
	ModeIrregular = os.ModeIrregular
	ModeDir       = os.ModeDir
)

var (
	ErrNotExist = os.ErrNotExist
	SkipDir     = filepath.SkipDir
)

// ValidPath reports whether the given path name
// is valid for use in a call to Open.
// Path names passed to open are unrooted, slash-separated
// sequences of path elements, like “x/y/z”.
// Path names must not contain a “.” or “..” or empty element,
// except for the special case that the root directory is named “.”.
//
// Paths are slash-separated on all systems, even Windows.
// Backslashes must not appear in path names.
func ValidPath(name string) bool {
	if name == "." {
		// special case
		return true
	}

	// Iterate over elements in name, checking each.
	for {
		i := 0
		for i < len(name) && name[i] != '/' {
			if name[i] == '\\' {
				return false
			}
			i++
		}
		elem := name[:i]
		if elem == "" || elem == "." || elem == ".." {
			return false
		}
		if i == len(name) {
			return true // reached clean ending
		}
		name = name[i+1:]
	}
}
