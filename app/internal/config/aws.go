package config

import (
	"github.com/spf13/viper"
)

type AWSS3Config struct {
	DefaultBucket string
}

type AWSConfig struct {
	Endpoint    string
	Region      string
	AWSS3Config AWSS3Config
}

func NewAWSConfig(v *viper.Viper) AWSConfig {
	v.SetDefault("AWS_ENDPOINT", "server://localhost:4566")
	v.SetDefault("AWS_REGION", "us-west-2")
	v.SetDefault("AWS_S3_DEFAULT_BUCKET", "default-bucket")

	return AWSConfig{
		Endpoint:    v.GetString("AWS_ENDPOINT"),
		Region:      v.GetString("AWS_REGION"),
		AWSS3Config: AWSS3Config{DefaultBucket: v.GetString("AWS_S3_DEFAULT_BUCKET")},
	}
}
