package middlewares

import (
	"net/http"
	"xrUncle/webs/goods-web/models"

	"github.com/gin-gonic/gin"
)

// IsAdminAuth 判断当前用户是否为管理员
// 前置条件是以及JWT验证
func IsAdminAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims, _ := ctx.Get("claims")
		currentUser := claims.(*models.CustomClaims)

		if currentUser.AuthorityId != 2 {
			ctx.JSON(http.StatusForbidden, gin.H{
				"msg": "无权限",
			})
			ctx.Abort() // 中止
			return
		}
		ctx.Next()
	}
}
