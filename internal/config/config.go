package config

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/parxyws/cozybox/internal/pkg/validator"
	"github.com/spf13/viper"
)

type Config struct {
	Server  ServerConfig        `mapstructure:"server" validate:"required"`
	WriteDB PostgresWriteConfig `mapstructure:"write_db" validate:"required"`
	ReadDB  PostgresReadConfig  `mapstructure:"read_db" validate:"required"`
	AWS     AwsConfig           `mapstructure:"aws" validate:"required"`
	Logger  LoggerConfig        `mapstructure:"logger" validate:"required"`
	Redis   RedisConfig         `mapstructure:"redis" validate:"required"`
	Mail    MailConfig          `mapstructure:"mail" validate:"required"`
	Admin   AdminConfig         `mapstructure:"admin" validate:"required"`
}

type ServerConfig struct {
	Host         string        `mapstructure:"host" validate:"required"`
	Port         int           `mapstructure:"port" validate:"required,min=1,max=65535"`
	BaseUrl      string        `mapstructure:"base_url" validate:"required,url"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" validate:"required"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" validate:"required"`
	SSL          bool          `mapstructure:"ssl"`
	Mode         string        `mapstructure:"mode" validate:"required,oneof=debug release test"`
	JWTSecretKey string        `mapstructure:"jwt_secret_key" validate:"required"`
}

type PostgresWriteConfig struct {
	User     string `mapstructure:"user" validate:"required"`
	Password string `mapstructure:"password" validate:"required"`
	Host     string `mapstructure:"host" validate:"required"`
	Port     int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	NameDB   string `mapstructure:"name_db" validate:"required"`
}

type PostgresReadConfig struct {
	User     string `mapstructure:"user" validate:"required"`
	Password string `mapstructure:"password" validate:"required"`
	Host     string `mapstructure:"host" validate:"required"`
	Port     int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	NameDB   string `mapstructure:"name_db" validate:"required"`
}

type LoggerConfig struct {
	Level       string `mapstructure:"level" validate:"required,oneof=trace debug info warn error fatal panic"`
	Caller      bool   `mapstructure:"caller"`
	Encoding    string `mapstructure:"encoding" validate:"omitempty,oneof=json text"`
	Development bool   `mapstructure:"development"`
}

type AwsConfig struct {
	Endpoint       string `mapstructure:"endpoint" validate:"required"`
	MiniEndpoint   string `mapstructure:"mini_endpoint" validate:"required"`
	MinioAccessKey string `mapstructure:"minio_access_key" validate:"required"`
	MinioSecretKey string `mapstructure:"minio_secret_key" validate:"required"`
	MinioLicense   string `mapstructure:"minio_license" validate:"required"`
	UseSSL         bool   `mapstructure:"use_ssl"`
	BucketName     string `mapstructure:"bucket_name"`
}

type RedisConfig struct {
	Host      string `mapstructure:"host" validate:"required"`
	Port      int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	Password  string `mapstructure:"password"`
	AuthDB    int    `mapstructure:"auth_db" validate:"min=0"`
	CacheDB   int    `mapstructure:"cache_db" validate:"min=0"`
	LimiterDB int    `mapstructure:"limiter_db" validate:"min=0"`
}

type MailConfig struct {
	Host     string `mapstructure:"host" validate:"required"`
	Port     int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	User     string `mapstructure:"user" validate:"required"`
	Password string `mapstructure:"password" validate:"required"`
}

type AdminConfig struct {
	User     string `mapstructure:"user" validate:"required"`
	Email    string `mapstructure:"email" validate:"required,email"`
	Password string `mapstructure:"password" validate:"required"`
}

func InitAppConfig() (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AutomaticEnv()

	path := os.Getenv("CONFIG_PATH")

	if path != "" {
		// If CONFIG_PATH is a file → use it directly
		if filepath.Ext(path) != "" {
			v.SetConfigFile(path)
		} else {
			// If CONFIG_PATH is a directory → search inside it
			v.AddConfigPath(path)
		}
	} else {
		// fallback: project root
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := errors.AsType[viper.ConfigFileNotFoundError](err); ok {
			return nil, err
		}

		return nil, err
	}

	config := new(Config)
	if err := v.Unmarshal(config); err != nil {
		return nil, err
	}

	// Set default for logger encoding if missing
	if config.Logger.Encoding == "" {
		config.Logger.Encoding = "json"
	}

	if err := validator.Validate.Struct(config); err != nil {
		return nil, err
	}

	return config, nil
}
