/*
The data type of all requests is json, and you need set the headers with "application/json"
*/

package main

import (
    "DosMq/db/mongo"
    "DosMq/db/redis"
    "DosMq/router"
    "DosMq/utils"
    "flag"
    "fmt"
    log "github.com/sirupsen/logrus"
    "time"
)

func main() {
    // time init
    loc, _ := time.LoadLocation("Asia/Chongqing")
    fmt.Printf("时区: %s\n", loc)

    // config init
    configName := flag.String("configName", "config", "config file's name.")
    configPath := flag.String("configPath", "./config", "config file's path.")
    utils.ConfigInit(configName, configPath)

    // log init
    utils.LogInit()

    // db init
    redis.Init()
    mongo.Init()

    // load router
    r := router.GetRouter()
    log.Info("router loaded……")

    // run server
    if err := r.Run(":8080"); err != nil {
        log.Error(err)
    }
}
