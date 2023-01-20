package main_test

import (
	"bytes"
	"embed"
	"testing"

	"celus-ti.com.br/qualitec/util"
)

//go:embed web/template/*
var templatesHTML embed.FS

func TestTemplate(t *testing.T) {

	w := util.MakeWrapper()
	f := util.MakeFiles(templatesHTML)

	w.Add(util.Templates("login", nil, f("base.html"), f("login.html")))
	w.Add(util.Templates("not_found", nil, f("base.html"), f("404.html")))

	b := new(bytes.Buffer)
	err := w.ExecuteTemplate(b, "not_found", nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(b.String())

	// b.Reset()
	// err = base.ExecuteTemplate(b, "not_found", nil)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// t.Log(b.String())
}
