package consumer

import (
    "github.com/gin-gonic/gin"
    log "github.com/sirupsen/logrus"
    "net/http"
    "DosMq/db"
)

func Hello(context *gin.Context) {
    name := context.Query("name")
    coon := db.RedisClient.Get()
    defer coon.Close()
    if _, err :=coon.Do("SET", "name", name); err != nil {
        log.Error(err)
    }

    log.WithFields(log.Fields{
        "name": name,
    }).Info("hello name")
    context.JSON(http.StatusOK, gin.H{"name": name})
}
