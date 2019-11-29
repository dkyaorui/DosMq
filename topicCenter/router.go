package topicCenter

import (
    "DosMq/middleware"
    "github.com/gin-gonic/gin"
)

func Routes(router *gin.Engine) {
    topicRouter := router.Group("/topic")
    topicRouter.Use(middleware.CheckAuthKeyMiddleware())
    {
        topicRouter.POST("/sign_up", TopicRegister)
        topicRouter.POST("/del", TopicDel)
        topicRouter.POST("/subscribe", SubscribeNews)
    }
}
