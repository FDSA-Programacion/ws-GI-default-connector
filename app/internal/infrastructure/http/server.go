package httpserver

import (
	"fmt"
	"net/http"
	"ws-int-httr/internal/infrastructure/config"
	"ws-int-httr/internal/infrastructure/http/handlers"
	"ws-int-httr/internal/infrastructure/http/middleware"
	"ws-int-httr/internal/infrastructure/logger"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

type Server struct {
	engine  *gin.Engine
	cfg     config.AppConfig
	handler *handlers.BookingHTTPHandler
}

func New(cfg config.AppConfig, handler *handlers.BookingHTTPHandler) *Server {
	r := gin.New()
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(gin.Recovery())

	srv := Server{
		engine:  r,
		cfg:     cfg,
		handler: handler,
	}
	srv.registerRoutes()
	return &srv
}

func (s *Server) registerRoutes() {
	wsGroup := s.engine.Group("/ws-int-httr")

	// Middlewares
	wsGroup.Use(middleware.SessionMiddleware())
	wsGroup.Use(middleware.CustomMiddleware())
	wsGroup.Use(logger.Logger())

	// HealthChecks
	wsGroup.GET("/health", healthCheckHandler)
	// wsGroup.GET("/healthcheck", healthCheckHandler)

	// Admin endpoints (con autenticación Basic Auth)
	wsGroup.GET("/admin/cache/reload", handlers.ReloadCacheHandler)

	// Login routes
	wsGroup.POST("/availability", s.handler.HandleAvail)
	wsGroup.POST("/prebook", s.handler.HandlePreBook)
	wsGroup.POST("/book", s.handler.HandleBook)
	wsGroup.POST("/cancel", s.handler.HandleCancel)

	wsGroup.POST("/ws", s.handler.HandleWebServiceEndpoint)
}

func healthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func (s *Server) Run() error {
	addr := fmt.Sprintf(":%d", s.cfg.ServerPort())

	return s.engine.Run(addr)
}
