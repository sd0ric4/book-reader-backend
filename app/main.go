package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sd0ric4/book-reader-backend/app/config"
	"github.com/sd0ric4/book-reader-backend/app/database"
	"github.com/sd0ric4/book-reader-backend/app/routes"
)

func main() {

	// 加载配置
	config.LoadConfig("../config/config.yaml")

	// 初始化数据库
	database.InitMySQL()
	// 创建Gin实例
	r := gin.Default()
	// 配置 CORS
	r.Use(cors.New(cors.Config{
		// 允许所有源
		AllowOrigins: []string{"*"},

		// 更安全的配置
		// AllowOrigins: []string{
		//     "http://localhost:3000",
		//     "https://yourdomain.com"
		// },

		// 允许方法
		AllowMethods: []string{
			"GET", "POST", "PUT", "DELETE", "OPTIONS",
		},

		// 允许头
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
		},

		// 是否允许发送凭证
		AllowCredentials: true,

		// 最大存活时间
		MaxAge: 12 * time.Hour,
	}))
	gin.ForceConsoleColor()
	// 初始化路由
	routes.SetupRoutes(r)

	// 启动服务
	err := r.Run(fmt.Sprintf("%s:%d", config.Config.Server.Host, config.Config.Server.Port))
	if err != nil {
		log.Fatal("Failed to start the server:", err)
	}
}
