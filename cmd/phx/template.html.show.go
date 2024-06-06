package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/serenize/snaker"
	"github.com/urfave/cli/v2"
)

const showHtmlTemplate = `
{{- $entity := lower .Entity -}}
package {{$entity}}html

import "{{.Mod}}/lib/{{.App}}/{{.Name}}/model"
import . "{{.Mod}}/lib/{{.App}}_web/components"

templ Show(data model.{{.Entity}}, err error) {
	@Layout() {

	}
}
`

type showHtmlParam struct {
	Name     string `validate:"required"` // context name
	Mod      string `validate:"-"`        // go module name
	App      string `validate:"-"`        // application name
	Entity   string `validate:"required"` // entity name
	filename string
	_created bool
}

func (p *showHtmlParam) bind(ctx *cli.Context, args ...string) {
	if len(args) != 2 || args[0] == "" || args[1] == "" {
		return
	}
	p.Name = args[0]
	p.Entity = args[1]
	p.App = ctx.String("app")
}

func (p *showHtmlParam) setMod(mod string) {
	p.Mod = mod
}

func (p *showHtmlParam) setApp(app string) {
	p.App = app
}

func (p *showHtmlParam) created() bool {
	return p._created
}

func (p *showHtmlParam) executeTemplate() error {
	entity := snaker.CamelToSnake(p.Entity)
	p.filename = fmt.Sprintf("lib/%s_web/controllers/%s_html/%s.show.templ", p.App, entity, entity)
	if err := os.MkdirAll(filepath.Dir(p.filename), os.ModePerm); err != nil {
		return err
	}
	// execute template
	temp, err := template.New("show.html").
		Funcs(template.FuncMap{
			"lower": strings.ToLower,
		}).
		Parse(showHtmlTemplate)
	if err != nil {
		return err
	}
	p._created, err = executeTemplate(p.filename, p, temp)
	return err
}

func (p *showHtmlParam) rollback() {
	if p._created {
		os.Remove(p.filename)
		fmt.Println("- removed:", p.filename)
	}
}
