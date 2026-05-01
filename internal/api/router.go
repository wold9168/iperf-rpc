package api

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(LoggerMiddleware(), CORSMiddleware())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	{
		api.GET("/health", handler.Health)
		api.POST("/iperf/run", handler.RunIperf)
		api.POST("/iperf/server/start", handler.StartServer)
		api.POST("/iperf/server/stop", handler.StopServer)
		api.GET("/iperf/status", handler.ServerStatus)
		api.POST("/http/run", handler.RunHttpTest)
		api.GET("/http/data", handler.ServeHttpData)
		api.POST("/http/upload", handler.ReceiveHttpUpload)
	}

	// 静态文件服务 (前端 dist)
	r.NoRoute(func(c *gin.Context) {
		c.File("./web/dist/index.html")
	})
	r.Static("/assets", "./web/dist/assets")

	return r
}
