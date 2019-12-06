package mq

import (
    Mongodb "DosMq/db/mongo"
    myRedis "DosMq/db/redis"
    MongoModule "DosMq/modules/mongo"
    "github.com/pkg/errors"
    log "github.com/sirupsen/logrus"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "sync"
    "time"
)

var MessageQueueMap map[string]*LFQueue

/*
count the num of topic in mongodb.And creat the lfQueue for all of them.
So this function must run after db.Init().
*/
func Init() {
    mongoUtils := Mongodb.Utils
    mongoUtils.OpenConn()
    mongoUtils.SetDB(mongoUtils.DBName)
    defer mongoUtils.CloseConn()

    result, err := mongoUtils.Distinct(MongoModule.DB_TOPIC, "_id", bson.M{})
    if err != nil {
        log.Errorf("[distinct err]:%+v", errors.WithMessage(err, "db error"))
        panic("db init error")
    }
    topicArray := result.([]interface{})
    MessageQueueMap = make(map[string]*LFQueue)
    for _, item := range topicArray {
        val := item.(primitive.ObjectID)
        MessageQueueMap[val.Hex()] = NewQue(1024)
    }
    log.Info("mq init success")
}

/*
   start n goroutine to get message in redis and then push into queue. N is topic's number.
   start n goroutine to get message in queue and then push to subscriber. N is topic's number.
*/
func Start() {
    mongoUtils := Mongodb.Utils
    mongoUtils.OpenConn()
    mongoUtils.SetDB(mongoUtils.DBName)
    defer mongoUtils.CloseConn()

    result, err := mongoUtils.Distinct(MongoModule.DB_TOPIC, "_id", bson.M{})
    if err != nil {
        log.Errorf("[distinct err]:%+v", errors.WithMessage(err, "db error"))
        panic("mq init error")
    }
    topicArray := result.([]interface{})
    // start message in que send to subscribe


    // start message in redis send to que
    var wg sync.WaitGroup
    for _, item := range topicArray {
        wg.Add(1)
        val := item.(primitive.ObjectID)
        go redisToQue(val.Hex())
    }
    log.Info("mq init success")
    wg.Wait()
}

func redisToQue(topicId string) {
    redisClient := &myRedis.RDbClient
    redisKey := MongoModule.GetMqKeyByTopicId(topicId)
    msgQue := MessageQueueMap[topicId]
    for {
        redisMqLen, err := redisClient.LLen(redisKey)
        if err != nil {
            log.Errorf("[find error]:%+v", errors.WithMessage(err, "redis error"))
        }
        if redisMqLen == 0 {
            time.Sleep(1000)
        } else {
            for i := int64(0); i < redisMqLen; i++ {
                reply, err := redisClient.RPop(redisKey)
                if err != nil {
                    log.Errorf("[find error]:%+v", errors.WithMessage(err, "redis error"))
                } else {
                    message := reply.(MongoModule.Message)
                    for {
                        mqErr := msgQue.Push(message)
                        if mqErr != nil {
                            log.Errorf("[push error]:%+v", errors.WithMessage(err, "redis error"))
                            continue
                        } else {
                            break
                        }
                    }
                }
            }
        }
    }
}

func pushMessageToSubscriber(topicId string) {

}
