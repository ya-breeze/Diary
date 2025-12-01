//nolint:forbidigo // it's okay to use fmt in this file
package config

import (
	"fmt"
	"reflect"

	"github.com/spf13/viper"
)

type Config struct {
	Users                       string `mapstructure:"users" default:""`
	JWTSecret                   string `mapstructure:"jwt_secret" default:""`
	SessionSecret               string `mapstructure:"session_secret" default:""`
	Verbose                     bool   `mapstructure:"verbose" default:"false"`
	Port                        int    `mapstructure:"port" default:"8080"`
	DBPath                      string `mapstructure:"dbpath" default:":memory:"`
	AssetPath                   string `mapstructure:"assetpath" default:"/tmp/diary-assets"`
	DisableImporters            bool   `mapstructure:"disableimporters" default:"false"`
	DisableCurrenciesRatesFetch bool   `mapstructure:"disablecurrenciesratesfetch" default:"false"`
	Issuer                      string `mapstructure:"issuer" default:"diary"`
	CookieName                  string `mapstructure:"cookiename" default:"diarycookie"`
	AllowedOrigins              string `mapstructure:"allowedorigins" default:"http://localhost:4200,http://localhost:8080"`
	DisableRateLimit            bool   `mapstructure:"disableratelimit" default:"false"`

	// Batch upload limits
	MaxPerFileSizeMB    int `mapstructure:"maxperfilesizemb" default:"200"`
	MaxBatchFiles       int `mapstructure:"maxbatchfiles" default:"100"`
	MaxBatchTotalSizeMB int `mapstructure:"maxbatchtotalsizemb" default:"1000"`
}

func InitiateConfig(cfgFile string) (*Config, error) {
	cfg := Config{}

	setDefaultsFromStruct(&cfg)
	viper.SetEnvPrefix("GB")
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
		viper.SetDefault(field.Name, defaultValue)
	}
}
