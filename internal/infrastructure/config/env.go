package config

import (
	"os"
	"strconv"
)

type Env struct {
	Host string `mapstructure:"HOST"`
	Port int    `mapstructure:"PORT"`
	Env  string `mapstructure:"ENV"`
}

var env *Env

// GetConfig returns the configuration for the application
func LoadEnv() {
	env = &Env{
		Host: GetStrEnvOrPanic("HOST"),
		Port: GetIntEnvOrPanic("PORT"),
		Env:  GetStrEnvOrPanic("ENV"),
	}
}

func GetStrEnvOrPanic(env string) string {
	res := os.Getenv(env)
	if len(res) == 0 {
		panic("Mandatory env variable not found:" + env)
	}
	return res
}

func GetIntEnvOrPanic(env string) int {
	res := os.Getenv(env)
	if len(res) == 0 {
		panic("Mandatory env variable not found:" + env)
	}
	i, err := strconv.Atoi(res)
	if err != nil {
		panic("Mandatory env variable not found:" + env)
	}
	return i
}

func GetBoolEnvOrPanic(env string) bool {
	res := os.Getenv(env)
	if len(res) == 0 {
		panic("Mandatory env variable not found:" + env)
	}
	b, err := strconv.ParseBool(res)
	if err != nil {
		panic("Mandatory env variable not found:" + env)
	}
	return b
}
