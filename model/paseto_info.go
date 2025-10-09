package model

type PasetoConfig struct {
	PubKey   string `mapstructure:"pub_key" json:"pub_key" yaml:"pub_key"`
	Implicit string `mapstructure:"implicit" json:"implicit" yaml:"implicit"`
}
type PasetoSecretConfig struct {
	SecretKey string `mapstructure:"secret_key" json:"secret_key" yaml:"secret_key"`
	Implicit  string `mapstructure:"implicit" json:"implicit" yaml:"implicit"`
}
type MySQL struct {
	DSN string `yaml:"dsn"`
}

type Redis struct {
	Address  string `yaml:"address"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type Kitex struct {
	Service         string `yaml:"service"`
	Address         string `yaml:"address"`
	MetricsPort     string `yaml:"metrics_port"`
	EnablePprof     bool   `yaml:"enable_pprof"`
	EnableGzip      bool   `yaml:"enable_gzip"`
	EnableAccessLog bool   `yaml:"enable_access_log"`
	LogLevel        string `yaml:"log_level"`
	LogFileName     string `yaml:"log_file_name"`
	LogMaxSize      int    `yaml:"log_max_size"`
	LogMaxBackups   int    `yaml:"log_max_backups"`
	LogMaxAge       int    `yaml:"log_max_age"`
}

type OTel struct {
	Endpoint string `yaml:"endpoint"`
	Insecure bool   `yaml:"insecure"`
}
