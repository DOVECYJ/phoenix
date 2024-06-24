package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/DOVECYJ/phoenix/cmd"
	"github.com/azer/snakecase"
	"github.com/jinzhu/inflection"
	"github.com/urfave/cli/v2"
)

const migrationTemplate = `package migrations

import (
	"context"

	"github.com/go-rel/rel"
)

func MigrateCreate{{.Entity}}(schema *rel.Schema) {
	schema.CreateTable("{{.Table}}", func(t *rel.Table) {
		t.ID("id", rel.Primary(true))
		// other fields goes here
		{{- range .Columns}}
		{{- if and (ne .Type "") (ne .Name "")}}
		t.{{.Type}}("{{.Name}}")
		{{- end}}
		{{- end}}
		// timestamp
		t.DateTime("created_at")
		t.DateTime("updated_at")
	})

	schema.Do(func(ctx context.Context, r rel.Repository) error {
		// add seeds
		return nil
	})
}

func RollbackCreate{{.Entity}}(schema *rel.Schema) {
	schema.DropTable("{{.Table}}")
}
`

type migrationParam struct {
	Version  string `validate:"required"`
	Entity   string `validate:"required"` // entity name
	Table    string `validate:"required"` // table name
	Columns  []column
	filename string
	_created bool
}

func (p *migrationParam) bind(ctx *cli.Context, args ...string) {
	if len(args) != 2 || args[0] == "" || args[1] == "" {
		return
	}
	p.Version = time.Now().Format("20060102150405")
	p.Entity = args[1]
	p.Table = ctx.String("table")
	if p.Table == "" {
		p.Table = snakecase.SnakeCase(inflection.Plural(p.Entity))
	}
	fields := ctx.StringSlice("fields")
	for _, field := range fields {
		ss := strings.Split(field, ":")
		if c, err := newColumn(ss...); err == nil {
			p.Columns = append(p.Columns, c)
		}
	}
}

func (p *migrationParam) setMod(mod string) {
	// p.Mod = mod
}

func (p *migrationParam) setApp(app string) {
	// p.App = app
}

func (p *migrationParam) created() bool {
	return p._created
}

func (p *migrationParam) executeTemplate() error {
	p.filename = fmt.Sprintf("priv/repo/migrations/%s_create_%s.go", p.Version, p.Table)
	dir := filepath.Dir(p.filename)
	ds, err := os.ReadDir(dir)
	if err == nil {
		for i := range ds {
			if !ds[i].IsDir() && strings.HasSuffix(ds[i].Name(), p.filename[35:]) {
				fmt.Printf("skipted: %s\n", p.filename)
				return nil
			}
		}
	}
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	// execute template
	temp, err := template.New("migration").
		Funcs(template.FuncMap{
			"snake":  snakecase.SnakeCase,
			"plural": inflection.Plural,
		}).
		Parse(migrationTemplate)
	if err != nil {
		return err
	}
	if p._created, err = executeTemplate(p.filename, p, temp); err != nil {
		return err
	}
	return cmd.Cmd("go fmt " + p.filename).Run()
}

func (p *migrationParam) rollback() {
	if p._created {
		if err := os.Remove(p.filename); err == nil {
			fmt.Println("- removed:", p.filename)
			p._created = false
		}
	}
}

type column struct {
	Name string
	Type string
}

func newColumn(items ...string) (c column, err error) {
	switch len(items) {
	// case 1:
	// 内嵌字段
	case 2:
		c.Name = snakecase.SnakeCase(items[0])
		c.Type = getTypeFunc(items[1])
	case 3:
		c.Name = items[2]
		c.Type = getTypeFunc(items[1])
	default:
		err = errors.New("field args length error")
	}
	return
}

func getTypeFunc(t string) string {
	switch t {
	default:
		return ""
	case "bool":
		return "Bool"
	case "int":
		return "Int"
	case "float32":
		return "Float"
	case "float64":
		return "Float"
	case "string":
		return "String"
	case "time.Time":
		return "DateTime"
	}
}
