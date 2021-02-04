package embed

import (
	"crypto/sha256"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	"github.com/visualfc/embed/resolve"
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
	dirs := make(map[string]bool)
	for _, file := range files {
		dir := strings.Split(file.Name, "/")
		if len(dir) > 1 {
			dir = dir[:len(dir)-1]
			var paths string
			for _, v := range dir {
				paths += v + "/"
				dirs[paths] = true
			}
		}
	}
	for dir, _ := range dirs {
		files = append(files, &File{Name: dir})
	}
	sort.SliceStable(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})
	return files
}

func (r *resolveFile) Files() (files []*File) {
	for _, v := range r.data {
		files = append(files, v)
	}
	sort.SliceStable(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
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
	return files, nil
}
