package main

import (
    "DosMq/conf"
    "DosMq/router"
    "DosMq/utils"
    "flag"
    log "github.com/sirupsen/logrus"
)

func main() {
    // config init
    configName := flag.String("configName", "config", "config file's name.")
    configPath := flag.String("configPath", "./config", "config file's path.")
    conf.Init(configName, configPath)

    // log init
    utils.LogInit()

    // load router
    r := router.GetRouter()
    log.Error("router loaded……")

    // run server
    if err := r.Run(":8080"); err != nil {
        log.Error(err)
    }
}
