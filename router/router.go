package router

import (
    "DosMq/consumer"
    "DosMq/producer"
    "DosMq/topic"
    "github.com/gin-gonic/gin"
)

// GetRouter return all routes
func GetRouter() *gin.Engine {
    router := gin.Default()
    consumer.Routes(router)
    producer.Routes(router)
    topic.Routes(router)
    return router
}
