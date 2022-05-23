package config

import "github.com/spf13/viper"

type AppConfig struct {
	AWSConfig        AWSConfig
	HTTPServerConfig HTTPServerConfig
}

// NewAppConfig creates a new AppConfig
func NewAppConfig(v *viper.Viper) AppConfig {
	return AppConfig{
		AWSConfig:        NewAWSConfig(v),
		HTTPServerConfig: NewHTTPServerConfig(v),
	}
}
