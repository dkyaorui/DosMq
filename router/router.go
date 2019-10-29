package router

import (
    "DosMq/consumer"
    "DosMq/producer"
    "github.com/gin-gonic/gin"
)

func GetRouter() *gin.Engine {
    router := gin.Default()
    consumer.Routes(router)
    producer.Routes(router)
    return router
}
