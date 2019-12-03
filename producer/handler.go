package producer

import (
    Mongodb "DosMq/db/mongo"
    myRedis "DosMq/db/redis"
    MongoModule "DosMq/modules/mongo"
    "DosMq/utils"
    "github.com/gin-gonic/gin"
    "github.com/gin-gonic/gin/binding"
    "github.com/pkg/errors"
    log "github.com/sirupsen/logrus"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "net/http"
    "time"
)

func SendHandler(c *gin.Context) {
    var owner MongoModule.Owner
    var message MongoModule.Message
    var err error
    if err = c.ShouldBindBodyWith(&owner, binding.JSON); err != nil {
        log.Infof("binding data error,struct:%s", "owner")
        c.AbortWithStatusJSON(http.StatusPartialContent, utils.RequestResult{
            Code: http.StatusPartialContent,
            Data: err.Error(),
        })
        return
    }
    if err = c.ShouldBindBodyWith(&message, binding.JSON); err != nil {
        log.Infof("binding data error,struct:%s", "message")
        c.AbortWithStatusJSON(http.StatusPartialContent, utils.RequestResult{
            Code: http.StatusPartialContent,
            Data: err.Error(),
        })
        return
    }

    mongoUtils := Mongodb.Utils
    mongoUtils.OpenConn()
    mongoUtils.SetDB(mongoUtils.DBName)
    defer mongoUtils.CloseConn()

    owner.HashCode = owner.GetHashCode()
    findResult, err := mongoUtils.FindOne(MongoModule.DB_OWNER, bson.M{"hash_code": owner.HashCode})
    if err != nil {
        log.Errorf("[find error] owner:%v", owner)
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }
    if err = findResult.Decode(&owner); err != nil {
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }

    message.Id = primitive.NewObjectID()
    message.TopicId = owner.TopicID
    message.Timestamp = time.Now().UnixNano()
    message.HashCode = message.GetHashCode()

    redisClient := &myRedis.RDbClient
    isExists, err := redisClient.Exists(message.GetRedisKey())
    if err != nil {
        log.Errorf("[find error]:%+v", errors.WithMessage(err, "redis error"))
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }
    if isExists {
        c.AbortWithStatusJSON(http.StatusAlreadyReported, utils.RequestResult{
            Code: http.StatusAlreadyReported,
            Data: "the message had been send",
        })
        return
    }
    err = redisClient.Set(message.GetRedisKey(), message, 0)
    if err != nil {
        log.Errorf("[set error]:%+v", errors.WithMessage(err, "redis error"))
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, utils.RequestResult{
        Code: http.StatusOK,
        Data: "the message has send",
    })
}
