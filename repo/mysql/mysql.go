package mysql

import (
	"github.com/DOVECYJ/phoenix/repo"
	"github.com/go-rel/mysql"
	_ "github.com/go-sql-driver/mysql"
)

func init() {
	repo.Register("mysql", repo.DriverFunc(mysql.Open))
}
