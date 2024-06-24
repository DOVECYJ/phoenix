package phoenix

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"github.com/DOVECYJ/phoenix/env"
	"github.com/go-rel/changeset"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

// The interface of Application
type IApplication interface {
	Start()
	Stop()
	Name() string
}

// The base application with a name. You can embed it into your own application:
//
//	type MyApplication struct {
//		Application
//	}
type Application struct {
	IApplication
	name string
}

func (a Application) Name() string {
	return a.name
}

func NewApplication(name string) Application {
	return Application{name: name}
}

// Blocked to wait for system kill signal, usually triggered by Ctrl+C.
func WaitForExitSignal() {
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	sig := <-exit
	slog.Info("system exit", "signal", sig)
}

// Run each application in independent goroutine, than wait for system kill signal.
// Wnen exit, stop each application in reverse order.
func RunApplications(applications ...IApplication) {
	// 记录进程PID
	MustWritePID("pid")
	// 应用启动
	for i := range applications {
		go applications[i].Start()
		slog.Info("application started", "name", applications[i].Name())

		defer func(a IApplication) {
			a.Stop()
			slog.Info("application stoped", "name", a.Name())
		}(applications[i])
	}
	// 等待退出信号
	WaitForExitSignal()
}

// Load config from file
func LoadConfig(name string) (err error) {
	name = "config/" + name
	viper.SetConfigFile(name)
	if err = viper.ReadInConfig(); err != nil {
		return
	}
	slog.Info("config loaded", "name", name)
	return nil
}

// Same to LoadConfig but it will panic when there is an error
func MustLoadConfig(name string) {
	if err := LoadConfig(name); err != nil {
		panic(err)
	}
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

// Like ConfigSLog but panic whan error.
func MustConfigSlog() {
	if err := ConfigSlog(); err != nil {
		panic(err)
	}
}

// Like ConfigEnv but panic when error.
func MustConfigEnv() {
	if err := env.ConfigEnv(); err != nil {
		panic(err)
	}
}

// write process id into file, it will panic when fail.
func MustWritePID(name string) {
	pid := os.Getpid()
	fpid, err := os.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer fpid.Close()
	if _, err = fmt.Fprint(fpid, pid); err != nil {
		panic(err)
	}
}

type FieldError map[string]error

func (e FieldError) Error() string {
	var sb strings.Builder
	for k, v := range e {
		sb.WriteString(k)
		sb.WriteByte(':')
		sb.WriteString(v.Error())
		sb.WriteByte('\n')
	}
	return sb.String()
}

func ExtractChangesetError(c *changeset.Changeset) FieldError {
	if c == nil || c.Error() == nil {
		return nil
	}
	err := FieldError{}
	for _, e := range c.Errors() {
		if e, ok := e.(changeset.Error); ok {
			err[e.Field] = e.Err
		}
	}
	return err
}

type CtxKey string

var (
	ID = CtxKey("id")
)

func PanicError(err error) {
	if err != nil {
		panic(err)
	}
}
