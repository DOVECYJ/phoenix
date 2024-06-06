package env

import (
	"time"

	"github.com/spf13/viper"
)

func StringOr(key, or string) string {
	if viper.InConfig(key) {
		return viper.GetString(key)
	}
	return or
}

func IntOr(key string, or int) int {
	if viper.InConfig(key) {
		return viper.GetInt(key)
	}
	return or
}

func Int32Or(key string, or int32) int32 {
	if viper.InConfig(key) {
		return viper.GetInt32(key)
	}
	return or
}

func Int64Or(key string, or int64) int64 {
	if viper.InConfig(key) {
		return viper.GetInt64(key)
	}
	return or
}

func Float32Or(key string, or float32) float32 {
	if viper.InConfig(key) {
		return float32(viper.GetFloat64(key))
	}
	return or
}

func Float64Or(key string, or float64) float64 {
	if viper.InConfig(key) {
		return viper.GetFloat64(key)
	}
	return or
}

func DurationOr(key string, or time.Duration) time.Duration {
	if viper.InConfig(key) {
		return viper.GetDuration(key)
	}
	return or
}
