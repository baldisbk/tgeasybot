package sm

import "time"

type Config struct {
	Period  time.Duration `yaml:"period"`
	Workers int           `yaml:"workers"`
}
