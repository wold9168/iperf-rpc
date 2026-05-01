package main

import (
	"log"

	"github.com/wold9168/iperf-rpc/internal/api"
	"github.com/wold9168/iperf-rpc/internal/config"
	"github.com/wold9168/iperf-rpc/internal/iperf"

	_ "github.com/wold9168/iperf-rpc/docs"
)

// @title           iperf-rpc API
// @version         1.0
// @description     基于 iperf3 的网络测速 RPC 服务
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	cfg := config.Load()

	executor := iperf.New()
	httpExecutor := iperf.NewHttpExecutor()
	handler := api.NewHandler(executor, httpExecutor)
	router := api.SetupRouter(handler)

	log.Printf("iperf-rpc server starting on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
