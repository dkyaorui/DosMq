package router

import (
    "DosMq/consumer"
    "DosMq/producer"
    "DosMq/topicCenter"
    "github.com/gin-gonic/gin"
)

func GetRouter() *gin.Engine {
    router := gin.Default()
    consumer.Routes(router)
    producer.Routes(router)
    topicCenter.Routes(router)
    return router
}
