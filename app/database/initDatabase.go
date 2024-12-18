package database

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/sd0ric4/book-reader-backend/app/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	MySQLDB *gorm.DB
	RedisDB *redis.Client
)

func InitDatabases() {
	InitMySQL()
	InitRedis()
}

func InitMySQL() {
	cfg := config.Config.MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	var err error
	MySQLDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting to MySQL: %s", err)
	}
}

func InitRedis() {
	cfg := config.Config.Redis
	ctx := context.Background() // 创建上下文
	RedisDB = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
	})

	_, err := RedisDB.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Error connecting to Redis: %s", err)
	}
}
