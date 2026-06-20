//nolint:forbidigo // it's okay to use fmt in this file
package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	JWTSecret        string `mapstructure:"jwt_secret" default:""`
	CookieSecure     bool   `mapstructure:"cookie_secure" default:"true"`
	SeedUsers        string `mapstructure:"seed_users" default:""`
	Verbose          bool   `mapstructure:"verbose" default:"false"`
	Port             int    `mapstructure:"port" default:"8080"`
	DataPath         string `mapstructure:"datapath" default:"./diary-data"`
	AllowedOrigins   string `mapstructure:"allowedorigins" default:"http://localhost:4200,http://localhost:8080"`
	DisableRateLimit bool   `mapstructure:"disableratelimit" default:"false"`

	// Batch upload limits
	MaxPerFileSizeMB    int `mapstructure:"maxperfilesizemb" default:"200"`
	MaxBatchFiles       int `mapstructure:"maxbatchfiles" default:"100"`
	MaxBatchTotalSizeMB int `mapstructure:"maxbatchtotalsizemb" default:"1000"`

	// Health check
	HealthCheckInterval string `mapstructure:"health_check_interval" default:"24h"`

	// Backup
	BackupInterval string `mapstructure:"backup_interval" default:"24h"`
	BackupMaxCount int    `mapstructure:"backup_max_count" default:"10"`

	// AI tagging — confidence threshold τ above which suggestions may be
	// auto-applied to untagged days when a family enables auto mode.
	AITaggingThreshold float64 `mapstructure:"ai_tagging_threshold" default:"0.8"`
}

func InitiateConfig(cfgFile string) (*Config, error) {
	cfg := Config{}

	setDefaultsFromStruct(&cfg)
	viper.SetEnvPrefix("DIARY")
	viper.AutomaticEnv()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err == nil {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
	}

	// Unmarshal the config into the Config struct
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if cfg.Verbose {
		fmt.Printf("Config: %+v\n", cfg)
	}

	return &cfg, nil
}

func setDefaultsFromStruct(s interface{}) {
	val := reflect.ValueOf(s).Elem()
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		defaultValue := field.Tag.Get("default")
		mapKey := field.Tag.Get("mapstructure")
		if mapKey == "" {
			mapKey = strings.ToLower(field.Name)
		}
		viper.SetDefault(mapKey, defaultValue)
		// AutomaticEnv doesn't reliably find keys with underscores (e.g. jwt_secret →
		// DIARY_JWT_SECRET), so bind each key explicitly.
		_ = viper.BindEnv(mapKey)
	}
}
