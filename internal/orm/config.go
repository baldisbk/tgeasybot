package orm

import "time"

type Config struct {
	Host     string `yaml:"host" env:"TGBOT_DB_HOST"`
	Port     int    `yaml:"port" env:"TGBOT_DB_PORT"`
	User     string `yaml:"user" env:"TGBOT_DB_USER"`
	Password string `yaml:"password" env:"TGBOT_DB_PASSWORD"`
	Database string `yaml:"database" env:"TGBOT_DB_DATABASE"`
	CertFile string `yaml:"certfile" env:"TGBOT_DB_CERTFILE"`
	TLS      bool   `yaml:"tls"`

	Deprecate time.Duration `yaml:"deprecate"`
}
