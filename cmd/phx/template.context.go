package main

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/DOVECYJ/phoenix/cmd"
	"github.com/azer/snakecase"
	"github.com/jinzhu/inflection"
	"github.com/urfave/cli/v2"
)

const contextTemplate = `package {{.Name}}

import (
	"context"
	"{{.Mod}}/lib/{{.App}}/{{.Name}}/model"
	"{{.Mod}}/pkg/repo"

	"github.com/go-rel/changeset"
	"github.com/go-rel/changeset/params"
	"github.com/go-rel/rel/where"
)

func List{{plural .Entity}}(ctx context.Context) (data []model.{{.Entity}}, err error) {
	err = repo.Repo.FindAll(ctx, &data)
	return
}

func Get{{.Entity}}(ctx context.Context, id int) (data model.{{.Entity}}, err error) {
	err = repo.Repo.Find(ctx, &data, where.Eq("id", id))
	return
}

func Create{{.Entity}}(ctx context.Context, params params.Params) (data model.{{.Entity}}, cs *changeset.Changeset, err error) {
	cs = Change{{.Entity}}(&data, params)
	if err = cs.Error(); err != nil {
		return
	}
	err = repo.Repo.Insert(ctx, &data, cs)
	return
}

func Update{{.Entity}}(ctx context.Context, data *model.{{.Entity}}, params params.Params) (cs *changeset.Changeset, err error) {
	cs = Change{{.Entity}}(data, params)
	if err = cs.Error(); err != nil {
		return
	}
	err = repo.Repo.Update(ctx, data, cs)
	return
}

func Delete{{.Entity}}(ctx context.Context, data model.{{.Entity}}) error {
	return repo.Repo.Delete(ctx, &data)
}

func Change{{.Entity}}(data *model.{{.Entity}}, params params.Params) *changeset.Changeset {
	return model.Change{{.Entity}}(data, params)
}
`

// Params for generate context
type contextParam struct {
	Name     string `validate:"required"` // context name
	Mod      string `validate:"-"`        // go module name
	App      string `validate:"-"`        // application name
	Entity   string `validate:"required"` // entity name
	filename string
	_created bool
}

// Bind command params to contextParam
// args[0] : context name
// args[1] : entity name
func (p *contextParam) bind(ctx *cli.Context, args ...string) {
	if len(args) != 2 || args[0] == "" || args[1] == "" {
		return
	}
	p.Name = args[0]
	p.Entity = args[1]
	p.App = ctx.String("app")
}

func (p *contextParam) setMod(mod string) {
	p.Mod = mod
}

func (p *contextParam) setApp(app string) {
	p.App = app
}

func (p *contextParam) created() bool {
	return p._created
}

// Execute context template
func (p *contextParam) executeTemplate() error {
	entity := snakecase.SnakeCase(p.Entity)
	p.filename = fmt.Sprintf("lib/%s/%s/%s.go", p.App, p.Name, entity)
	if err := os.MkdirAll(filepath.Dir(p.filename), os.ModePerm); err != nil {
		return err
	}
	// execute template
	temp, err := template.New("context").
		Funcs(template.FuncMap{"plural": inflection.Plural}).
		Parse(contextTemplate)
	if err != nil {
		return err
	}
	if p._created, err = executeTemplate(p.filename, p, temp); err != nil {
		return err
	}
	// fmt go source file
	return cmd.Cmd("go fmt " + p.filename).Run()
}

func (p *contextParam) rollback() {
	if p._created {
		os.Remove(p.filename)
		fmt.Println("- removed:", p.filename)
	}
}
