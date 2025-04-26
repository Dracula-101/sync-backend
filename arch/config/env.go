package config

import (
	"os"
	"strconv"

	"github.com/subosito/gotenv"
)

type Env struct {
	Host     string `mapstructure:"HOST"`
	Port     int    `mapstructure:"PORT"`
	Env      string `mapstructure:"ENV"`
	LogLevel string `mapstructure:"LOG_LEVEL"`

	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PWD"`
	DBHost     string `mapstructure:"DB_HOST"`
	DBName     string `mapstructure:"DB_NAME"`

	IpDBHost     string `mapstructure:"IP_DB_HOST"`
	IpDBPort     int    `mapstructure:"IP_DB_PORT"`
	IpDBUser     string `mapstructure:"IP_DB_USER"`
	IpDBPassword string `mapstructure:"IP_DB_PWD"`
	IpDBName     string `mapstructure:"IP_DB_NAME"`

	RedisHost     string `mapstructure:"REDIS_HOST"`
	RedisPort     int    `mapstructure:"REDIS_PORT"`
	RedisPassword string `mapstructure:"REDIS_PWD"`
	RedisDB       int    `mapstructure:"REDIS_DB"`
}

func NewEnv(file string) Env {
	_ = gotenv.Load(file)
	env := Env{
		Host:          GetStrEnvOrPanic("HOST"),
		Port:          GetIntEnvOrPanic("PORT"),
		Env:           GetStrEnvOrPanic("ENV"),
		LogLevel:      GetStrEnvOrPanic("LOG_LEVEL"),
		DBUser:        GetStrEnvOrPanic("DB_USER"),
		DBPassword:    GetStrEnvOrPanic("DB_PASSWORD"),
		DBHost:        GetStrEnvOrPanic("DB_HOST"),
		DBName:        GetStrEnvOrPanic("DB_NAME"),
		IpDBHost:      GetStrEnvOrPanic("IP_DB_HOST"),
		IpDBPort:      GetIntEnvOrPanic("IP_DB_PORT"),
		IpDBUser:      GetStrEnvOrPanic("IP_DB_USER"),
		IpDBPassword:  GetStrEnvOrPanic("IP_DB_PASSWORD"),
		IpDBName:      GetStrEnvOrPanic("IP_DB_NAME"),
		RedisHost:     GetStrEnvOrPanic("REDIS_HOST"),
		RedisPort:     GetIntEnvOrPanic("REDIS_PORT"),
		RedisDB:       GetIntEnvOrPanic("REDIS_DB"),
		RedisPassword: GetStrEnvOrPanic("REDIS_PASSWORD"),
	}
	return env
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
