package topicCenter

import (
    "DosMq/middleware"
    "github.com/gin-gonic/gin"
)

func Routes(router *gin.Engine) {
    topicRouter := router.Group("/topic")
    topicRouter.Use(middleware.CheckAuthKeyMiddleware())
    {
        topicRouter.POST("/sign_up", TopicRegisterHandler)
        topicRouter.POST("/del", TopicDelHandler)
        topicRouter.POST("/subscribe", SubscribeNewsHandler)
        topicRouter.POST("/cancel_subscribe", CancelSubscribeHandler)
    }
}
