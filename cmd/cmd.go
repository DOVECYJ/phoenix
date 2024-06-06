package cmd

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

// A exec command wraper make run command like in shell.
//
// Usage:
//
//	err := Cmd("go fmt ./...").Run()
//
// If you want to specify a directory for command, just call 'RunOn':
//
//	err := Cmd("go fmt ./...").RunOn("hello")
//
// We only return the error of the runing result and log the command's
// output when there is an error via slog.
type Cmd string

// Run command
func (c Cmd) Run() error {
	return c.run(nil)
}

// Run command on dir
func (c Cmd) RunOn(dir string) error {
	return c.run(func(cmd *exec.Cmd) {
		cmd.Dir = dir
	})
}

// Run command and get output string
func (c Cmd) Call() (string, error) {
	return c.runOutput(nil)
}

// Run command on dir and get output string
func (c Cmd) CallOn(dir string) (string, error) {
	return c.runOutput(func(cmd *exec.Cmd) {
		cmd.Dir = dir
	})
}

func (c Cmd) runOutput(fn func(*exec.Cmd)) (string, error) {
	if c == "" {
		return "", errors.New("command is empty")
	}
	ss := strings.Split(string(c), " ")
	cmd := exec.Command(ss[0], ss[1:]...)
	if fn != nil {
		fn(cmd)
	}
	bs, err := cmd.CombinedOutput()
	return string(bs), err
}

func (c Cmd) run(fn func(*exec.Cmd)) error {
	if c == "" {
		return errors.New("command is empty")
	}
	ss := strings.Split(string(c), " ")
	cmd := exec.Command(ss[0], ss[1:]...)
	if fn != nil {
		fn(cmd)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// CmdSet allow you to run a set of commands at one call.
//
// Usage:
//
//	err := CmdSet([]string{
//		"go mod init hello",
//		"go fmt ./...",
//	}).Run()
//
// or you can specify a directory for all commands via 'RunOn':
//
//	err := CmdSet([]string{
//		"go mod init hello",
//		"go fmt ./...",
//	}).RunOn("hello")
//
// We only return the error as result and log the command's output
// when there is an error vis slog.
type CmdSet []string

// Run commands
func (c CmdSet) Run() error {
	for _, s := range c {
		if err := Cmd(s).Run(); err != nil {
			return err
		}
	}
	return nil
}

// Run commads on dir
func (c CmdSet) RunOn(dir string) error {
	for _, s := range c {
		if err := Cmd(s).RunOn(dir); err != nil {
			return err
		}
	}
	return nil
}
