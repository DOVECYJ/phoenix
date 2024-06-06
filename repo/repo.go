package repo

import (
	"errors"

	"github.com/go-rel/rel"
)

var (
	driverMap = map[string]Driver{}
)

func Open(driverName, dataSourceName string) (rel.Repository, error) {
	driver, ok := driverMap[driverName]
	if !ok {
		return nil, errors.New("no driver available")
	}
	adapter, err := driver.Open(dataSourceName)
	if err != nil {
		return nil, err
	}
	return rel.New(adapter), nil
}

func MustOpen(driverName, dataSourceName string) rel.Repository {
	repo, err := Open(driverName, dataSourceName)
	if err != nil {
		panic(err)
	}
	return repo
}

type Driver interface {
	Open(dsn string) (rel.Adapter, error)
}

type DriverFunc func(dsn string) (rel.Adapter, error)

func (d DriverFunc) Open(dsn string) (rel.Adapter, error) {
	return d(dsn)
}

func Register(driverName string, driver Driver) {
	driverMap[driverName] = driver
}
