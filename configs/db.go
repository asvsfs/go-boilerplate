package config

type DB struct {
	User     string `mapstructure:"user"`
	Host     string `mapstructure:"host"`
	Port     uint   `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Password string `mapstructure:"password"`
}

type MongoSQLDB struct {
	User     string `mapstructure:"user"`
	Host     string `mapstructure:"host"`
	Port     uint   `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Password string `mapstructure:"password"`
}
