package middleware

import (
	"io"
	"strings"
	"time"

	"ws-int-httr/internal/infrastructure/http/handlers/generic"
	"ws-int-httr/internal/infrastructure/session"

	"github.com/gin-gonic/gin"
)

func SessionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session.New(c.Request.Context())

		c.Next()

		session.Clear()
	}
}

func CustomMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionCtx := session.FromContext()

		rawData, err := c.GetRawData()
		if err == nil && len(rawData) > 0 {
			sessionCtx.Data().RequestRawData = rawData
			c.Request.Body = io.NopCloser(strings.NewReader(string(rawData)))
		}

		sessionCtx.Data().StartTime = time.Now()

		c.Next()

		go generic.IncrementCounter("URI#"+c.Request.RequestURI, generic.TAG_COUNTER_TOTAL)
	}
}
