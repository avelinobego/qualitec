package util

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
)

type TempleWrapper struct {
	current    *template.Template
	templates  map[string]*template.Template
	fileSystem embed.FS
	functions  template.FuncMap
}

func MakeWrapper(fileSystem embed.FS, functions template.FuncMap) (result *TempleWrapper) {
	result = &TempleWrapper{current: nil,
		templates:  make(map[string]*template.Template),
		fileSystem: fileSystem,
		functions:  functions}
	return
}

func (w *TempleWrapper) Find(name string) (err error) {
	err = nil
	if found, ok := w.templates[name]; ok {
		w.current = found
		return
	}
	err = fmt.Errorf("template '%s' not found", name)
	return
}

func (w *TempleWrapper) ExecuteTemplate(wr io.Writer, name string, data any) (err error) {
	err = w.Find(name)
	if err != nil {
		return
	}
	err = w.current.ExecuteTemplate(wr, name, data)
	return
}

func (w *TempleWrapper) Add(name string, files ...string) {
	temp := template.New(name)
	temp.Funcs(w.functions)
	for _, n := range files {
		b, err := w.fileSystem.ReadFile(n)
		if err != nil {
			panic(err)
		}
		temp, err = temp.Parse(string(b))
		if err != nil {
			panic(err)
		}
		w.templates[name] = temp
	}

}

func (w *TempleWrapper) Execute(wr io.Writer, data any) (err error) {
	if w.current == nil {
		err = errors.New("template not found")
		return
	}
	err = w.current.Execute(wr, data)
	return
}
