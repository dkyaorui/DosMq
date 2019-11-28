package topicCenter

import (
    "github.com/gin-gonic/gin"
)

func Routes(router *gin.Engine) {
    topicRouter := router.Group("/topic")
    {
        topicRouter.POST("/sign_up", TopicRegister)
    }
}
