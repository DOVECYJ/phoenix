package phoenix

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-rel/changeset"
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

// context key type
type CtxKey string

var (
	ID = CtxKey("id")
)

// panic on none nil error
func PanicError(err error) {
	if err != nil {
		panic(err)
	}
}
