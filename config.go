package phoenix

import (
	"log/slog"
	"path"
	"unsafe"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

func init() {
	BeforeLoadConfig("log", func() {
		viper.SetDefault("log.size", 100)
		viper.SetDefault("log.backups", 1000)
		viper.SetDefault("log.age", 30)
		viper.SetDefault("log.level", "info")
	})
	AfterLoadCondig("log", ConfigSlog)
}

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

// Config slog by config file.
// log.name is the name of log file, when empty, slog behave default.
// log.size is the max size of single log file in MB.
// log.backups is max count of log files.
// log.age is the max keep time of each log file.
// log.level is the log level.
//
//	[log]
//	name = 'app.log'
//	size = 1024
//	backups = 1000
//	age = 30
//	level = 'debug'
//
// The log filed will default go into 'log' directory.
func ConfigSlog() error {
	var c LogConfig
	if err := c.LoadKey("log"); err != nil {
		return err
	}
	slog.Info("load log config", "config", c)
	if c.Name != "" {
		output := &lumberjack.Logger{
			Filename:   path.Join("logs", c.Name),
			MaxSize:    c.Size,    //MB
			MaxBackups: c.Backups, //最大日志保留数量
			MaxAge:     c.Age,     //最大日志保留时长
			Compress:   true,
			LocalTime:  true,
		}
		handler := slog.NewTextHandler(output, &slog.HandlerOptions{Level: c.LogLevel()})
		slogger := slog.New(handler)
		slog.SetDefault(slogger)
	}
	return nil
}

var (
	beforeLoadConfig = map[string]func(){}
	afterLoadConfig  = map[string]func() error{}
)

// Register function run before load config
func BeforeLoadConfig(name string, action func()) {
	beforeLoadConfig[name] = action
}

// Register function run after load config
func AfterLoadCondig(name string, action func() error) {
	afterLoadConfig[name] = action
}

// Load config from file
func LoadConfig(name string) (err error) {
	for k, v := range beforeLoadConfig {
		v()
		slog.Info("before load config", "action", k)
	}
	name = "config/" + name
	viper.SetConfigFile(name)
	if err = viper.ReadInConfig(); err != nil {
		return
	}
	slog.Info("config loaded", "name", name)
	for k, v := range afterLoadConfig {
		if err = v(); err != nil {
			slog.Error("after load config", "action", k, "error", err)
			return err
		}
		slog.Info("after load config", "action", k)
	}
	return nil
}

// Same to LoadConfig but it will panic when there is an error
func MustLoadConfig(name string) {
	PanicError(LoadConfig(name))
}
