package handlers

import (
	"encoding/json"
	"net/http"

	"ws-int-httr/internal/domain"
	"ws-int-httr/internal/infrastructure/config"
	"ws-int-httr/internal/infrastructure/logger"
	"ws-int-httr/internal/infrastructure/persistence"
	"ws-int-httr/internal/infrastructure/registry"

	"github.com/gin-gonic/gin"
)

// ReloadCacheHandler maneja la petición de recarga de caché con autenticación Basic Auth
func ReloadCacheHandler(c *gin.Context) {
	logger.Infof("", "Reloading cache...")

	// Extraer configuración del registry
	cfg, exists := registry.Get[config.AppConfig]("config")
	if !exists {
		logger.Errorf("", "Config not found in registry")
		c.JSON(http.StatusInternalServerError, domain.ReloadCacheRS{Status: "KO"})
		return
	}

	// Validar autenticación Basic Auth
	username, password, ok := c.Request.BasicAuth()
	if !ok {
		logger.Infof("", "Cache reload attempt without credentials")
		c.Header("WWW-Authenticate", `Basic realm="Admin Area"`)
		c.JSON(http.StatusUnauthorized, domain.ReloadCacheRS{Status: "KO"})
		return
	}

	if username != cfg.AdminUsername() || password != cfg.AdminPassword() {
		logger.Infof("", "Cache reload attempt with invalid credentials (user: %s)", username)
		c.JSON(http.StatusUnauthorized, domain.ReloadCacheRS{Status: "KO"})
		return
	}

	// Recargar caché
	repo, exists := registry.Get[persistence.RepositoryService]("repository")
	if !exists {
		logger.Errorf("", "Repository service not found in registry")
		c.JSON(http.StatusInternalServerError, domain.ReloadCacheRS{Status: "KO"})
		return
	}

	// Obtener cache service del repository y recargar
	cacheService := repo.GetCacheService()
	if cacheService == nil {
		logger.Errorf("", "Cache service not available")
		c.JSON(http.StatusInternalServerError, domain.ReloadCacheRS{Status: "KO"})
		return
	}

	// Recargar todos los cachés
	persistence.LoadAllCache(cacheService, cfg.ProviderIdList()...)

	logger.Infof("", "Cache reloaded successfully!")

	rs := domain.ReloadCacheRS{Status: "OK"}
	responseData, _ := json.Marshal(&rs)

	c.Data(http.StatusOK, "application/json", responseData)
}
