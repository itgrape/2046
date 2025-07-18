package main

import (
	"log"
	"slurm-dashboard/config"
	"slurm-dashboard/internal/api"
	"slurm-dashboard/internal/store"
)

func main() {
	// 1. 加载配置
	cfg := config.LoadConfig()

	// 2. 初始化Token存储
	tokenStore := store.NewTokenStore()

	// 3. 初始化路由
	router := api.NewRouter(cfg, tokenStore)

	// 4. 启动服务
	log.Println("Go backend server is running on :" + cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
