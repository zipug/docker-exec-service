package config

import (
	"errors"
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Redis struct {
	Host          string `toml:"host" env:"REDIS_HOST" env-default:"localhost"`
	Port          int    `toml:"port" env:"REDIS_PORT" env-default:"6379"`
	DB            int    `toml:"db" env:"REDIS_DB" env-default:"0"`
	User          string `toml:"user" env:"REDIS_USER"`
	Password      string `toml:"password" env:"REDIS_USER_PASSWORD"`
	RedisPassword string `toml:"redis_password" env:"REDIS_PASSWORD"`
}

type Postgres struct {
	Host           string `toml:"host" env:"POSTGRES_HOST" env-default:"localhost"`
	Port           int    `toml:"port" env:"POSTGRES_PORT" env-default:"5432"`
	User           string `toml:"user" env:"POSTGRES_USER"`
	Password       string `toml:"password" env:"POSTGRES_PASSWORD"`
	DBName         string `toml:"db_name" env:"POSTGRES_DB_NAME"`
	SSLMode        string `toml:"ssl_mode" env:"POSTGRES_SSL_MODE" env-default:"disable"`
	MigrationsPath string `toml:"migrations_path" env:"POSTGRES_MIGRATIONS_PATH" env-required:"true"`
}

type MiniO struct {
	Host              string        `toml:"host" env:"MINIO_HOST"`
	Port              int           `toml:"port" env:"MINIO_PORT"`
	User              string        `toml:"user" env:"MINIO_ROOT_USER"`
	Password          string        `toml:"password" env:"MINIO_ROOT_PASSWORD"`
	BucketArticles    string        `toml:"articles_bucket" env:"MINIO_ARTICLES_BUCKET"`
	BucketAttachments string        `toml:"attachments_bucket" env:"MINIO_ATTACHMENTS_BUCKET"`
	BucketAvatars     string        `toml:"avatars_bucket" env:"MINIO_AVATARS_BUCKET"`
	UrlLifetime       time.Duration `toml:"url_lifetime" env:"MINIO_URL_LIFETIME" env-default:"4h"`
	UseSsl            bool          `toml:"use_ssl" env:"MINIO_USE_SSL"`
}

type Docker struct {
	ImageName string `toml:"image_name" env:"TELEGRAM_IMAGE_NAME"`
	Timeout   int    `toml:"timeout" env:"TELEGRAM_TIMEOUT" env-default:"10"`
}

type ExecutorConfig struct {
	ServerURL  string   `toml:"server" env:"SERVER_URL"`
	Redis      Redis    `toml:"redis"`
	Postgres   Postgres `toml:"postgres"`
	MiniO      MiniO    `toml:"minio"`
	Docker     Docker   `toml:"docker"`
	configPath string
}

func NewConfigService() *ExecutorConfig {
	var path string

	flag.StringVar(&path, "config", "", "path to config file")
	flag.Parse()

	if path == "" {
		if cfg_path, ok := os.LookupEnv("CONFIG_PATH"); ok {
			path = cfg_path
		}
	}

	cfg := &ExecutorConfig{configPath: path}
	if err := cfg.load(); err != nil {
		panic(err)
	}
	return cfg
}

func (cfg *ExecutorConfig) load() error {
	if cfg.configPath == "" {
		return errors.New("config path is not set")
	}

	if err := cleanenv.ReadConfig(cfg.configPath, cfg); err != nil {
		return err
	}

	return nil
}
