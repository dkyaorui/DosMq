package producer

import "github.com/gin-gonic/gin"

func Routes(router *gin.Engine) {
    producerRouter := router.Group("/producer")
    {
        producerRouter.POST("send", SendHandler)
    }
}
