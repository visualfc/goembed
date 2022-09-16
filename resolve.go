package goembed

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"go/printer"
	"go/token"
	"io/ioutil"
	"path"
	"path/filepath"
	"sort"

	"github.com/visualfc/goembed/resolve"
)

// File is embed data info
type File struct {
	Name string
	Data []byte
	Hash [16]byte // truncated SHA256 hash
}

// Resolve is load embed data interface
type Resolve interface {
	Load(dir string, fset *token.FileSet, em *Embed) ([]*File, error)
	Files() []*File
}

type resolveFile struct {
	data map[string]*File
}

// NewResolve create load embed data interface
func NewResolve() Resolve {
	return &resolveFile{make(map[string]*File)}
}

// BuildFS is build files to new files list with directory
func BuildFS(files []*File) []*File {
	have := make(map[string]bool)
	var list []*File
	for _, file := range files {
		if !have[file.Name] {
			have[file.Name] = true
			list = append(list, file)
		}
		for dir := path.Dir(file.Name); dir != "." && !have[dir]; dir = path.Dir(dir) {
			have[dir] = true
			list = append(list, &File{Name: dir + "/"})
		}
	}
	sort.Slice(list, func(i, j int) bool {
		return embedFileLess(list[i].Name, list[j].Name)
	})
	return list
}

func (r *resolveFile) Files() (files []*File) {
	for _, v := range r.data {
		files = append(files, v)
	}
	sort.Slice(files, func(i, j int) bool {
		return embedFileLess(files[i].Name, files[j].Name)
	})
	return
}

func (r *resolveFile) Load(dir string, fset *token.FileSet, em *Embed) ([]*File, error) {
	list, err := resolve.ResolveEmbed(dir, em.Patterns)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", em.Pos, err)
	}
	var files []*File
	for _, v := range list {
		fpath := filepath.Join(dir, v)
		f, ok := r.data[fpath]
		if !ok {
			data, err := ioutil.ReadFile(fpath)
			if err != nil {
				return nil, fmt.Errorf("%v: embed %v: %w", em.Pos, em.Patterns, err)
			}
			f = &File{
				Name: v,
				Data: data,
			}
			if len(data) > 0 {
				hash := sha256.Sum256(data)
				copy(f.Hash[:], hash[:16])
			}
			r.data[fpath] = f
		}
		files = append(files, f)
	}
	if em.Kind != EmbedFiles && len(files) > 1 {
		var buf bytes.Buffer
		printer.Fprint(&buf, fset, em.Spec.Type)
		return nil, fmt.Errorf("%v: invalid go:embed: multiple files for type %v", fset.Position(em.Spec.Names[0].NamePos), buf.String())
	}
	sort.Slice(files, func(i, j int) bool {
		return embedFileLess(files[i].Name, files[j].Name)
	})
	return files, nil
}

func embedFileNameSplit(name string) (dir, elem string, isDir bool) {
	if name[len(name)-1] == '/' {
		isDir = true
		name = name[:len(name)-1]
	}
	i := len(name) - 1
	for i >= 0 && name[i] != '/' {
		i--
	}
	if i < 0 {
		return ".", name, isDir
	}
	return name[:i], name[i+1:], isDir
}

// embedFileLess implements the sort order for a list of embedded files.
// See the comment inside ../../../../embed/embed.go's Files struct for rationale.
func embedFileLess(x, y string) bool {
	xdir, xelem, _ := embedFileNameSplit(x)
	ydir, yelem, _ := embedFileNameSplit(y)
	return xdir < ydir || xdir == ydir && xelem < yelem
}
