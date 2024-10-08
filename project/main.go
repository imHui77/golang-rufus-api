package main

import (
	"log"
	"rufus-api/api"
	"rufus-api/middleware"
	"rufus-api/models/mqtt"
	"rufus-api/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化RabbitMQ連接
	rmq, err := mqtt.InitRabbitMQ("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ: %v", err)
	}
	defer rmq.Close()

	// 啟動日誌消費者
	err = services.StartLogConsumer(rmq)
	if err != nil {
		log.Fatalf("Failed to start log consumer: %v", err)
	}

	// 設置Gin路由
	r := gin.Default()

	// 使用日誌中間件
	r.Use(middleware.Logger(rmq))

	// 設置API路由
	api.SetupRoutes(r)

	// 啟動服務器
	r.Run(":8080")
}