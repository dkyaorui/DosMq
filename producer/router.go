package producer

import "github.com/gin-gonic/gin"

func Routes(router *gin.Engine) {
    producerRouter := router.Group("/api/producer")
    {
        producerRouter.POST("send", SendHandler)
    }
}
