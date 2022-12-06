package main

import (
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"net/http"
	"os"

	"celus-ti.com.br/qualitec/util"
	"celus-ti.com.br/qualitec/web"
	"celus-ti.com.br/qualitec/web/controller"
	"github.com/Masterminds/sprig"
	humanize "github.com/dustin/go-humanize"
	_ "github.com/go-sql-driver/mysql"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"

	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

type configuration struct {
	DB       string
	DBEarth  string
	Site     string
	User     string
	Passwd   string
	Web_Root string
}

func main() {
	var err error

	// Loading configuration
	log.Println("Carregando configurações...")
	file, err := os.Open("conf.json")
	if err != nil {
		log.Fatal(err)
	}
	decoder := json.NewDecoder(file)
	cfg := configuration{}
	err = decoder.Decode(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Carregando templates...")
	t := template.New("celus").Funcs(sprig.FuncMap()).Funcs(template.FuncMap{
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
	})

	filepath.Walk(cfg.Web_Root+"/template", func(path string, f os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}
		if f.IsDir() {
			return nil
		}

		buffer, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}

		t.New(f.Name()).Parse(string(buffer))
		log.Println(f.Name())
		return nil
	})

	log.Printf("Conectando banco de dados %s\n", cfg.DB)
	db, _ := sqlx.Open("mysql", cfg.User+":"+cfg.Passwd+"@"+cfg.DB+"?parseTime=true&loc=UTC")
	db2, _ := sqlx.Open("mysql", cfg.User+":"+cfg.Passwd+"@"+cfg.DBEarth+"?parseTime=true&loc=UTC")
	// Por razões de segurança apenas
	cfg.Passwd = ""

	ctx := &controller.Context{
		DB:       db,
		DBEarth:  db2,
		Template: t,
		Store: sessions.NewCookieStore(
			securecookie.GenerateRandomKey(16),
			securecookie.GenerateRandomKey(16)),
		//Services: services,
	}
	ctx.Store.Options.MaxAge = 0

	routes := mux.NewRouter()
	ajax := routes.Headers("X-Requested-With", "XMLHttpRequest").Subrouter()
	ajax.Handle("/", controller.HandlerSession(ctx, controller.DeviceList))
	ajax.Handle("/dashboard", controller.HandlerSession(ctx, controller.DeviceList))
	ajax.Handle("/devices", controller.HandlerSession(ctx, controller.DeviceList))
	ajax.Handle("/device/{devflag}/{sub:(?:graph|history|hour|day|week|month)}", controller.HandlerSession(ctx, controller.DeviceView))
	ajax.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		controller.NotFound(ctx, w, r)
	})

	fileServer := http.FileServer(http.Dir(cfg.Web_Root + "/html"))

	routes.Handle("/favicon.ico", fileServer)
	routes.Handle("/css/{file}", fileServer)
	routes.Handle("/fonts/{file}", fileServer)
	routes.Handle("/images/{file}", fileServer)
	routes.Handle("/img/{file}", fileServer)
	routes.Handle("/js/{file}", fileServer)
	routes.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		controller.Login(ctx, w, r)
	})
	routes.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		controller.Logout(ctx, w, r)
	}).Methods("GET")
	routes.HandleFunc("/signin", func(w http.ResponseWriter, r *http.Request) {
		controller.SignIn(ctx, w, r)
	}).Methods("POST")

	routes.PathPrefix("/").Handler(controller.HandlerSession(ctx, controller.Base))

	webServer := handlers.CompressHandlerLevel(routes, gzip.DefaultCompression)

	log.Println("Iniciado HTTP e HTTPS...")
	if !prod {
		log.Println(http.ListenAndServe(":12345", webServer))
		return
	}

	go func() {
		err := http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			http.Redirect(w, req, "https://"+cfg.Site+req.RequestURI, http.StatusTemporaryRedirect)
		}))
		if err != nil {
			log.Println(err)
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
