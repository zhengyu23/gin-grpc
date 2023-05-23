package initialize

import (
	"net/http"
	"xrUncle/webs/goods-web/middlewares"
	"xrUncle/webs/goods-web/router"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Routers() *gin.Engine {
	Router := gin.Default() // 生成 默认gin (具有Logger和Recovery两种中间件)

	Router.Use(middlewares.Cors()) // 添加中间件, 处理跨域问题

	// 健康检查接口
	Router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"success": true,
		})
	})

	ApiGroup := Router.Group("/v1")
	router.InitGoodsRouter(ApiGroup)
	router.InitCategoryRouter(ApiGroup)
	router.InitBannerRouter(ApiGroup)
	router.InitBrandRouter(ApiGroup)

	zap.S().Info("配置用户相关的url")
	return Router
}
