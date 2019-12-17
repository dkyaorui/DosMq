package topic

import (
    "DosMq/middleware"
    "github.com/gin-gonic/gin"
)

// Routes the routes of topic
func Routes(router *gin.Engine) {
    topicRouter := router.Group("/api/topic")
    topicRouter.Use(middleware.CheckAuthKeyMiddleware())
    {
        topicRouter.POST("/sign_up", RegisterTopicHandler)
        topicRouter.POST("/del", DelTopicHandler)
        topicRouter.POST("/subscribe", SubscribeNewsHandler)
        topicRouter.POST("/cancel_subscribe", CancelSubscribeHandler)
    }
}
