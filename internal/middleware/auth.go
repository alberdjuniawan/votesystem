package middleware

import (
	"strings"

	jwtpkg "github.com/alberdjuniawan/votesystem/internal/shared/jwt"
	"github.com/alberdjuniawan/votesystem/internal/shared/response"
	"github.com/gin-gonic/gin"
)

func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.NewError(c, response.ErrUnauthorized, nil)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.NewError(c, response.ErrUnauthorized, "invalid authorization format")
			c.Abort()
			return
		}

		claims, err := jwtpkg.ValidateToken(parts[1], jwtSecret)
		if err != nil {
			response.NewError(c, response.ErrUnauthorized, err.Error())
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

func GetUserID(c *gin.Context) string {
	return c.GetString("user_id")
}

func GetUserRole(c *gin.Context) string {
	return c.GetString("user_role")
}
