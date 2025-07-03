package jobqueue

import (
	"github.com/spf13/viper"
)

type Config struct {
	MaxRetries   int
	DelaySeconds int
}

func LoadConfig() Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AutomaticEnv()

	viper.SetDefault("max_retries", 3)
	viper.SetDefault("delay_seconds", 5)

	_ = viper.ReadInConfig()

	return Config{
		MaxRetries:   viper.GetInt("max_retries"),
		DelaySeconds: viper.GetInt("delay_seconds"),
	}
}
