package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Cors option请求 —— 处理跨域问题
func Cors() gin.HandlerFunc {
	// 添加请求头 - Allow Origin
	//			- Allow Headers
	//			- Allow Methods
	//			- Expose Headers
	//			- Allow-Credentials
	return func(c *gin.Context) {
		method := c.Request.Method

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token, x-token")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PATCH, PUT")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
	}
}
