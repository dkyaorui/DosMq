package consumer

import (
    "github.com/gin-gonic/gin"
    log "github.com/sirupsen/logrus"
    "net/http"
)

func Hello(context *gin.Context) {
    name := context.Query("name")
    log.WithFields(log.Fields{
        "name": name,
    }).Info("hello name")
    context.JSON(http.StatusOK, gin.H{"name": name})
}
