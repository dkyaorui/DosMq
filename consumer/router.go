package consumer

import "github.com/gin-gonic/gin"

func Routes(router *gin.Engine) {
	consumerRouter := router.Group("/consumer")
	{
	    consumerRouter.POST("/get", GetHandler)
	}
}
