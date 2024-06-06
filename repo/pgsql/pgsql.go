package pgsql

import (
	"github.com/DOVECYJ/phoenix/repo"
	"github.com/go-rel/postgres"
	_ "github.com/lib/pq"
)

func init() {
	repo.Register("mysql", repo.DriverFunc(postgres.Open))
}
