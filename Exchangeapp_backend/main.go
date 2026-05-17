package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"exchangeapp/config"
	"exchangeapp/router"
)

// main 是后端服务入口：加载配置、注册路由并优雅退出。
func main() {
	// 初始化配置、数据库、Redis 连接。
	config.InitConfig()

	// 装配路由（包含公开接口和鉴权接口分组）。
	r := router.SetupRouter()

	port := config.AppConfig.App.Port

	if port == "" {
		port = ":8080"
	}

	// 创建 HTTP 服务实例。
	srv := &http.Server{
		Addr:    port,
		Handler: r,
	}

	// 后台启动服务。
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 捕获中断信号并执行优雅停机。
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}
