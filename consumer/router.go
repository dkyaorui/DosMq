package consumer

import "github.com/gin-gonic/gin"

func Routes(router *gin.Engine) {
	consumerRouter := router.Group("/api/consumer")
	{
	    consumerRouter.POST("/get", GetHandler)
	}
}
