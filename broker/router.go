package broker

import "github.com/gin-gonic/gin"

func Routes(router *gin.Engine) {
    brokerRouter := router.Group("/broker")
    {
        brokerRouter.POST("/sign_up_topic", TopicRegister)
    }
}