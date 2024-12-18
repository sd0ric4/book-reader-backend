package config_test

import (
	"testing"

	"github.com/sd0ric4/book-reader-backend/app/config"
)

func TestLoadConfig(t *testing.T) {

	// 调用 LoadConfig 函数
	config.LoadConfig("../../../config/config.yaml")

	// 判断配置文件是否加载成功
	if config.Config == nil {
		t.Error("config is nil")
	}

	// 判断配置文件中的数据库配置是否加载成功
	if config.Config.MySQL.DBName == "" {
		t.Error("mysql config is not loaded")
	}
	if config.Config.MySQL.Host == "" {
		t.Error("mysql config is not loaded")
	}
	if config.Config.MySQL.Port == 0 {
		t.Error("mysql config is not loaded")
	}
	if config.Config.MySQL.User == "" {
		t.Error("mysql config is not loaded")
	}
	if config.Config.MySQL.Password == "" {
		t.Error("mysql config is not loaded")
	}

	// 判断配置文件中的服务器配置是否加载成功
	if config.Config.Server.Port == 0 {
		t.Error("server config is not loaded")
	}

	// 判断配置文件中的 Redis 配置是否加载成功
	if config.Config.Redis.Host == "" {
		t.Error("redis config is not loaded")
	}
	if config.Config.Redis.Port == 0 {
		t.Error("redis config is not loaded")
	}

	// 输出配置文件中的配置
	t.Logf("config: %+v", config.Config)
}

func TestMySQLConfig(t *testing.T) {
	config.LoadConfig("../../../config/config.yaml")

	if config.Config.MySQL.DBName == "" {
		t.Error("mysql dbname is not loaded")
	}
	if config.Config.MySQL.Host == "" {
		t.Error("mysql host is not loaded")
	}
	if config.Config.MySQL.Port == 0 {
		t.Error("mysql port is not loaded")
	}
	if config.Config.MySQL.User == "" {
		t.Error("mysql user is not loaded")
	}
	if config.Config.MySQL.Password == "" {
		t.Error("mysql password is not loaded")
	}
}

func TestServerConfig(t *testing.T) {
	config.LoadConfig("../../../config/config.yaml")

	if config.Config.Server.Port == 0 {
		t.Error("server port is not loaded")
	}
}

func TestRedisConfig(t *testing.T) {
	config.LoadConfig("../../../config/config.yaml")

	if config.Config.Redis.Host == "" {
		t.Error("redis host is not loaded")
	}
	if config.Config.Redis.Port == 0 {
		t.Error("redis port is not loaded")
	}
}
