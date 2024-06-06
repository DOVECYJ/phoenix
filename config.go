package phoenix

import (
	"log/slog"
	"unsafe"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// A base config struct than provide Laod functions.
//
// Usage:
//
//	type MyConfig struct {
//		Configer[MyConfig]
//		Name string
//	}
//
// If you have a config in toml file like this:
//
//	name = 'hello'
//
// than you can load config as follow:
//
//	var c config
//	err := c.Load()
//
// If your config is in a section like:
//
//	[my]
//	name = 'hello'
//
// than you need to use LoadKey:
//
//	err := c.LoadKey("my")
type Configer[T any] struct{}

// Load config from viper to T
func (c *Configer[T]) Load() error {
	return viper.Unmarshal((*T)(unsafe.Pointer(c)))
}

// Load config key in viper to T
func (c *Configer[T]) LoadKey(key string) error {
	return viper.UnmarshalKey(key, (*T)(unsafe.Pointer(c)))
}

func (c *Configer[T]) LoadAndValide() error {
	if err := c.Load(); err != nil {
		return nil
	}
	return c.Validate()
}

func (c *Configer[T]) LoadKeyAndValide(key string) error {
	if err := c.LoadKey(key); err != nil {
		return nil
	}
	return c.Validate()
}

func (c *Configer[T]) Validate() error {
	return validator.New().Struct((*T)(unsafe.Pointer(c)))
}

type LogConfig struct {
	Configer[LogConfig] `json:"-"`
	Name                string
	Size                int
	Backups             int
	Age                 int
	Level               string
}

func (c LogConfig) LogLevel() slog.Leveler {
	var level slog.LevelVar
	switch viper.GetString("log.level") {
	case "error":
		level.Set(slog.LevelError)
	case "warn":
		level.Set(slog.LevelWarn)
	case "info":
		level.Set(slog.LevelInfo)
	case "debug":
		level.Set(slog.LevelDebug)
	default:
		level.Set(slog.LevelInfo)
	}

	return &level
}
