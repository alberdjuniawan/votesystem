package middleware

import (
	"context"

	"github.com/alberdjuniawan/votesystem/internal/shared/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader("X-Request-ID")
		if reqID == "" {
			reqID = "req_" + uuid.New().String()
		}

		c.Set("request_id", reqID)

		ctx := context.WithValue(c.Request.Context(), logger.ReqIDKey, reqID)
		c.Request = c.Request.WithContext(ctx)

		c.Header("X-Request-ID", reqID)

		c.Next()
	}
}
