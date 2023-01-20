package main

import (
	"compress/gzip"
	"crypto/tls"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"strings"

	"net/http"
	"os"

	"celus-ti.com.br/qualitec/util"
	"celus-ti.com.br/qualitec/web"
	"celus-ti.com.br/qualitec/web/controller"
	"github.com/Masterminds/sprig"
	humanize "github.com/dustin/go-humanize"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/exp/maps"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"

	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
)

type configuration struct {
	DB       string
	DBEarth  string
	Site     string
	User     string
	Passwd   string
	Web_Root string
}

func nocacheHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "no-cache,must-revalidate")
		next.ServeHTTP(w, r)
	})
}

//go:embed web/template/*
var templatesHTML embed.FS

//go:embed web/html
var staticFiles embed.FS

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {

	// Loading configuration
	log.Println("Carregando configurações...")
	file := util.Check(os.Open("conf.json"))(log.Fatal)

	decoder := json.NewDecoder(file)
	cfg := configuration{}
	if err := decoder.Decode(&cfg); err != nil {
		log.Fatal(err)
	}

	log.Println("Carregando templates...")
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

	w := util.MakeWrapper()
	f := util.MakeFiles(templatesHTML)
	w.Add(util.Templates("pagination", funcs, f("pagination.html")))
	w.Add(util.Templates("login", funcs, f("base.html"), f("login.html")))
	w.Add(util.Templates("device-list", funcs, f("base.html"), f("device-list.html"), f("components.html")))
	w.Add(util.Templates("dashboard", funcs, f("base.html"), f("dashboard.html")))
	w.Add(util.Templates("history", funcs, f("base.html"), f("device-history.html"), f("device-component.html")))
	w.Add(util.Templates("graph", funcs, f("base.html"), f("device-graph.html"), f("device-component.html")))
	w.Add(util.Templates("not_found", funcs, f("base.html"), f("404.html")))

	log.Printf("Conectando banco de dados %s\n", cfg.DB)
	db := util.Check(sqlx.Open("mysql", cfg.User+":"+cfg.Passwd+"@"+cfg.DB+"?parseTime=true&loc=UTC"))(log.Fatal)
	db2 := util.Check(sqlx.Open("mysql", cfg.User+":"+cfg.Passwd+"@"+cfg.DBEarth+"?parseTime=true&loc=UTC"))(log.Fatal)

	// Por razões de segurança apenas
	cfg.Passwd = ""

	ctx := &controller.Context{
		DB:       db,
		DBEarth:  db2,
		Template: w,
		Store: sessions.NewCookieStore(
			securecookie.GenerateRandomKey(16),
			securecookie.GenerateRandomKey(16)),
	}
	ctx.Store.Options.MaxAge = 0

	routes := mux.NewRouter()
	routes.Handle("/", controller.HandlerSession(ctx, controller.DeviceList))
	routes.Handle("/dashboard", controller.HandlerSession(ctx, controller.DeviceList))
	routes.Handle("/devices", controller.HandlerSession(ctx, controller.DeviceList))
	routes.Handle("/device/{sub:(?:graph|history)}/{devflag}", controller.HandlerSession(ctx, controller.DeviceView))
	routes.Path("/device/graph/{devflag}").
		Queries("rangeBy", "(^$|day|week|month)").
		Handler(controller.HandlerSession(ctx, controller.DeviceView))

	if sub, err := fs.Sub(staticFiles, "web/html"); err == nil {
		fileServer := http.FileServer(http.FS(sub))
		routes.Handle("/favicon.ico", fileServer)
		routes.Handle("/css/{file}", fileServer)
		routes.Handle("/fonts/{file}", fileServer)
		routes.Handle("/images/{file}", fileServer)
		routes.Handle("/img/{file}", fileServer)
		routes.Handle("/js/{file}", fileServer)
	} else {
		log.Fatal(err)
	}

	routes.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		controller.Login(ctx, w, r)
	})
	routes.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		controller.Logout(ctx, w, r)
	}).Methods("GET")
	routes.HandleFunc("/signin", func(w http.ResponseWriter, r *http.Request) {
		controller.SignIn(ctx, w, r)
	}).Methods("POST")

	webServer := handlers.CompressHandlerLevel(routes, gzip.DefaultCompression)

	log.Println("Iniciado HTTP e HTTPS...")
	if !prod {
		log.Println(http.ListenAndServe(":12345", webServer))
		return
	}

	go func() {
		err := http.ListenAndServe(":80", nocacheHandler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			http.Redirect(w, req, fmt.Sprintf("https://%s%s", cfg.Site, req.RequestURI), http.StatusTemporaryRedirect)
		})))

		if err != nil {
			log.Fatal()
		}
	}()

	// Configurações HTTPS
	tlsConf := &tls.Config{
		MinVersion:               tls.VersionTLS10,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, // FOR HTTP/2 SUPPORT
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
		},
		NextProtos: []string{"h2"},
	}

	srv := &http.Server{
		Handler:   webServer,
		TLSConfig: tlsConf,
		//TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0), // Disable HTTP2
	}

	log.Println(srv.ListenAndServeTLS("portal.cdns.com.br.crt", "portal.cdns.com.br.key"))
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
