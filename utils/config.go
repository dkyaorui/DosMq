package utils

import (
    "github.com/spf13/viper"
    "os"
)

func ConfigInit(configName *string, configPath *string) {
    viper.SetConfigName(*configName)
    viper.AddConfigPath(*configPath)
    redisPassword := os.Getenv("REDIS_PASSWORD")
    viper.Set("redisPassword", redisPassword)
    if err:=viper.ReadInConfig(); err != nil {
        panic(err)
    }
}