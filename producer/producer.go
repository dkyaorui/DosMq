package producer

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

func Hello(context *gin.Context) {
    name := context.Query("name")
    context.JSON(http.StatusOK, name)
}