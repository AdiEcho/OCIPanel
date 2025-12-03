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
	schedulerService := router.Setup(r, cfg)

	// 启动定时任务服务
	schedulerService.Start()
	defer schedulerService.Stop()

	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
