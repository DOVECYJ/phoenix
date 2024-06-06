package sqlite3

import (
	"github.com/DOVECYJ/phoenix/repo"
	"github.com/go-rel/sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

func init() {
	repo.Register("mysql", repo.DriverFunc(sqlite3.Open))
}
