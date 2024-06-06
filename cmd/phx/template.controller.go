package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/DOVECYJ/phoenix/cmd"
	"github.com/azer/snakecase"
	"github.com/jinzhu/inflection"
	"github.com/urfave/cli/v2"
)

const (
	controllerTemplate = `package controllers
{{$entity := lower .Entity}}
import (
	"fmt"
	"{{.Mod}}/lib/{{.App}}/{{.Name}}"
	"{{.Mod}}/lib/{{.App}}/{{.Name}}/model"
	{{$entity}}html "{{.Mod}}/lib/{{.App}}_web/controllers/{{snake .Entity}}_html"
	"log/slog"
	"net/http"

	"github.com/DOVECYJ/phoenix/binding"
	"github.com/DOVECYJ/phoenix/render"
	"github.com/DOVECYJ/phoenix/router"
)

type {{.Entity}}Controller struct {
	router.IResource
}

func ({{.Entity}}Controller) Index(w http.ResponseWriter, r *http.Request) {
	data, err := {{.Name}}.List{{plural .Entity}}(r.Context())
	render.HTML(w, {{$entity}}html.Index(data, err))
}

func ({{.Entity}}Controller) Edit(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("id").(int)
	data, err := {{.Name}}.Get{{.Entity}}(r.Context(), id)
	if err != nil {
		render.HTML(w, {{$entity}}html.Edit(data, err))
		return
	}
	render.HTML(w, {{$entity}}html.Edit(data, nil))
}

func ({{.Entity}}Controller) New(w http.ResponseWriter, r *http.Request) {
	render.HTML(w, {{$entity}}html.New(nil))
}

func ({{.Entity}}Controller) Show(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("id").(int)
	data, err := {{.Name}}.Get{{.Entity}}(r.Context(), id)
	if err != nil {
		render.HTML(w, {{$entity}}html.Show(data, err))
		return
	}
	render.HTML(w, {{$entity}}html.Show(data, nil))
}

func ({{.Entity}}Controller) Create(w http.ResponseWriter, r *http.Request) {
	params, err := binding.Attr(r)
	if err != nil {
		render.HTML(w, {{$entity}}html.New(err))
		return
	}
	
	data, _, err := {{.Name}}.Create{{.Entity}}(r.Context(), params)
	if err != nil {
		render.HTML(w, {{$entity}}html.New(err))
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/{{.Path}}/%d", data.ID), http.StatusFound)
}

func ({{.Entity}}Controller) Update(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("id").(int)
	params, err := binding.Attr(r)
	if err != nil {
		render.HTML(w, {{$entity}}html.Edit(model.{{.Entity}}{}, err))
		return
	}
	
	data, err := {{.Name}}.Get{{.Entity}}(r.Context(), id)
	if err != nil {
		render.HTML(w, {{$entity}}html.Edit(data, err))
		return
	}
	
	_, err = {{.Name}}.Update{{.Entity}}(r.Context(), &data, params)
	if err != nil {
		render.HTML(w, {{$entity}}html.Edit(data, err))
		return
	}
	http.Redirect(w, r, "/{{.Path}}", http.StatusFound)
}

func ({{.Entity}}Controller) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("id").(int)
	data, err := {{.Name}}.Get{{.Entity}}(r.Context(), id)
	if err != nil {
		http.Redirect(w, r, "/{{.Path}}", http.StatusFound)
		return
	}
	err = {{.Name}}.Delete{{.Entity}}(r.Context(), data)
	if err != nil {
		slog.Error("delete user", "error", err)
		return
	}
	http.Redirect(w, r, "/{{.Path}}", http.StatusFound)
}
`

	apiControllerTemplate = `package controllers

import (
	"net/http"
	"{{.Mod}}/lib/{{.App}}/{{.Name}}"

	"github.com/DOVECYJ/phoenix/binding"
	"github.com/DOVECYJ/phoenix/render"
)
{{$Entities := plural .Entity}}
func List{{$Entities}}(w http.ResponseWriter, r *http.Request) {
	data, err := {{.Name}}.List{{$Entities}}(r.Context())
	if err != nil {
		render.Render(w, err)
		return
	}
	render.ApiData(w, data)
}

func Get{{.Entity}}(w http.ResponseWriter, r *http.Request) {
	id := binding.ContextID(r)
	data, err := {{.Name}}.Get{{.Entity}}(r.Context(), id)
	if err != nil {
		render.Render(w, err)
		return
	}
	render.ApiData(w, data)
}

func Create{{.Entity}}(w http.ResponseWriter, r *http.Request) {
	param, err := binding.Attr(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data, _, err := {{.Name}}.Create{{.Entity}}(r.Context(), param)
	if err != nil {
		render.Render(w, err)
		return
	}
	render.ApiData(w, data)
}

func Update{{.Entity}}(w http.ResponseWriter, r *http.Request) {
	id := binding.ContextID(r)
	param, err := binding.Attr(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data, err := {{.Name}}.Get{{.Entity}}(r.Context(), id)
	if err != nil {
		render.Render(w, err)
		return
	}
	_, err = {{.Name}}.Update{{.Entity}}(r.Context(), &data, param)
	if err != nil {
		render.Render(w, err)
		return
	}
	render.ApiData(w, data)
}

func Delete{{.Entity}}(w http.ResponseWriter, r *http.Request) {
	id := binding.ContextID(r)
	data, err := {{.Name}}.Get{{.Entity}}(r.Context(), id)
	if err != nil {
		render.Render(w, err)
		return
	}
	if err := {{.Name}}.Delete{{.Entity}}(r.Context(), data); err != nil {
		render.Render(w, err)
		return
	}
	render.ApiData(w, nil)
}
`
)

