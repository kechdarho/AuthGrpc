package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"os"
	myecdsa "sso/internal/lib/encryption/ecdsa"
	"sync"
	"time"
)

type EnvType string

const (
	test EnvType = "test"
	prod EnvType = "prod"
	dev  EnvType = "dev"
)

type (
	Config struct {
		Environment EnvType    `yaml:"ENVIRONMENT"`
		HTTP        HTTP       `yaml:"server"`
		Sqlite      Sqlite     `yaml:"sqlite"`
		Cache       Cache      `yaml:"cache"`
		Logger      Logger     `yaml:"logger"`
		GRPC        GRPCConfig `yaml:"grpc"`
		JWT         JWT        `yaml:"jwt"`
	}

	HTTP struct {
		Host           string        `yaml:"HTTP_HOST"`
		Port           string        `yaml:"HTTP_PORT"`
		MaxHeaderBytes int           `yaml:"HTTP_MAX_HEADER_BYTES" default:"1"`
		ReadTimeout    time.Duration `yaml:"HTTP_READ_TIMEOUT" default:"10s"`
		WriteTimeout   time.Duration `yaml:"HTTP_WRITE_TIMEOUT" default:"10s"`
	}

	GRPCConfig struct {
		Port    int           `yaml:"GRPC_PORT"`
		Timeout time.Duration `yaml:"GRPC_TIMEOUT" default:"10h"`
	}

	Cache struct {
		DefaultExpiration time.Duration `yaml:"DEFAULT_EXPIRATION_CACHE" default:"4h"`
		CleanupInterval   time.Duration `yaml:"CLEANUP_INTERVAL_CACHE" default:"4h"`
	}

	Sqlite struct {
		Path string `yaml:"storage_path" env-required:"true"`
	}

	Logger struct {
		Level string `yaml:"LOGGER_LEVEL" default:"info"`
	}

	JWT struct {
		AccessTokenTTL  time.Duration `yaml:"accessTokenTTL"`
		RefreshTokenTTL time.Duration `yaml:"refreshTokenTTL"`
		PrivateKeyPath  string        `yaml:"privateKeyPath"`
		SecretKey       []byte
	}
)

func (c *Config) IsDev() bool {
	return c.Environment == dev
}

func (c *Config) IsTest() bool {
	return c.Environment == test
}

func (c *Config) IsProd() bool {
	return c.Environment == prod
}

var (
	instance Config
	once     sync.Once
)

func Get() *Config {
	once.Do(func() {
		configFile := "config.yaml"

		file, err := os.Open(configFile)
		if err != nil {
			log.Fatalf("Error opening config file: %v", err)
		}
		defer func(file *os.File) {
			_ = file.Close()
		}(file)

		data, err := io.ReadAll(file)
		if err != nil {
			log.Fatalf("Error reading config file: %v", err)
		}

		err = yaml.Unmarshal(data, &instance)
		if err != nil {
			log.Fatalf("Error unmarshaling YAML: %v", err)
		}

		switch instance.Environment {
		case test, prod, dev:
		default:
			log.Fatal("config environment should be test, prod or dev")
		}

		instance.JWT.SecretKey, err = myecdsa.GenerateECDSAKey(instance.JWT.PrivateKeyPath)
		if err != nil {
			log.Fatalf("Error initializing JWT secret key: %v", err)
		}

		if instance.IsDev() {
			configBytes, err := yaml.Marshal(&instance)
			if err != nil {
				log.Fatalf("Error marshaling config to YAML: %v", err)
			}
			fmt.Println("Configuration:n", string(configBytes))
		}

	})
	return &instance
}
