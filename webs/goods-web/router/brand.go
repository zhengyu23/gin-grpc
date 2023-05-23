package router

import (
	"xrUncle/webs/goods-web/api/brands"
	"xrUncle/webs/goods-web/middlewares"

	"github.com/gin-gonic/gin"
)

// InitBrandRouter
// 1. 商品的api接口开发完成
// 2. 图片的坑
func InitBrandRouter(Router *gin.RouterGroup) {
	// Trace() 为该组添加链路追踪
	BrandRouter := Router.Group("brands").Use(middlewares.Trace())
	{
		BrandRouter.GET("", brands.BrandList)          // 获取品牌列表
		BrandRouter.POST("", brands.NewBrand)          // 新建品牌
		BrandRouter.DELETE("/:id", brands.DeleteBrand) // 删除品牌
		BrandRouter.PUT("/:id", brands.UpdateBrand)    // 修改品牌信息
	}

	CategoryBrandRouter := Router.Group("categorybrands")
	{
		CategoryBrandRouter.GET("", brands.CategoryBrandList)            // 获取分类品牌聚合列表
		CategoryBrandRouter.GET("/:id", brands.GetBrandListByCategoryId) // 获取分类下的品牌列表
		CategoryBrandRouter.POST("", brands.NewCategoryBrand)            // 新增分类品牌聚合信息
		CategoryBrandRouter.DELETE("/:id", brands.DeleteCategoryBrand)   // 删除分类品牌聚合信息
		CategoryBrandRouter.PUT("/:id", brands.UpdateCategoryBrand)      // 修改分类品牌聚合信息
	}
}
