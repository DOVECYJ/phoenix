package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/DOVECYJ/phoenix/cmd"
	"github.com/azer/snakecase"
	"github.com/urfave/cli/v2"
)

const (
	modelFieldTemplate = "{{- define \"fields\"}}" +
		"\tID        int `json:\"id\"`\n" +
		"\tCreatedAt time.Time `json:\"created_at\"`\n" +
		"\tUpdatedAt time.Time `json:\"updated_at\"`" +
		"{{end}}"

	modelTemplate = `package model

import (
	"time"

	"github.com/go-rel/changeset"
	"github.com/go-rel/changeset/params"
)

// The entity of {{.Entity}}
type {{.Entity}} struct {
	{{template "fields"}}
	{{- range .Fields}}
	{{.Name}} {{.Type}} {{with .Tag}}{{.}}{{end}}
	{{- end}}
}
{{if .Table}}
// Override table name to be '{{.Table}}'
func ({{.Entity}}) Table() string {
    return "{{.Table}}"
}
{{end}}
func Change{{.Entity}}(data *{{.Entity}}, params params.Params) *changeset.Changeset {
	ch := changeset.Cast(data, params, []string{
		{{- range .Fields}}
		"{{.Column}}",
		{{- end}}
	})
	changeset.ValidateRequired(ch, []string{
		{{- range .Fields}}
		"{{.Column}}",
		{{- end}}
	})
	return ch
}
`
)

type modelParam struct {
	Name     string `validate:"required"` // context
	App      string `validate:"-"`        // application
	Entity   string `validate:"required"`
	Table    string `validate:"-"` // rename table if need
	Fields   []field
	filename string
	_created bool
}

func (p *modelParam) bind(ctx *cli.Context, args ...string) {
	if len(args) != 2 || args[0] == "" || args[1] == "" {
		return
	}
	p.Name = args[0]
	p.Entity = args[1]
	p.App = ctx.String("app")
	p.Table = ctx.String("table")
	fields := ctx.StringSlice("fields")
	for _, field := range fields {
		ss := strings.Split(field, ":")
		if f, err := newField(ss...); err == nil {
			p.Fields = append(p.Fields, f)
		}
	}
}

func (p *modelParam) setMod(mod string) {
	// p.Mod = mod
}

func (p *modelParam) setApp(app string) {
	p.App = app
}

func (p *modelParam) created() bool {
	return p._created
}

func (p *modelParam) executeTemplate() error {
	entity := snakecase.SnakeCase(p.Entity)
	p.filename = fmt.Sprintf("lib/%s/%s/model/%s.go", p.App, p.Name, entity)
	if err := os.MkdirAll(filepath.Dir(p.filename), os.ModePerm); err != nil {
		return err
	}
	// execute template
	temp, err := template.New("model").Parse(modelTemplate)
	if err != nil {
		return err
	}
	temp, err = temp.Parse(modelFieldTemplate)
	if err != nil {
		return err
	}
	if p._created, err = executeTemplate(p.filename, p, temp); err != nil {
		return err
	}
	return cmd.Cmd("go fmt " + p.filename).Run()
}

func (p *modelParam) rollback() {
	if p._created {
		os.Remove(p.filename)
		fmt.Println("- removed:", p.filename)
	}
}

type field struct {
	Name   string `validate:"required"` // struct field name in golang
	Type   string `validate:"-"`        // struct field type in golang
	Column string `validate:"required"` // column name in database
	Tag    tag
}

func newField(args ...string) (f field, err error) {
	switch len(args) {
	case 1:
		f.Name = args[0]
		f.Column = snakecase.SnakeCase(args[0])
		f.Tag.add("json", f.Column)
	case 2:
		f.Name = args[0]
		f.Type = args[1]
		f.Column = snakecase.SnakeCase(args[0])
		f.Tag.add("json", f.Column)
	case 3:
		f.Name = args[0]
		f.Type = args[1]
		f.Column = args[2]
		f.Tag.add("json", snakecase.SnakeCase(f.Name))
		f.Tag.add("db", f.Column)
	default:
		err = errors.New("field args length error")
	}
	return
}

type tag []string

func (t *tag) add(name string, value ...string) {
	if len(value) == 0 {
		return
	}
	*t = append(*t, fmt.Sprintf("%s:\"%s\"", name, strings.Join(value, ",")))
}

func (t tag) String() string {
	if len(t) == 0 {
		return ""
	}
	return "`" + strings.Join(t, " ") + "`"
}
