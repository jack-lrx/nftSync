package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"log"
	"net/http"
	"strings"
)

var jwtSecret = []byte("your-secret-key") // 建议从配置读取

// 简单 JWT 校验中间件（可扩展）
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//头格式校验
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" || !strings.HasPrefix(tokenStr, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized: missing or invalid token"})
			return
		}
		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

		// 解析和校验 JWT
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			// 校验签名算法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			log.Printf("JWT校验失败: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized: invalid token"})
			return
		}

		// 校验 claims（过期、签发方、受众等）
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if exp, ok := claims["exp"].(float64); ok {
				if int64(exp) < (c.Request.Context().Value("nowUnix").(int64)) {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized: token expired"})
					return
				}
			}
			// 可选: 校验 iss、aud
			// if claims["iss"] != "your-issuer" { ... }
			// if claims["aud"] != "your-audience" { ... }

			// 注入用户信息到 context
			c.Set("user_id", claims["sub"])
			c.Set("role", claims["role"])
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized: invalid claims"})
			return
		}

		c.Next()
	}
}
