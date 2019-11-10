package main

import (
	"DosMq/db/mongo"
	"DosMq/db/redis"
	"DosMq/router"
	"DosMq/utils"
	"flag"
	log "github.com/sirupsen/logrus"
)

func main() {
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
	log.Error("router loaded……")

	// run server
	if err := r.Run(":8080"); err != nil {
		log.Error(err)
	}
}
