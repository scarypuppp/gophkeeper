package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v6"
)

const (
	defaultTokenExpiresSeconds = 24 * 60 * 60 * 30 // 30 days
	defaultSecretKey           = "default-secret-key"
)

type Config struct {
	SecretKey       string `env:"SECRET_KEY"`
	Address         string `env:"RUN_ADDRESS"`
	DatabaseURI     string `env:"DATABASE_URI"`
	TokenExpSeconds int64  `env:"TOKEN_EXP"`
	S3Endpoint      string `env:"S3_ENDPOINT"`
	S3AccessKey     string `env:"S3_ACCESS_KEY"`
	S3SecretKey     string `env:"S3_SECRET_KEY"`
	S3Bucket        string `env:"S3_BUCKET"`
}

func GetConfig() (*Config, error) {
	return parseConfig(os.Args[1:])
}

func parseConfig(args []string) (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	fs := flag.NewFlagSet("gophkeeper", flag.ContinueOnError)
	secretKeyFlag := fs.String("k", "", "application secret key")
	addrFlag := fs.String("a", "", "server address host:port")
	databaseURIFlag := fs.String("d", "", "database dsn string")
	tokenExpFlag := fs.Int64("e", 0, "jwt token expires in seconds")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if cfg.SecretKey == "" {
		if *secretKeyFlag != "" {
			cfg.SecretKey = *secretKeyFlag
		} else {
			cfg.SecretKey = defaultSecretKey
		}
	}
	if cfg.Address == "" {
		if *addrFlag == "" {
			return nil, fmt.Errorf("service address should not be empty")
		}
		cfg.Address = *addrFlag
	}
	if cfg.DatabaseURI == "" {
		if *databaseURIFlag == "" {
			return nil, fmt.Errorf("database uri should not be empty")
		}
		cfg.DatabaseURI = *databaseURIFlag
	}
	if cfg.TokenExpSeconds == 0 {
		if *tokenExpFlag != 0 {
			cfg.TokenExpSeconds = *tokenExpFlag
		} else {
			cfg.TokenExpSeconds = defaultTokenExpiresSeconds
		}
	}

	return &cfg, nil
}
