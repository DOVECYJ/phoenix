package main

import (
	"context"
	"flag"
	"reflect"
	"strings"
	"unsafe"

	"github.com/go-playground/validator/v10"
	"github.com/urfave/cli/v2"
)

var validate = validator.New()

type binder interface {
	bind(*cli.Context, ...string)
}

// Bind command flags to b and validate it
func bindAndValide(ctx *cli.Context, b ...binder) error {
	args, err := parseFlags(ctx)
	if err != nil {
		return err
	}
	for i := range b {
		b[i].bind(ctx, args...)
		if err := validate.Struct(b[i]); err != nil {
			return err
		}
	}
	return nil
}

// Parse command flags form ctx
func parseFlags(ctx *cli.Context) ([]string, error) {
	args := ctx.Args().Slice()
	if len(args) == 0 {
		return args, nil
	}
	var i int
	for ; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-") || strings.HasPrefix(args[i], "--") {
			break
		}
	}
	_ctx := (*struct {
		context.Context
		App           *cli.App
		Command       *cli.Command
		shellComplete bool
		flagSet       *flag.FlagSet
		parentContext *cli.Context
	})(unsafe.Pointer(ctx))

	err := _ctx.flagSet.Parse(args[i:])
	return args[:i], err
}

// Deprecated: This is the reflect implementation of 'parseFlags'
func parseFlagsReflect(ctx *cli.Context) ([]string, error) {
	args := ctx.Args().Slice()
	if len(args) == 0 {
		return nil, nil
	}

	var i int
	for ; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-") || strings.HasPrefix(args[i], "--") {
			break
		}
	}

	v := reflect.ValueOf(ctx).Elem().FieldByName("flagSet")
	fs := reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem().Interface().(*flag.FlagSet)

	err := fs.Parse(args[i:])
	return args[:i], err
}
