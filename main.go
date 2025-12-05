package main

import (
	"github.com/adiecho/oci-panel/internal/config"
	"github.com/adiecho/oci-panel/internal/database"
	"github.com/adiecho/oci-panel/internal/router"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	cfg := config.Load()

	if err := database.InitDB(cfg.Database.DSN); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	r := gin.Default()
	services := router.Setup(r, cfg)

	// 启动定时任务服务
	services.Scheduler.Start()
	defer services.Scheduler.Stop()

	// 启动创建实例任务服务
	services.Task.Start()
	defer services.Task.Stop()

	// 启动 Telegram Bot（如果已配置并启用）
	_, _, tgEnabled := services.Telegram.GetConfig()
	if tgEnabled {
		services.Telegram.StartBot()
		defer services.Telegram.StopBot()
	}

	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
