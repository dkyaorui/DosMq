package consumer

import (
    Mongodb "DosMq/db/mongo"
    mongoModule "DosMq/modules/mongo"
    "DosMq/mq"
    "DosMq/utils"
    "github.com/gin-gonic/gin"
    "github.com/gin-gonic/gin/binding"
    "github.com/pkg/errors"
    log "github.com/sirupsen/logrus"
    "go.mongodb.org/mongo-driver/bson"
    "net/http"
    "strings"
)

/*
 query the topic_id by subscriber,and get the message in topic's queue
*/

func GetHandler(c *gin.Context) {
    var subscriber mongoModule.Subscriber
    var err error

    if err = c.ShouldBindBodyWith(&subscriber, binding.JSON); err != nil {
        log.Info("binding data error, struct: subscriber")
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

    subscriber.HashCode = subscriber.GetHashCode()
    findResult, err := mongoUtils.FindOne(mongoModule.DB_SUBSCRIBER, bson.M{"hash_code": subscriber.HashCode})
    if err != nil {
        log.Errorf("[find error] subscriber:%v, err:%+v", subscriber,
            errors.WithMessage(err, "find error"))
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }
    if err = findResult.Decode(&subscriber); err != nil {
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }
    log.Infof("subscriber's topic_id:%s", subscriber.TopicId.Hex())

    msgQue := mq.MessageQueueMap[subscriber.TopicId.Hex()]
    queItem, msgErr := msgQue.Pop()
    if msgErr != nil {
        if strings.Contains(msgErr.Error(), "lockFreeQueue is empty") {
            c.AbortWithStatusJSON(http.StatusOK, utils.RequestResult{
                Code: http.StatusOK,
                Data: msgErr.Error(),
            })
            return
        }
    }
    message, ok := queItem.(mongoModule.Message)
    if ok {
        // return message to subscriber
        log.Info("[pull message] message:%+v", message)
        c.JSON(http.StatusOK, utils.RequestResult{
            Code: http.StatusOK,
            Data: message.Value,
        })
    } else {
        log.Errorf("[decode error] message:%v, err:%s", subscriber, "queItem is not message")
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: "Message que is wrong",
        })
        return
    }
}
