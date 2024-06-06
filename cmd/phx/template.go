package main

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/urfave/cli/v2"
)

//go:embed _templates/*
var templets embed.FS

// The config for generate new project.
type config struct {
	Dir        string `validate:"required"`                           // 项目目录
	Name       string `validate:"required"`                           // 项目名
	Mod        string `validate:"required"`                           // module名
	App        string `validate:"required"`                           // application名
	Database   string `validate:"required,oneof=mysql pgsql sqlite3"` // 数据库类型
	NoDatabase bool   // 不使用数据库
	NoRedis    bool   // 不使用redis
	NoHtml     bool   // 不生成HTML
}

func (c config) getByName(key string) string {
	switch strings.ToLower(key) {
	case "dir":
		return c.Dir
	case "name":
		return c.Name
	case "mod":
		return c.Mod
	case "app":
		return c.App
	case "database":
		return c.Database
	default:
		return ""
	}
}

func (c *config) bind(ctx *cli.Context, args ...string) {
	if len(args) != 1 || args[0] == "" {
		return
	}
	c.Name = filepath.Base(args[0])
	c.Dir = filepath.Clean(args[0])
	if c.Mod = ctx.String("mod"); c.Mod == "" {
		c.Mod = c.Name
	}
	if c.App = ctx.String("app"); c.App == "" {
		c.App = c.Name
	}
	if c.Database = ctx.String("database"); c.Database == "" {
		c.Database = "mysql"
	}
	c.NoDatabase = ctx.Bool("no-database")
	c.NoRedis = ctx.Bool("no-redis")
	c.NoHtml = ctx.Bool("no-html")
}

// Initialize project in an empty direatory.
func initProject(c config) error {
	ds, err := os.ReadDir(c.Dir)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	} else if len(ds) > 0 {
		return fmt.Errorf("directory '%s' is not empty", c.Dir)
	}

	var createdFiles []string
	err = fs.WalkDir(templets, "_templates",
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			name := filepath.Join(c.Dir, strings.TrimPrefix(path, "_templates"))
			if c.NoDatabase && (strings.HasSuffix(name, "repo.go.tp") || strings.HasSuffix(name, "migrate.go.tp")) {
				return nil
			}
			if c.NoRedis && strings.HasSuffix(name, "cache.go.tp") {
				return nil
			}
			if c.NoHtml && (strings.Contains(name, "components") || strings.Contains(name, "page_html") || strings.Contains(name, "assets")) {
				return nil
			}
			if strings.Contains(name, "$") {
				name = os.Expand(name, c.getByName)
			}
			if strings.HasSuffix(path, ".tp") {
				// 模板
				name = strings.TrimSuffix(name, ".tp")
				err = execute(name, path, c)
			} else {
				// 文件
				err = copy(name, path)
			}
			createdFiles = append(createdFiles, name)
			fmt.Printf("create: *%s\n", name)
			return err
		},
	)
	if err != nil {
		// TODO: 失败时清理生成的文件
		return err
	}
	if err = os.MkdirAll(filepath.Join(c.Dir, "_build"), os.ModePerm); err != nil {
		return err
	}
	if !c.NoDatabase {
		if err = os.MkdirAll(filepath.Join(c.Dir, "priv", "repo", "migrations"), os.ModePerm); err != nil {
			return err
		}
	}
	if !c.NoDatabase && c.Database == "sqlite3" {
		f, err := os.Create(filepath.Join(c.Dir, c.Name+"_dev.db"))
		if err != nil {
			return err
		}
		defer f.Close()
	}
	return err
}

// Execute src template and write to dest.
func execute(dst, src string, c config) error {
	if err := os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
		return err
	}
	fw, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer fw.Close()

	templ, err := template.ParseFS(templets, src)
	if err != nil {
		return err
	}

	return templ.Execute(fw, c)
}

// Copy src file to dest.
func copy(dst, src string) error {
	if err := os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
		return err
	}

	fw, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer fw.Close()

	fr, err := templets.Open(src)
	if err != nil {
		return err
	}
	defer fr.Close()

	_, err = io.Copy(fw, fr)
	return err
}

func executeTemplate(filename string, data any, temp *template.Template) (bool, error) {
	_, err := os.Stat(filename)
	if !errors.Is(err, os.ErrNotExist) {
		fmt.Printf("! skipted: %s\n", filename)
		return false, nil
	}

	f, err := os.Create(filename)
	if err != nil {
		return false, err
	}
	defer f.Close()

	fmt.Printf("* create: %s\n", filename)
	return true, temp.Execute(f, data)
}

type generator interface {
	executeTemplate() error
	rollback()
	created() bool
	modSetter
	appSetter
	binder
}

type modSetter interface {
	setMod(string)
}
type appSetter interface {
	setApp(string)
}

func generateCode(hasMod, hasApp bool, generators ...generator) (err error) {
	if !hasMod {
		if err = setMod(func(mod string) {
			for _, g := range generators {
				g.setMod(mod)
			}
		}); err != nil {
			return
		}
	}
	if !hasApp {
		if err = setApp(func(app string) {
			for _, g := range generators {
				g.setApp(app)
			}
		}); err != nil {
			return
		}
	}
	// generate code
	var i int
	for i = range generators {
		if err = generators[i].executeTemplate(); err != nil {
			break
		}
	}
	if err != nil {
		for ; i >= 0; i-- {
			generators[i].rollback()
		}
	}
	return
}
