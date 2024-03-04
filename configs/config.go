package config

import (
	"os"

	"github.com/cockroachdb/errors"
	"github.com/spf13/viper"
)

var (
	Confs = Config{}
)

type Config struct {
	Port            int         `mapstructure:"port"`
	SSL             bool        `mapstructure:"ssl"`
	MaintenanceMode bool        `mapstructure:"maintenanceMode"`
	Debug           bool        `mapstructure:"debug"`
	LogPath         string      `mapstructure:"logPath"`
	ReleaseMode     string      `mapstructure:"releaseMode"`
	Database        DB          `mapstructure:"database"`
}

func (g *Config) SetMaintenance(mode bool) {
	Confs.MaintenanceMode = mode // TODO: maybe use viper.unmarshal(&Confs)
	viper.Set("maintenanceMode", mode)
}

// Load returns configs
func (g *Config) Load(path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return g.file(path)
	}

	return errors.Newf("file not exists")
}

// file func
func (g *Config) file(path string) error {

	// name of config file (without extension)
	// REQUIRED if the config file does not have the extension in the name
	// path to look for the config file in
	viper.SetConfigFile(path)
	// viper.SetDefault()
	if err := viper.ReadInConfig(); err != nil {
		return err

	}

	return viper.Unmarshal(&Confs)
}
