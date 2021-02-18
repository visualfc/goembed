package goembed

import (
	"crypto/sha256"
	"io/ioutil"
	"path"
	"path/filepath"
	"sort"

	"github.com/visualfc/goembed/resolve"
)

type File struct {
	Name string
	Data []byte
	Hash [16]byte // truncated SHA256 hash
	Err  error
}

var (
	data = []byte{104, 101, 108, 108, 111}
)

func init() {
}

type Resolve interface {
	Load(dir string, em *Embed) ([]*File, error)
	Files() []*File
}

type resolveFile struct {
	data map[string]*File
}

func NewResolve() Resolve {
	return &resolveFile{make(map[string]*File)}
}

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

func (r *resolveFile) Load(dir string, em *Embed) ([]*File, error) {
	list, err := resolve.ResolveEmbed(dir, em.Patterns)
	if err != nil {
		return nil, err
	}
	var files []*File
	for _, v := range list {
		fpath := filepath.Join(dir, v)
		f, ok := r.data[fpath]
		if !ok {
			data, err := ioutil.ReadFile(fpath)
			f = &File{
				Name: v,
				Data: data,
				Err:  err,
			}
			if len(data) > 0 {
				hash := sha256.Sum256(data)
				copy(f.Hash[:], hash[:16])
			}
			r.data[fpath] = f
		}
		files = append(files, f)
	}
	sort.Slice(files, func(i, j int) bool {
		return embedFileLess(files[i].Name, files[j].Name)
	})
	return files, nil
}

// func embedFileNameSplit(name string) (dir, elem string) {
// 	pos := strings.LastIndex(name, "/")
// 	if pos >= 0 {
// 		return name[:pos], name[pos+1:]
// 	}
// 	return name, ""
// }

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
