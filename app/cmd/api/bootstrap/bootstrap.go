package bootstrap

import (
	"fmt"
	"os"

	"ws-int-httr/internal/application"
	"ws-int-httr/internal/domain"
	"ws-int-httr/internal/infrastructure/config"
	httpserver "ws-int-httr/internal/infrastructure/http"
	"ws-int-httr/internal/infrastructure/http/handlers"
	"ws-int-httr/internal/infrastructure/httr_client"
	"ws-int-httr/internal/infrastructure/logger"
	"ws-int-httr/internal/infrastructure/persistence"
	"ws-int-httr/internal/infrastructure/persistence/cache"
	"ws-int-httr/internal/infrastructure/registry"
	"ws-int-httr/internal/infrastructure/serializer"
)

func Run() error {
	cfg, err := config.LoadYamlConfig("application.yml")
	if err != nil {
		customErr := domain.ErrorConfigNotFound
		customErr.Err = fmt.Errorf("error cargando configuración: %v", err)
		panic(customErr)
	}

	// Configure logging
	logger.InitLog(cfg.LogLevel(), cfg.LogPath())
	// logger.Infof("", "Loaded configuration: %+v", maskConfig(cfg))
	logger.Infof("", "Loaded configuration: %+v", cfg)

	// Conecta a base de datos
	if dbCfg, ok := cfg.(config.DBConfig); ok {
		if err := persistence.InitDB(dbCfg); err != nil {
			logger.Errorf("", "Failed to initialize db: %v", err)
			return err
		}
	} else {
		logger.Errorf("", "AppConfig no implementa DBConfig")
		return fmt.Errorf("AppConfig no implementa DBConfig")
	}

	cacheService := cache.NewCacheService()
	persistence.LoadAllCache(cacheService, cfg.ProviderIdList()...)

	repositoryService := persistence.NewRepositoryService(cacheService)

	// Initialize structured logger
	structuredLogger, err := logger.NewFileStructuredLogger(cfg.LogPath())
	if err != nil {
		logger.Errorf("", "Failed to initialize structured logger: %v", err)
	} else {
		registry.Register("structuredLogger", structuredLogger)
	}

	// Registramos servicios disponibles en toda la aplicación
	registry.Register("repository", repositoryService)
	registry.Register("config", cfg)

	serializerSer := serializer.NewGoSerializer()

	// Initialize driven output adapters
	otClient := httr_client.NewHttrClientImpl(cfg, serializerSer)

	// Initialize Application Logic (Service)
	bookingService := application.NewBookingService(otClient)

	// Initialize DRIVING Inbound/Input Adapters
	handler := handlers.NewBookingHTTPHandler(bookingService, serializerSer)

	// Create and Configure the HTTP Server (Gin)
	server := httpserver.New(cfg, handler)

	host := getHostName()
	logger.Infof("", "Starting server on %v:%v", host, cfg.ServerPort())

	return server.Run()
}

func maskConfig(cfg config.AppConfig) map[string]interface{} {
	masked := map[string]interface{}{
		"serverPort":         cfg.ServerPort(),
		"logLevel":           cfg.LogLevel(),
		"logPath":            cfg.LogPath(),
		"bookingCodeVersion": cfg.BookingCodeVersion(),
		"providerSearchURL":  cfg.ProviderSearchURL(),
		"providerQuoteURL":   cfg.ProviderQuoteURL(),
		"providerBookURL":    cfg.ProviderBookURL(),
		"providerCancelURL":  cfg.ProviderCancelURL(),
	}
	return masked
}

func getHostName() string {
	host, err := os.Hostname()
	if err != nil {
		logger.Errorf("", "Failed to get hostname: %v", err)
		return "localhost"
	}
	return host
}
