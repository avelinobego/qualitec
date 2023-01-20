package controller_test

import (
	"fmt"
	"net/url"
	"testing"

	"celus-ti.com.br/qualitec/web"
	"celus-ti.com.br/qualitec/web/controller"
)

type Struct struct {
	URL  *web.URLQuery `json:"-"`
	Nome string
}

func TestEncode(t *testing.T) {
	v := make(url.Values)
	v["Nome"] = []string{"Bego"}
	u := web.NewURLQuery(v, "/teste")
	u.Clients = []int16{1, 2, 3, 4, 5}

	s := &Struct{Nome: "Avelino bego", URL: u}

	e := controller.EncodeBase64(s)
	t.Log(e)
}

func TestDecode(t *testing.T) {
	s := &Struct{}
	e := "eyJOb21lIjoiQXZlbGlubyBiZWdvIn0="
	controller.DecodeBase64(e, s)
	fmt.Printf("%v\n", s)
	fmt.Printf("%v\n", s.URL.Encode())
}
