rem Comandos
go get -u -v -t github.com/kardianos/govendor
go get -u -v -t github.com/jteeuwen/go-bindata/...
go get -u -v -t golang.org/x/tools/cmd/goimports

govendor fetch github.com/go-sql-driver/mysql@=v1.4.1
govendor fetch github.com/jmoiron/sqlx@v1.2.0
govendor fetch golang.org/x/crypto/bcrypt
govendor fetch golang.org/x/crypto/blowfish
govendor fetch github.com/gorilla/securecookie
govendor fetch github.com/gorilla/mux
govendor fetch github.com/gorilla/sessions
govendor fetch github.com/gorilla/handlers
govendor fetch github.com/elazarl/go-bindata-assetfs

govendor fetch github.com/dustin/go-humanize

govendor sync github.com/rubenv/sql-migrate/...
govendor sync gopkg.in/gorp.v1
govendor fetch github.com/jmoiron/sqlx/reflectx
govendor fetch github.com/jinzhu/now
govendor fetch github.com/Masterminds/sprig