type controllerParam struct {
	Name     string `validate:"required"` // context name
	Mod      string `validate:"-"`        // go module name
	App      string `validate:"-"`        // application name
	Entity   string `validate:"required"` // entity name
	Path     string
	filename string
	_created bool
}

func (p *controllerParam) bind(ctx *cli.Context, args ...string) {
	if len(args) != 2 || args[0] == "" || args[1] == "" {
		return
	}
	p.Name = args[0]
	p.Entity = args[1]
	p.App = ctx.String("app")
	p.Path = snakecase.SnakeCase(inflection.Plural(p.Entity))
}

func (p *controllerParam) setMod(mod string) {
	p.Mod = mod
}

func (p *controllerParam) setApp(app string) {
	p.App = app
}

func (p *controllerParam) created() bool {
	return p._created
}

func (p *controllerParam) executeTemplate() error {
	entity := snakecase.SnakeCase(p.Entity)
	p.filename = fmt.Sprintf("lib/%s_web/controllers/%s_controller.go", p.App, entity)
	if err := os.MkdirAll(filepath.Dir(p.filename), os.ModePerm); err != nil {
		return err
	}
	// execute template
	temp, err := template.New("controller").
		Funcs(template.FuncMap{
			"lower":  strings.ToLower,
			"snake":  snakecase.SnakeCase,
			"plural": inflection.Plural,
		}).
		Parse(controllerTemplate)
	if err != nil {
		return err
	}
	if p._created, err = executeTemplate(p.filename, p, temp); err != nil {
		return err
	}
	return cmd.Cmd("go fmt " + p.filename).Run()
}

func (p *controllerParam) rollback() {
	if p._created {
		os.Remove(p.filename)
		fmt.Println("- removed:", p.filename)
	}
}

type apiControllerParam struct {
	controllerParam
}

func (p *apiControllerParam) executeTemplate() error {
	entity := snakecase.SnakeCase(p.Entity)
	p.filename = fmt.Sprintf("lib/%s_web/controllers/%s_controller.go", p.App, entity)
	if err := os.MkdirAll(filepath.Dir(p.filename), os.ModePerm); err != nil {
		return err
	}
	// execute template
	temp, err := template.New("api_controller").
		Funcs(template.FuncMap{"plural": inflection.Plural}).
		Parse(apiControllerTemplate)
	if err != nil {
		return err
	}
	if p._created, err = executeTemplate(p.filename, p, temp); err != nil {
		return err
	}
	return cmd.Cmd("go fmt " + p.filename).Run()
}
