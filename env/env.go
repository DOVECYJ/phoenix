package env

import (
	"errors"
	"fmt"
	"strings"

	"github.com/DOVECYJ/phoenix"
	"github.com/spf13/viper"
)

func init() {
	phoenix.AfterLoadCondig("env", ConfigEnv)
}

// Program's running environment
type ENV string

const (
	dev  ENV = "dev"
	test ENV = "test"
	prod ENV = "prod"
)

var (
	env         ENV
	ServiceName string
)

var (
	ErrLackServiceName = errors.New("lack of service name, add service = 'xxx' in your config")
)

// Read env form config file.
func ConfigEnv() (err error) {
	envstr := strings.ToLower(viper.GetString("env"))
	switch envstr {
	case string(dev):
		env = dev
	case string(test):
		env = test
	case string(prod):
		env = prod
	default:
		err = fmt.Errorf("env not set, env=%s", envstr)
	}

	if ServiceName = viper.GetString("service"); ServiceName == "" {
		return ErrLackServiceName
	}
	return
}

func IsDev() bool {
	return env == dev
}

func IsTest() bool {
	return env == test
}

func IsProd() bool {
	return env == prod
}
