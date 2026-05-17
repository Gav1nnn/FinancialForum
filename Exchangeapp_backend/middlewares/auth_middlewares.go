package middlewares

import (
	"exchangeapp/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 校验 Authorization 头中的 JWT，并注入用户名到上下文。
func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("Authorization")
		if token == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Misssing Authorization Header"})
			ctx.Abort()
			return
		}
		username, err := utils.ParseJWT(token)

		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invaild token"})
			ctx.Abort()
			return
		}

		ctx.Set("username", username)
		ctx.Next()

	}
}
