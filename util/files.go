package util

import (
	"embed"
	"io/fs"
)

type files struct {
	e embed.FS
	n map[string]string
}

// Cria uma função que auxilia pesquisa pro um nome de arquivo
// sem precisar passar o caminho inteiro
func MakeFiles(value embed.FS) func(n string) string {
	result := &files{e: value}
	result.init()
	return func(n string) string {
		return result.N(n)
	}
}

func (f *files) init() *files {
	f.n = make(map[string]string)
	fs.WalkDir(f.e, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		f.n[d.Name()] = path
		return nil
	})
	return f
}

func (f files) N(n string) string {
	return f.n[n]
}
