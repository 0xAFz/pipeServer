package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	RedisHost         string
	CassandraHost     string
	CassandraKeyspace string
	Token             string
	ServerAddr        string
	ClientURL         string
	ProxyAddr         string
}

var AppConfig *Config

func LoadConfig() {
	env := os.Getenv("GO_ENV")
	if env == "dev" {
		viper.SetConfigFile(".env")
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("Error reading .env file: %v", err)
		}
	}

	viper.AutomaticEnv()

	AppConfig = &Config{
		RedisHost:         viper.GetString("REDIS_HOST"),
		CassandraHost:     viper.GetString("CASSANDRA_HOST"),
		CassandraKeyspace: viper.GetString("CASSANDRA_KEYSPACE"),
		Token:             viper.GetString("TOKEN"),
		ServerAddr:        viper.GetString("SERVER_ADDR"),
		ClientURL:         viper.GetString("CLIENT_URL"),
		ProxyAddr:         viper.GetString("PROXY_ADDR"),
	}
}
