package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Channel struct {
	ID        int    `yaml:"id"`
	Code      string `yaml:"code"`
	AuthToken string `yaml:"authToken"`
}

type YamlConfig struct {
	ServerPortValue         int    `yaml:"serverPort"`
	LogLevelValue           string `yaml:"logLevel"`
	LogPathValue            string `yaml:"logPath"`
	ProviderNameValue       string `yaml:"providerName"`
	ProviderCodeValue       string `yaml:"providerCode"`
	BookingCodeVersionValue string `yaml:"internalBookingCodeVersion"`
	AvailToValue            int    `yaml:"availTO"`
	AdminUsernameValue      string `yaml:"adminUsername"`
	AdminPasswordValue      string `yaml:"adminPassword"`
	DefaultEmailValue       string `yaml:"defaultEmail"`
	DefaultPhoneValue       string `yaml:"defaultPhone"`

	Channels           []Channel `yaml:"channels"`
	DBYamlConfig       `yaml:"database"`
	ProviderYamlConfig `yaml:"provider"`
}

type DBYamlConfig struct {
	Host          string `yaml:"host"`
	Port          int    `yaml:"port"`
	User          string `yaml:"username"`
	Pass          string `yaml:"password"`
	Driver        string `yaml:"driver"`
	SID           string `yaml:"SID"`
	FetchRowCount int    `yaml:"fetchRowCount"`
}

type ProviderYamlConfig struct {
	SearchURL            string    `yaml:"searchURL"`
	QuoteURL             string    `yaml:"quoteURL"`
	BookURL              string    `yaml:"bookURL"`
	CancelURL            string    `yaml:"cancelURL"`
	AuthToken            string    `yaml:"authToken"`
	UsernameB2B          string    `yaml:"usernameB2B"`
	PasswordB2B          string    `yaml:"passwordB2B"`
	UsernameB2C          string    `yaml:"usernameB2C"`
	PasswordB2C          string    `yaml:"passwordB2C"`
	TimeoutMs            int       `yaml:"timeoutMs"`
	MaxRoomsPerOccupancy int       `yaml:"maxRoomsPerOccupancy"`
	Channels             []Channel `yaml:"-"`
}

func (c *YamlConfig) ProviderName() string { return c.ProviderNameValue }
func (c *YamlConfig) ProviderCode() string { return c.ProviderCodeValue }

func (c *YamlConfig) ServerPort() int            { return c.ServerPortValue }
func (c *YamlConfig) LogLevel() string           { return c.LogLevelValue }
func (c *YamlConfig) LogPath() string            { return c.LogPathValue }
func (c *YamlConfig) BookingCodeVersion() string { return c.BookingCodeVersionValue }
func (c *YamlConfig) AvailTo() int               { return c.AvailToValue }
func (c *YamlConfig) AdminUsername() string      { return c.AdminUsernameValue }
func (c *YamlConfig) AdminPassword() string      { return c.AdminPasswordValue }
func (c *YamlConfig) DefaultEmail() string       { return c.DefaultEmailValue }
func (c *YamlConfig) DefaultPhone() string       { return c.DefaultPhoneValue }

func (c *DBYamlConfig) DBHost() string      { return c.Host }
func (c *DBYamlConfig) DBPort() int         { return c.Port }
func (c *DBYamlConfig) DBUser() string      { return c.User }
func (c *DBYamlConfig) DBPass() string      { return c.Pass }
func (c *DBYamlConfig) DBDriver() string    { return c.Driver }
func (c *DBYamlConfig) DBSID() string       { return c.SID }
func (c *DBYamlConfig) DBFechRowCount() int { return c.FetchRowCount }

func (c *ProviderYamlConfig) ProviderSearchURL() string {
	return c.SearchURL
}
func (c *ProviderYamlConfig) ProviderQuoteURL() string {
	return c.QuoteURL
}
func (c *ProviderYamlConfig) ProviderBookURL() string {
	return c.BookURL
}
func (c *ProviderYamlConfig) ProviderCancelURL() string {
	return c.CancelURL
}
func (c *ProviderYamlConfig) ProviderAuthToken() string {
	return c.AuthToken
}
func (c *ProviderYamlConfig) ProviderTimeoutMs() int { return c.TimeoutMs }

// ProviderIdList extrae los IDs de los canales configurados
func (c *ProviderYamlConfig) ProviderIdList() []int {
	ids := make([]int, 0, len(c.Channels))
	for _, channel := range c.Channels {
		ids = append(ids, channel.ID)
	}
	return ids
}

func (c *ProviderYamlConfig) ProviderMaxRoomsPerOccupancy() int { return c.MaxRoomsPerOccupancy }

// ProviderAuthForChannel devuelve el authToken del canal según su código.
func (c *YamlConfig) ProviderAuthForChannel(channelCode string) string {
	normalizedCode := strings.TrimSpace(strings.ToUpper(channelCode))

	// Buscar en la lista de canales
	for _, channel := range c.Channels {
		if strings.TrimSpace(strings.ToUpper(channel.Code)) == normalizedCode {
			return channel.AuthToken
		}
	}

	return ""
}

func LoadYamlConfig(path string) (AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error leyendo archivo: %w", err)
	}
	var cfg YamlConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error parseando YAML: %w", err)
	}

	// Copiar los canales del nivel raíz al ProviderYamlConfig
	// para que ProviderIdList() y los métodos de autenticación puedan acceder a ellos
	cfg.ProviderYamlConfig.Channels = cfg.Channels

	return &cfg, nil
}
