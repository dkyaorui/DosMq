package mq

import (
    Mongodb "DosMq/db/mongo"
    MongoModule "DosMq/modules/mongo"
    "github.com/pkg/errors"
    log "github.com/sirupsen/logrus"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
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

}
