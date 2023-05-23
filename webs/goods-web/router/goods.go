package router

import (
	"xrUncle/webs/goods-web/middlewares"

	"github.com/gin-gonic/gin"

	"xrUncle/webs/goods-web/api/goods"
)

func InitGoodsRouter(Router *gin.RouterGroup) {
	// GoodsRouter := Router.Group("goods").Use(middlewares.Trace())
	// Trace() 为该组添加链路追踪
	GoodsRouter := Router.Group("goods").Use(middlewares.Trace())
	{
		GoodsRouter.GET("", goods.List)              //商品列表
		GoodsRouter.GET("/:id", goods.Detail)        //获取商品的详情
		GoodsRouter.GET("/:id/stocks", goods.Stocks) //获取商品的库存

		// 以下接口添加了 登录鉴权+管理员
		GoodsRouter.POST("", middlewares.JWTAuth(), middlewares.IsAdminAuth(), goods.New)
		GoodsRouter.DELETE("/:id", middlewares.JWTAuth(), middlewares.IsAdminAuth(), goods.Delete)
		GoodsRouter.PUT("/:id", middlewares.JWTAuth(), middlewares.IsAdminAuth(), goods.Update)
		GoodsRouter.PATCH("/:id", middlewares.JWTAuth(), middlewares.IsAdminAuth(), goods.UpdateStatus)
	}
}
