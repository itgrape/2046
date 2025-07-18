package api

import (
	"net/http"
	"path"
	"slurm-dashboard/config"
	"slurm-dashboard/internal/store"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(cfg *config.Config, tokenStore *store.TokenStore) *gin.Engine {
	router := gin.Default()

	// CORS 配置
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")
	router.Use(cors.New(corsConfig))

	// 静态文件服务
	router.StaticFS("/assets", http.Dir("./static/assets"))
	router.NoRoute(func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.File(path.Join("./static", "index.html"))
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "API route not found"})
		}
	})

	// 公开的登录路由
	router.POST("/api/login", LoginHandler(cfg, tokenStore))

	// 受保护的API v1路由组
	apiV1 := router.Group("/api/v1")
	apiV1.Use(AuthMiddleware(cfg))
	{
		apiV1.GET("/cluster/status", GetClusterStatusHandler(cfg, tokenStore))
		apiV1.GET("/jobs", GetJobsHandler(cfg, tokenStore))
		jobGroup := apiV1.Group("/job")
		{
			jobGroup.POST("/submit", SubmitJobHandler(cfg, tokenStore))
			jobGroup.POST("/allocate", AllocateJobHandler(cfg, tokenStore))
			jobGroup.DELETE("/:job_id", HandleDeleteJob(cfg, tokenStore))
		}
	}

	return router
}
