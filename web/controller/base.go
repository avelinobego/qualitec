package controller

import (
	"encoding/gob"
	"log"
	"net/http"
	"net/url"

	"celus-ti.com.br/qualitec/database/model"
	"celus-ti.com.br/qualitec/util"
	"celus-ti.com.br/qualitec/web"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
)

type HandlerFunc func(*Context, http.ResponseWriter, *http.Request) (int, error)

type Context struct {
	DB       *sqlx.DB
	DBEarth  *sqlx.DB
	Template *util.TempleWrapper
	Store    *sessions.CookieStore
	URL      *web.URLQuery
	User     *model.User
}

func init() {
	gob.Register(&model.User{})
	gob.Register([]model.Customer{})
}

func HandlerSession(c *Context, f HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)

		session, err := c.Store.Get(r, "session")
		if err != nil {
			log.Println(err)
		}

		user, ok := session.Values["user"].(*model.User)
		if !ok {
			http.Redirect(w, r, "/login?next="+url.QueryEscape(r.URL.RequestURI()), http.StatusFound)
			return
		}

		// Make a copy of context
		ctx := (*c)
		ctx.User = user
		ctx.URL = web.NewURLQuery(r.URL.Query(), r.URL.Path)
		status, err := f(&ctx, w, r)

		if status != http.StatusOK {
			switch status {
			case http.StatusNotFound:
				notFound(c, w, r)
			case http.StatusInternalServerError:
				if err == nil {
					http.Error(w, http.StatusText(status), http.StatusInternalServerError)
				} else {
					log.Println(err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			default:
				http.Error(w, http.StatusText(status), status)
			}
		}

	})
}

func notFound(c *Context, w http.ResponseWriter, r *http.Request) {
	data := struct {
		URLPath string
	}{URLPath: r.URL.Path}

	err := c.Template.ExecuteTemplate(w, "not_found", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func Login(c *Context, w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Next":   r.URL.Query().Get("next"),
		"Status": r.URL.Query().Get("status"),
	}

	err := c.Template.ExecuteTemplate(w, "login", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Logout(c *Context, w http.ResponseWriter, r *http.Request) {
	session, err := c.Store.Get(r, "session")
	if err == nil {
		session.Values["user"] = nil
	}
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func SignIn(c *Context, w http.ResponseWriter, r *http.Request) {
	validation := web.NewFormValidation(r)
	password := validation.RequiredString("password")
	email := validation.RequiredString("username")
	next := validation.RequiredString("next")
	if next == "" {
		next = "/"
	}

	if password == "" || email == "" {
		http.Redirect(w, r, "/login?status=1&next="+url.QueryEscape(next), http.StatusFound)
		return
	}

	user, err := model.UserByEmail(c.DB, email)
	if err != nil || !user.IsValidPassword(password) {
		http.Redirect(w, r, "/login?status=3&next="+url.QueryEscape(next), http.StatusFound)
		return
	}

	session, err := c.Store.New(r, "session")
	if err != nil {
		log.Println(err)
	}
	user.Mssalkjqw = ""
	session.Values["user"] = &user
	err = session.Save(r, w)
	if err != nil {
		log.Println(err)
	}

	http.Redirect(w, r, next, http.StatusFound)
}
