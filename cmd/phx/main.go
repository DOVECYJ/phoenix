package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/DOVECYJ/phoenix/cmd"
	"github.com/urfave/cli/v2"
)

var (
	ErrAbort = errors.New("aborted")
)

// code generator
//
// Usage:
//
//	phx new hello --mod github.com/chenyj/hello --app hello
//	phx new hello --no-html --no-database --no-redis
//	phx gen.context user User --app hello
//	phx gen.html user User --table users --fields Name:string --app hello
//	phx gen.api user User --table users --fields Name:string --app hello
//	phx build
//	phx run
//	phx migrate
//	phx rollback
//	phx rollback --s 1
//	phx rollback --v 202405061234
func main() {
	app := &cli.App{
		Name:        "phx",
		Description: "phoenix command line tool",
		Version:     "v0.0.1",
		Commands: []*cli.Command{
			{ // new project
				Name:  "new",
				Usage: "create a new project",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "mod",
						Usage: "go module name",
						Value: "",
					},
					&cli.StringFlag{
						Name:  "app",
						Usage: "application name",
						Value: "",
					},
					&cli.StringFlag{
						Name:  "database",
						Usage: "database to use",
						Value: "mysql",
					},
					&cli.BoolFlag{
						Name:  "no-database",
						Usage: "disable database",
						Value: false,
					},
					&cli.BoolFlag{
						Name:  "no-redis",
						Usage: "disable redis",
						Value: false,
					},
					&cli.BoolFlag{
						Name:  "no-html",
						Usage: "disable html",
						Value: false,
					},
				},
				Action: func(ctx *cli.Context) error {
					var c config
					if err := bindAndValide(ctx, &c); err != nil {
						return err
					}
					// fmt.Printf("%#v\n", c)
					if err := initProject(c); err != nil {
						return err
					}
					if err := cmd.CmdSet([]string{
						"go mod init " + c.Mod, // go mod init
						"templ templ fmt .",    // templ fmt
						"templ generate",       // templ generate
						"go fmt ./...",         // go fmt
					}).RunOn(c.Dir); err != nil {
						return err
					}
					yes, err := yesORno(
						func() error { return cmd.Cmd("go mod tidy").RunOn(c.Dir) },
						"Fetch and install dependencies? [Y/N]: ")
					if err != nil {
						return err
					}
					fmt.Println("We are almost there! The following steps are missing: \n\n\t $ cd", c.Dir)
					if !yes {
						fmt.Println("\nThen fetch and install dependencies:\n\n\t$ go mod tidy")
					}
					fmt.Println("\nStart your Phoenix app with:\n\n\t$ go run .")
					return nil
				},
			},
			{ // clean project
				Name:  "clean",
				Usage: "clean project",
				Action: func(ctx *cli.Context) error {
					name := ctx.Args().First()
					if name == "" {
						return errors.New("project name can not be empty")
					}
					dir, err := os.Getwd()
					if err != nil {
						return err
					}
					dir = filepath.Join(dir, name)
					_, err = yesORno(
						func() error { return os.RemoveAll(dir) },
						"You want to remove all items in '%s' ? [Y/N]: ",
						dir)
					return err
				},
			},
			{ // generate html
				Name:  "gen.html",
				Usage: "generate html",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "app",
						Usage: "application name",
						Value: "",
					},
					&cli.StringFlag{
						Name:  "table",
						Usage: "database table name",
						Value: "",
					},
					&cli.StringSliceFlag{
						Name:  "fields",
						Usage: "entity fields(Field:Type,Field:Type...)",
						Value: nil,
					},
				},
				Action: func(ctx *cli.Context) error {
					var (
						modelParam      = new(modelParam)
						contextParam    = new(contextParam)
						controllerParam = new(controllerParam)
						migrationParam  = new(migrationParam)
						migrateParam    = new(migrateParam)
						indexHtmlParam  = new(indexHtmlParam)
						newHtmlParam    = new(newHtmlParam)
						editHtmlParam   = new(editHtmlParam)
						showHtmlParam   = new(showHtmlParam)
					)
					if err := bindAndValide(ctx,
						modelParam,
						contextParam,
						controllerParam,
						migrationParam,
						indexHtmlParam,
						newHtmlParam,
						editHtmlParam,
						showHtmlParam); err != nil {
						return err
					}
					migrateParam.bindMigration(migrationParam)
					if err := generateCode(ctx.IsSet("mod"), ctx.IsSet("app"),
						modelParam,
						contextParam,
						controllerParam,
						migrationParam,
						migrateParam,
						indexHtmlParam,
						newHtmlParam,
						editHtmlParam,
						showHtmlParam); err != nil {
						return err
					}
					if err := cmd.Cmd("templ generate").Run(); err != nil {
						return err
					}
					fmt.Printf("\nAdd the resource to your router in\n"+
						"lib/%s_web/router.go:\n\n\t"+
						"root.Route(\"/%s\", router.Resource(controllers.%sController{}))\n",
						controllerParam.App, controllerParam.Path, controllerParam.Entity)
					if !migrateParam.created() {
						fmt.Printf("\nAdd the migration to your migrate in\n"+
							"priv/repo/migrate.go\n\n\t"+
							"{%s, migrations.MigrateCreate%s, migrations.RollbackCreate%s},\n\n",
							migrateParam.Version, migrateParam.Entity, migrateParam.Entity)
					}
					return nil
				},
			},
			{ // generate api
				Name:  "gen.api",
				Usage: "generate api controller",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "app",
						Usage: "application name",
						Value: "",
					},
					&cli.StringFlag{
						Name:  "table",
						Usage: "database table name",
						Value: "",
					},
					&cli.StringSliceFlag{
						Name:  "fields",
						Usage: "entity fields(Field:Type,Field:Type...)",
						Value: nil,
					},
				},
				Action: func(ctx *cli.Context) error {
					var (
						modelParam      = new(modelParam)
						contextParam    = new(contextParam)
						controllerParam = new(apiControllerParam)
						migrationParam  = new(migrationParam)
						migrateParam    = new(migrateParam)
					)
					if err := bindAndValide(ctx,
						modelParam,
						contextParam,
						controllerParam,
						migrationParam); err != nil {
						return err
					}
					migrateParam.bindMigration(migrationParam)
					if err := generateCode(ctx.IsSet("mod"), ctx.IsSet("app"),
						modelParam,
						contextParam,
						controllerParam,
						migrationParam,
						migrateParam); err != nil {
						return err
					}
					if !migrateParam.created() {
						fmt.Printf("\nAdd the migration to your migrate in\n"+
							"priv/repo/migrate.go\n\n\t"+
							"{%s, migrations.MigrateCreate%s, migrations.RollbackCreate%s},\n\n",
							migrateParam.Version, migrateParam.Entity, migrateParam.Entity)
					}
					return nil
				},
			},
			{ //generate context
				Name:  "gen.context",
				Usage: "generate context",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "app",
						Usage: "application name",
						Value: "",
					},
					&cli.StringFlag{
						Name:  "table",
						Usage: "database table name",
						Value: "",
					},
					&cli.StringSliceFlag{
						Name:  "fields",
						Usage: "entity fields(Field:Type,Field:Type...)",
						Value: nil,
					},
				},
				Action: func(ctx *cli.Context) error {
					var (
						modelParam     = new(modelParam)
						contextParam   = new(contextParam)
						migrationParam = new(migrationParam)
						migrateParam   = new(migrateParam)
					)
					if err := bindAndValide(ctx,
						modelParam,
						contextParam,
						migrationParam); err != nil {
						return err
					}
					migrateParam.bindMigration(migrationParam)
					if err := generateCode(ctx.IsSet("mod"), ctx.IsSet("app"),
						modelParam,
						contextParam,
						migrationParam,
						migrateParam); err != nil {
						return err
					}
					if !migrateParam.created() {
						fmt.Printf("\nAdd the migration to your migrate in\n"+
							"priv/repo/migrate.go\n\n\t"+
							"{%s, migrations.MigrateCreate%s, migrations.RollbackCreate%s},\n\n",
							migrateParam.Version, migrateParam.Entity, migrateParam.Entity)
					}
					return nil
				},
			},
			{ // build service
				Name:  "build",
				Usage: "build service",
				Action: func(ctx *cli.Context) error {
					return cmd.CmdSet([]string{
						"templ generate",
						"go build -o _build .",
					}).Run()
				},
			},
			{ // run service
				Name:  "run",
				Usage: "run service",
				Action: func(ctx *cli.Context) error {
					return cmd.CmdSet([]string{
						"templ generate",
						"go run .",
					}).Run()
				},
			},
			{ // migrate schema
				Name:  "migrate",
				Usage: "migrate tables",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "config",
						Usage: "config filename",
						Value: "",
					},
				},
				Action: func(ctx *cli.Context) (err error) {
					smd := "go run priv/repo/migrate.go migrate"
					if ctx.IsSet("config") {
						smd = smd + " --config " + ctx.String("config")
					}
					if err = cmd.Cmd(smd).Run(); err == nil {
						fmt.Println("Migrate Success!")
					}
					return
				},
			},
			{ // rollback migration
				Name:  "rollback",
				Usage: "rollback migration",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:  "v",
						Usage: "version to rollback to",
						Value: -1,
					},
					&cli.IntFlag{
						Name:  "n",
						Usage: "steps to rollback",
						Value: -1,
					},
					&cli.StringFlag{
						Name:  "config",
						Usage: "config filename",
						Value: "application.toml",
					},
				},
				Action: func(ctx *cli.Context) (err error) {
					smd := "go run priv/repo/migrate.go rollback"
					if ctx.IsSet("v") {
						smd = smd + " --v " + ctx.String("v")
					}
					if ctx.IsSet("n") {
						smd = smd + " --n " + ctx.String("n")
					}
					if ctx.IsSet("config") {
						smd = smd + " --config " + ctx.String("config")
					}
					err = cmd.Cmd(smd).Run()
					if err == nil {
						fmt.Println("Rollback Success!")
					}
					return
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}

