package model

type PasetoConfig struct {
	PubKey   string `mapstructure:"pub_key" json:"pub_key" yaml:"pub_key"`
	Implicit string `mapstructure:"implicit" json:"implicit" yaml:"implicit"`
}
type PasetoSecretConfig struct {
	SecretKey string `mapstructure:"secret_key" json:"secret_key" yaml:"secret_key"`
	Implicit  string `mapstructure:"implicit" json:"implicit" yaml:"implicit"`
}
