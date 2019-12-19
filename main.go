/*
The data type of all requests is json, and you need set the headers with "application/json"

The data type of all response is json.
eg:
{
    "Code": 200,
    "Data": {
        "Code": 200,
        "Data": "msg"
    }
}
*/

package main

import (
	"DosMq/consumer"
	"DosMq/db/mongo"
	"DosMq/db/redis"
	"DosMq/mq"
	"DosMq/producer"
	"DosMq/topic"
	"DosMq/utils"
	"flag"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	log "github.com/sirupsen/logrus"
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
	r := GetRouter()
	log.Info("router loaded……")

	// start listen message queue
	go mq.StartProcess()

	// run server
	if err := r.Run(":8080"); err != nil {
		log.Error(err)
	}
}

// GetRouter return all routes
func GetRouter() *gin.Engine {
	router := gin.Default()
	consumer.Routes(router)
	producer.Routes(router)
	topic.Routes(router)
	return router
}
