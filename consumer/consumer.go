package consumer

import (
    "DosMq/db/mongo"
    "DosMq/db/redis"
    "context"
    "github.com/gin-gonic/gin"
    log "github.com/sirupsen/logrus"
    "net/http"
    "time"
)

type Ns struct {
    Name string "name"
}

func Hello(c *gin.Context) {
    name := c.Query("name")
    database := mongo.GetMongoDataBase()
    collection := database.Collection("test")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    result, err := collection.InsertOne(ctx, Ns{Name: name})
    if err != nil {
        log.Error("insert error")
    }
    log.Infof("insert %s", result.InsertedID)
    coon := redis.RedisClient.Get()
    defer coon.Close()
    if _, err := redis.Delete(coon, "name", name); err != nil {
        log.Error(err)
    }

    log.WithFields(log.Fields{
        "name": name,
    }).Info("hello name")
    c.JSON(http.StatusOK, gin.H{"name": name})
}
