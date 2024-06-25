package config

import (
	"strings"

	"github.com/spf13/viper"
)

func Init() {
	viper.AddConfigPath("./")
	viper.AddConfigPath("./config")
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	if err := viper.ReadInConfig(); err != nil {
		panic("init config failed. " + err.Error())
	}
}

func InitForTest() {
	jsonConfig := `{
        "create_transaction_timeout": "3",
        "max_retries": 2,
        "try_timeout": 1
    }`
	viper.SetConfigType("json")
	if err := viper.ReadConfig(strings.NewReader(jsonConfig)); err != nil {
		panic("failed to init config")
	}
}