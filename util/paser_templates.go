package util

import (
	"errors"
	"html/template"
	"io"
	"os"
)

func Templates(name string, fs template.FuncMap, files ...string) (result *template.Template) {
	result = template.New(name)
	for _, n := range files {
		b, err := os.ReadFile(n)
		if err != nil {
			panic(err)
		}
		result.Funcs(fs)
		result, err = result.Parse(string(b))
		if err != nil {
			panic(err)
		}
	}
	return
}

type TempleWrapper struct {
	template.Template
	current   *template.Template
	templates []*template.Template
}

func (w *TempleWrapper) Find(name string) (err error) {
	err = nil
	for _, t := range w.templates {
		if name == t.Name() {
			w.current = t
			return
		}
	}
	if w.current == nil {
		err = errors.New("template not found")
	}
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

func (w *TempleWrapper) Add(t *template.Template) {
	w.templates = append(w.templates, t)
}

func MakeWrapper() (result *TempleWrapper) {
	result = &TempleWrapper{current: template.New(""), templates: []*template.Template{}}
	return
}

func (w *TempleWrapper) Execute(wr io.Writer, data any) (err error) {
	if w.current == nil {
		err = errors.New("template not found")
		return
	}
	err = w.current.Execute(wr, data)
	return
}
