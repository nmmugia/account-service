package config

import (
	"account-service/src/utils"

	"github.com/spf13/viper"
)

var (
	IsProd     bool
	AppName    string
	AppHost    string
	AppPort    int
	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
	DBPort     int
)

func loadConfig() {
	configPaths := []string{
		"./",     // For app
		"../../", // For test folder
	}

	for _, path := range configPaths {
		viper.SetConfigFile(path + ".env")

		if err := viper.ReadInConfig(); err == nil {
			utils.Log.Infof("Config file loaded from %s", path)
			return
		}
	}

	utils.Log.Error("Failed to load any config file")
}

func init() {
	loadConfig()

	// server confi
	IsProd = viper.GetString("APP_ENV") == "prod"
	AppName = viper.GetString("APP_NAME")
	AppHost = viper.GetString("APP_HOST")
	AppPort = viper.GetInt("APP_PORT")

	// db config
	DBHost = viper.GetString("DB_HOST")
	DBUser = viper.GetString("DB_USER")
	DBPassword = viper.GetString("DB_PASSWORD")
	DBName = viper.GetString("DB_NAME")
	DBPort = viper.GetInt("DB_PORT")
}
