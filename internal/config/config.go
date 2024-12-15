package config

import (
	"flag"
	"os"

	"github.com/baldisbk/tgbot/pkg/envconfig"
	"github.com/baldisbk/tgbot/pkg/tgapi"

	"github.com/baldisbk/tgeasybot/internal/orm"
	"github.com/baldisbk/tgeasybot/internal/poller"
	"github.com/baldisbk/tgeasybot/internal/sm"
	"github.com/baldisbk/tgeasybot/internal/timer"

	"golang.org/x/xerrors"
	"gopkg.in/yaml.v3"
)

var (
	defaultPath = "/etc/tgbot/config.yaml"
	develPath   = "config.yaml"
)

var configPath = flag.String("config", "", "path to config")
var develMode = flag.Bool("devel", false, "development mode")

type ConfigFlags struct {
	Path  string `yaml:"-"`
	Devel bool   `yaml:"-"`
}

type Config struct {
	ConfigFlags

	ApiConfig    tgapi.Config  `yaml:"tgapi"`
	DBConfig     orm.Config    `yaml:"db"`
	TimerConfig  timer.Config  `yaml:"timer"`
	PollerConfig poller.Config `yaml:"poller"`
	WorkerConfig sm.Config     `yaml:"worker"`
}

func ParseCustomConfig(config interface{}) (*ConfigFlags, error) {
	flag.Parse()

	flags := ConfigFlags{
		Devel: *develMode,
		Path:  *configPath,
	}
	if flags.Path == "" {
		if flags.Devel {
			flags.Path = develPath
		} else {
			flags.Path = defaultPath
		}
	}
	contents, err := os.ReadFile(flags.Path)
	if err != nil {
		return nil, xerrors.Errorf("read config: %w", err)
	}
	if err := yaml.Unmarshal(contents, config); err != nil {
		return nil, xerrors.Errorf("parse config: %w", err)
	}
	return &flags, nil
}

func ParseConfig() (*Config, error) {
	var cfg Config
	flags, err := ParseCustomConfig(&cfg)
	if err != nil {
		return nil, xerrors.Errorf("parse: %w", err)
	}
	if err := envconfig.UnmarshalEnv(&cfg); err != nil {
		return nil, xerrors.Errorf("parse env: %w", err)
	}
	cfg.ConfigFlags = *flags
	return &cfg, nil
}
