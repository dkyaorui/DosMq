package conf

import (
    "fmt"
    "github.com/spf13/viper"
)

func Init(configName *string, configPath *string) {
    viper.SetConfigName(*configName)
    viper.AddConfigPath(*configPath)
    if err:=viper.ReadInConfig(); err != nil {
        fmt.Println(err)
    }
}