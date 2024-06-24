package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/DOVECYJ/phoenix/cmd"
	"github.com/urfave/cli/v2"
)

var migrateTemplate = `{%s, migrations.MigrateCreate%s, migrations.RollbackCreate%s},
    //{anchor(do not delete)}`

type migrateParam struct {
	Version   string `validate:"required"`
	Entity    string `validate:"required"` // entity name
	backup    bytes.Buffer
	_injected bool
}

func (p *migrateParam) bindMigration(m *migrationParam) {
	p.Version = m.Version
	p.Entity = m.Entity
}

func (p *migrateParam) bind(ctx *cli.Context, args ...string) {}

func (p *migrateParam) setMod(mod string) {
	// p.Mod = mod
}

func (p *migrateParam) setApp(app string) {
	// p.App = app
}

func (p *migrateParam) created() bool {
	return p._injected
}

func (p *migrateParam) executeTemplate() error {
	f, err := os.OpenFile("priv/repo/migrate.go", os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	tr := io.TeeReader(f, &p.backup)
	bs, err := io.ReadAll(tr)
	if err != nil {
		return err
	}

	replaced := false
	rem := []byte(`//\\`)
	if bytes.Contains(bs, rem) {
		bs = bytes.Replace(bs, rem, []byte{}, 1)
		replaced = true
	}
	rem = []byte("//{anchor(do not delete)}")
	if bytes.Contains(bs, rem) {
		insert := []byte(fmt.Sprintf(migrateTemplate, p.Version, p.Entity, p.Entity))
		bs = bytes.Replace(bs, rem, insert, 1)
		replaced = true
	}
	if !replaced {
		return nil
	}

	if err = f.Truncate(0); err != nil {
		return err
	}
	if _, err = f.WriteAt(bs, 0); err != nil {
		return err
	}
	p._injected = true
	fmt.Println("* inject:", "priv/repo/migrate.go")
	return cmd.Cmd("go fmt priv/repo/migrate.go").Run()
}

func (p *migrateParam) rollback() {
	if p._injected {
		f, err := os.OpenFile("priv/repo/migrate.go", os.O_WRONLY, os.ModePerm)
		if err != nil {
			return
		}
		defer f.Close()

		if _, err = io.Copy(f, &p.backup); err == nil {
			fmt.Println("- uninject:", "priv/repo/migrate.go")
			p._injected = false
		}
	}
}
