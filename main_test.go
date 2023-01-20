package main_test

import (
	"embed"
	"html/template"
	"strings"
	"testing"

	"celus-ti.com.br/qualitec/util"
	"celus-ti.com.br/qualitec/web"
	"github.com/Masterminds/sprig"
	humanize "github.com/dustin/go-humanize"
	"golang.org/x/exp/maps"
)

//go:embed web/template/*
var templatesHTML embed.FS

func TestTemplate(t *testing.T) {

	funcs := template.FuncMap{
		"Round":            util.RoundPlus,
		"CommafWithDigits": humanize.CommafWithDigits,
		"RangeStruct":      util.RangeStructer,
		"NewVar":           newVar,
		"SetVar":           setVar,
		"GetVar":           getVar,
		"Dict":             web.Dict,
		"StrLeft":          util.StrLeft,
		"FormatByte":       util.FormatBytes1024,
		"FormatCEP":        util.FormatCEP,
		"FormatPhone":      util.FormatPhone,
		"FormatTime":       util.FormatTime,
		"FormatTimeH":      util.FormatTimeH,
		"FormatDate":       util.FormatDate,
		"FormatCurrency":   util.FormatCurrency,
		"MulFloat64": func(f1, f2 float64) float64 {
			return f1 * f2
		},
		"FormatFloat": func(f float64, d int) string {
			if d > 0 {
				return humanize.FormatFloat("###,###."+strings.Repeat("#", d), f)
			} else {
				return humanize.FormatFloat("###,###", f)
			}
		},
		"Uint8ToInt": func(a uint8) int {
			return int(a)
		},
		"Percentual": func(a, b int) int {
			return int(float64(float64(a) / float64(b) * 100))
		},
		"Minus": func(a, b int) int {
			return a - b
		},
	}

	maps.Copy(funcs, sprig.FuncMap())

	w := util.MakeWrapper(templatesHTML, funcs)
	w.Add("pagination",
		"web/template/pagination.html")
	w.Add("login",
		"web/template/base.html",
		"web/template/login.html")
	w.Add("device-list",
		"web/template/base.html",
		"web/template/device-list.html",
		"web/template/components.html")
	w.Add("dashboard",
		"web/template/base.html",
		"web/template/dashboard.html")
	w.Add("history",
		"web/template/base.html",
		"web/template/device-history.html",
		"web/template/device-component.html")
	w.Add("graph",
		"web/template/base.html",
		"web/template/device-graph.html",
		"web/template/device-component.html")
	w.Add("not_found",
		"web/template/base.html",
		"web/template/404.html")

	b := new(strings.Builder)
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

func newVar(v interface{}) (*interface{}, error) {
	x := interface{}(v)
	return &x, nil
}

func setVar(x *interface{}, v interface{}) (string, error) {
	*x = v
	return "", nil
}

func getVar(x *interface{}) (interface{}, error) {
	return *x, nil
}
