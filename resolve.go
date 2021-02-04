package embed

type File struct {
	Name string
	Data []byte
	Hash [16]byte
}

type Resolve struct {
	data map[string]*File
}
