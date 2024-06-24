package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/azer/snakecase"
	"github.com/urfave/cli/v2"
)

const editHtmlTemplate = `
{{- $entity := lower .Entity -}}
package {{$entity}}html

import "{{.Mod}}/lib/{{.App}}/{{.Name}}/model"
import . "{{.Mod}}/lib/{{.App}}_web/components"

templ Edit(data model.{{.Entity}}, err error) {
	@Layout() {

	}
}
`

type editHtmlParam struct {
	Name     string `validate:"required"` // context name
	Mod      string `validate:"-"`        // go module name
	App      string `validate:"-"`        // application name
	Entity   string `validate:"required"` // entity name
	filename string
	_created bool
}

func (p *editHtmlParam) bind(ctx *cli.Context, args ...string) {
	if len(args) != 2 || args[0] == "" || args[1] == "" {
		return
	}
	p.Name = args[0]
	p.Entity = args[1]
	p.App = ctx.String("app")
}

func (p *editHtmlParam) setMod(mod string) {
	p.Mod = mod
}

func (p *editHtmlParam) setApp(app string) {
	p.App = app
}

func (p *editHtmlParam) created() bool {
	return p._created
}

func (p *editHtmlParam) executeTemplate() error {
	entity := snakecase.SnakeCase(p.Entity)
	p.filename = fmt.Sprintf("lib/%s_web/controllers/%s_html/%s.edit.templ", p.App, entity, entity)
	if err := os.MkdirAll(filepath.Dir(p.filename), os.ModePerm); err != nil {
		return err
	}
	// execute template
	temp, err := template.New("edit.html").
		Funcs(template.FuncMap{
			"lower": strings.ToLower,
		}).
		Parse(editHtmlTemplate)
	if err != nil {
		return err
	}
	p._created, err = executeTemplate(p.filename, p, temp)
	return err
}

func (p *editHtmlParam) rollback() {
	if p._created {
		if err := os.Remove(p.filename); err != nil {
			fmt.Println("- removed:", p.filename)
			p._created = false
		}
	}
}
