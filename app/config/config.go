package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

var Config *ConfigStruct

type ConfigStruct struct {
	Server ServerConfig `yaml:"server"`
	MySQL  MySQLConfig  `yaml:"mysql"`
	Redis  RedisConfig  `yaml:"redis"`
	JWT    JWTConfig    `yaml:"jwt"`
	S3     S3           `yaml:"s3"`
}

func LoadConfig(configPath string) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	err = yaml.Unmarshal(data, &Config)
	if err != nil {
		log.Fatalf("Error unmarshalling config file, %s", err)
	}
}
