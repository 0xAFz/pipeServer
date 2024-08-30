package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	CassandraHost     string
	CassandraKeyspace string
	Token             string
	ServerAddr        string
	ClientURL         string
	ProxyAddr         string
}

var AppConfig *Config

func LoadConfig() {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	AppConfig = &Config{
		CassandraHost:     viper.GetString("CASSANDRA_HOST"),
		CassandraKeyspace: viper.GetString("CASSANDRA_KEYSPACE"),
		Token:             viper.GetString("TOKEN"),
		ServerAddr:        viper.GetString("SERVER_ADDR"),
		ClientURL:         viper.GetString("CLIENT_URL"),
		ProxyAddr:         viper.GetString("PROXY_ADDR"),
	}
}
