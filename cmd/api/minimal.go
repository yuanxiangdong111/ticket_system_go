package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("程序启动 - 2026-03-18")
	fmt.Println("=====================")

	// 简化版初始化
	r := gin.Default()

	// 测试路由
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"code":    200,
			"message": "服务正常",
			"time":    time.Now(),
		})
	})

	r.GET("/", func(c *gin.Context) {
		c.String(200, "门票系统")
	})

	// 启动服务
	fmt.Println("正在启动服务器...")
	fmt.Println("访问: http://localhost:8080/health")

	if err := r.Run(":8080"); err != nil {
		fmt.Printf("服务器启动失败: %v\n", err)
	}
}