// Answer yes or no and run fn when answer is yes.
func yesORno(fn func() error, prompt string, args ...any) (bool, error) {
	var answer string
	fmt.Printf(prompt, args...)
	fmt.Scanln(&answer)
	if answer == "y" || answer == "Y" {
		return true, fn()
	}
	return false, nil
}

func setMod(fn func(string)) error {
	mod, err := getMod()
	if err != nil {
		return err
	}
	fn(mod)
	return nil
}

func setApp(fn func(string)) error {
	apps, app := getApps(), ""
	if len(apps) == 0 {
		return errors.New("can not find application, please specify an app with --app")
	}
	var answer string
	if len(apps) == 1 {
		app = apps[0]
		fmt.Printf("Will generate html and controller in app '%s' [Y/N]: ", app)
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			return ErrAbort
		}
	} else {
		fmt.Printf("There are %d app in you project, which one would you like to put your html and controller in:", len(apps))
		for i, app := range apps {
			fmt.Printf("%2d : %s\n", i+1, app)
		}
	SELECT:
		fmt.Print("enter :exit to quit. And your select: ")
		fmt.Scanln(&answer)
		if answer == ":exit" {
			return ErrAbort
		}
		if slices.Contains(apps, answer) {
			app = answer
		} else {
			i, err := strconv.Atoi(answer)
			if err != nil {
				fmt.Println(err)
				goto SELECT
			}
			if i < 1 || i > len(apps) {
				fmt.Println("your select is out of range:", i)
				goto SELECT
			}
			app = apps[i]
		}
	}
	fn(app)
	return nil
}

func getMod() (string, error) {
	f, err := os.Open("go.mod")
	if err != nil {
		return "", err
	}
	defer f.Close()

	bf := bufio.NewReader(f)
	var line string
	for !strings.HasPrefix(line, "module") {
		line, err = bf.ReadString('\n')
		if err != nil {
			return "", nil
		}
	}
	if len(line) < 8 {
		return "", fmt.Errorf("go.mod format error: %s", line)
	}
	return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
}

func getApps() []string {
	des, err := os.ReadDir("./lib")
	if err != nil {
		return nil
	}
	var apps []string
	var m = make(map[string]struct{})
	for _, de := range des {
		if !de.IsDir() {
			continue
		}
		name := de.Name()
		if strings.HasSuffix(name, "_web") {
			app := strings.TrimSuffix(name, "_web")
			if _, ok := m[app]; ok {
				apps = append(apps, app)
			} else {
				m[name] = struct{}{}
			}
		} else {
			if _, ok := m[name+"_web"]; ok {
				apps = append(apps, name)
			} else {
				m[name] = struct{}{}
			}
		}
	}
	return apps
}
