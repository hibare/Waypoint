package config

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	commonLogger "github.com/hibare/GoCommon/v2/pkg/logger"
	"github.com/hibare/Waypoint/internal/constants"
	"github.com/spf13/viper"
)

var (
	// ErrSecretKeyEmpty indicates that the secret key is empty.
	ErrSecretKeyEmpty = errors.New("secret key is empty")

	// ErrDBLicenseKeyEmpty indicates that the database license key is empty.
	ErrAssetDirEmpty = errors.New("asset directory cannot be empty")
)

// Environment represents the application environment.
type Environment string

const (
	// EnvironmentProduction represents the production environment.
	EnvironmentProduction Environment = "production"

	// EnvironmentDevelopment represents the development environment.
	EnvironmentDevelopment Environment = "development"

	// EnvironmentTesting represents the testing environment.
	EnvironmentTesting Environment = "testing"
)

const (
	DefaultAssetDirPath = "./data"
)

type CoreConfig struct {
	Environment Environment `mapstructure:"environment"`
	SecretKey   string      `mapstructure:"secret_key"`
	DataDir     string      `mapstructure:"data_dir"`
}

func (c *CoreConfig) PostProcess() {
	// Resolve asset dir path to absolute path
	absPath, err := filepath.Abs(c.DataDir)
	if err == nil {
		c.DataDir = absPath
	}
}

func (c *CoreConfig) Validate() error {
	if c.SecretKey == "" {
		return ErrSecretKeyEmpty
	}

	if c.DataDir == "" {
		return ErrAssetDirEmpty
	}
	return nil
}

// Config holds the entire application configuration.
type Config struct {
	Core    CoreConfig    `mapstructure:"core"`
	Server  ServerConfig  `mapstructure:"server"`
	DB      DBConfig      `mapstructure:"db"`
	MaxMind MaxMindConfig `mapstructure:"maxmind"`
	Logger  LoggerConfig  `mapstructure:"logger"`
	OIDC    OIDCConfig    `mapstructure:"oidc"`
}

// Validate validates the entire configuration.
func (c *Config) Validate() error {
	// Skip OIDC validation if not configured
	if c.OIDC.IssuerURL != "" {
		if err := c.OIDC.Validate(); err != nil {
			return err
		}
	}

	var vFuncs = []func() error{
		c.DB.Validate,
		c.Core.Validate,
		c.MaxMind.Validate,
		c.Server.Validate,
		c.Logger.Validate,
	}

	for _, vf := range vFuncs {
		if err := vf(); err != nil {
			return err
		}
	}

	return nil
}

// PostProcess performs post-processing on the configuration.
func (c *Config) PostProcess() {
	c.Core.PostProcess()
	c.Server.PostProcess()
}

func (c *Config) getViper(ctx context.Context, configPath string) *viper.Viper {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Config search paths.
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
		v.AddConfigPath("/etc/waypoint/")
	}

	// Environment variable binding.
	v.SetEnvPrefix(strings.ToUpper(constants.ProgramIdentifier))
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	// Environment variable bindings.
	// With SetEnvPrefix("WAYPOINT") + SetEnvKeyReplacer("."→"_") + AutomaticEnv(),
	// each key maps to WAYPOINT_<UPPER_KEY> automatically (e.g. core.secret_key → WAYPOINT_CORE_SECRET_KEY).
	envKeys := []string{
		"core.environment",
		"core.secret_key",
		"core.data_dir",
		"db.dsn",
		"db.max_idle_conn",
		"db.max_open_conn",
		"db.conn_max_lifetime",
		"server.listen_addr",
		"server.listen_port",
		"server.base_url",
		"server.write_timeout",
		"server.read_timeout",
		"server.idle_timeout",
		"server.wait_timeout",
		"server.request_timeout",
		"server.cert_file",
		"server.key_file",
		"logger.level",
		"logger.mode",
		"maxmind.license_key",
		"maxmind.auto_update",
		"maxmind.auto_update_interval",
		"oidc.issuer_url",
		"oidc.client_id",
		"oidc.client_secret",
	}

	for _, key := range envKeys {
		if err := v.BindEnv(key); err != nil {
			slog.WarnContext(ctx, "Failed to bind environment variable",
				slog.String("config", key),
				slog.String("error", err.Error()))
		}
	}

	// Set default values.
	v.SetDefault("core.environment", EnvironmentProduction)
	v.SetDefault("core.data_dir", DefaultAssetDirPath)
	v.SetDefault("db.max_idle_conn", DefaultDBMaxIdleConn)
	v.SetDefault("db.max_open_conn", DefaultDBMaxOpenConn)
	v.SetDefault("db.conn_max_lifetime", DefaultDBConnMaxLifetime)
	v.SetDefault("server.listen_addr", DefaultServerListenAddr)
	v.SetDefault("server.listen_port", DefaultServerListenPort)
	v.SetDefault("server.base_url", "http://localhost:5000")
	v.SetDefault("server.read_timeout", DefaultServerReadTimeout)
	v.SetDefault("server.write_timeout", DefaultServerWriteTimeout)
	v.SetDefault("server.idle_timeout", DefaultServerIdleTimeout)
	v.SetDefault("server.wait_timeout", DefaultServerWaitTimeout)
	v.SetDefault("server.request_timeout", DefaultServerRequestTimeout)
	v.SetDefault("logger.level", commonLogger.LogLevelInfo)
	v.SetDefault("logger.mode", commonLogger.LogModePretty)
	v.SetDefault("maxmind.license_key", "")
	v.SetDefault("maxmind.auto_update", DefaultMaxMindAutoUpdate)
	v.SetDefault("maxmind.auto_update_interval", DefaultMaxMindAutoUpdateInterval)
	v.SetDefault("oidc.issuer_url", "")
	v.SetDefault("oidc.client_id", "")
	v.SetDefault("oidc.client_secret", "")

	return v
}

// Current holds the current application configuration.
var Current *Config

// Load loads the configuration from environment variables and/or config file.
func Load(ctx context.Context, configPath string) (*Config, error) {
	cfg := &Config{}
	v := cfg.getViper(ctx, configPath)

	// Try read config.
	if err := v.ReadInConfig(); err != nil {
		if _, ok := errors.AsType[viper.ConfigFileNotFoundError](err); ok {
			slog.WarnContext(ctx, "No config file found, relying on env vars/defaults")
		} else {
			return nil, err
		}
	} else {
		slog.DebugContext(ctx, "Using config file", slog.String("file", v.ConfigFileUsed()))
	}

	// Unmarshal into config.
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Post-process config.
	cfg.PostProcess()

	// Initialize logger.
	commonLogger.InitLogger(&cfg.Logger.Level, &cfg.Logger.Mode)

	// Create data dir
	if err := os.MkdirAll(cfg.Core.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDataDirCreation, err)
	}

	Current = cfg

	slog.InfoContext(ctx, "Loaded config")
	return cfg, nil
}

// LoadDefault loads the configuration with default settings (no context required).
func LoadDefault() error {
	ctx := context.Background()
	_, err := Load(ctx, "")
	return err
}

// GenerateConfigFile generates a config file with default values.
func GenerateConfigFile(ctx context.Context, configPath string) (string, error) {
	cfg := &Config{}
	v := cfg.getViper(ctx, configPath)

	// Unmarshal viper's defaults into the config struct
	if err := v.Unmarshal(cfg); err != nil {
		return "", fmt.Errorf("%w: %w", ErrConfigUnmarshal, err)
	}

	return v.ConfigFileUsed(), v.WriteConfig()
}
