package database

import "database/sql"

type Get interface {
	Get(interface{}, string, ...interface{}) error
}

type Select interface {
	Select(interface{}, string, ...interface{}) error
}

type Exec interface {
	Begin() (*sql.Tx, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type NamedExec interface {
	NamedExec(query string, arg interface{}) (sql.Result, error)
}

type DB interface {
	Get
	Select
	Exec
	NamedExec
}
