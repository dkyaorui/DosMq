package mq

import (
    Mongodb "DosMq/db/mongo"
    myRedis "DosMq/db/redis"
    MongoModule "DosMq/modules/mongo"
    "github.com/pkg/errors"
    log "github.com/sirupsen/logrus"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "net/http"
    "strings"
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
    var wg = &sync.WaitGroup{}
    // start message in que send to subscribe
    for _, item := range topicArray {
        wg.Add(1)
        val := item.(primitive.ObjectID)
        findResult, err := mongoUtils.FindOne(MongoModule.DB_TOPIC, bson.M{"_id": val})
        if err != nil {
            log.Errorf("err:%+v", errors.WithMessage(err, "queue init wrong"))
            continue
        }
        var topic MongoModule.Topic
        if err = findResult.Decode(&topic); err != nil {
            log.Errorf("err:%+v", errors.WithMessage(err, "queue init wrong"))
            continue
        }
        if strings.Compare(strings.ToLower(topic.ProcessMessageType), "push") == 0 {
            go pushMessageToSubscriber(val.Hex())
        }
    }
    // start message in redis send to que
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
    // init
    msgQue := MessageQueueMap[topicId]
    mongoUtils := Mongodb.Utils
    mongoUtils.OpenConn()
    mongoUtils.SetDB(mongoUtils.DBName)
    topicIdObj, err := primitive.ObjectIDFromHex(topicId)
    if err != nil {
        log.Errorf("err:%+v", errors.WithMessage(err, "topic id is wrong"))
    }
    findResult, err := mongoUtils.FindMore(MongoModule.DB_SUBSCRIBER, bson.M{"topic_id": topicIdObj})
    if err != nil {
        log.Errorf("err:%+v", errors.WithMessage(err, "[find err] find the topic wrong"))
        return
    }
    resultLength := len(findResult)
    subscribers := make([]MongoModule.Subscriber, resultLength)
    for index, item := range findResult {
        var subscriber MongoModule.Subscriber
        itemByte, _ := bson.Marshal(item)
        err = bson.Unmarshal(itemByte, &subscriber)
        if err != nil {
            log.Errorf("err:%+v", errors.WithMessage(err, "the item is not subscriber"))
            return
        }
        subscribers[index] = subscriber
    }
    mongoUtils.CloseConn()
    // main
    for {
        item, msgErr := msgQue.Pop()
        if msgErr != nil {
            if strings.Contains(msgErr.Error(), "lockFreeQueue is empty") {
                continue
            }
            log.Error("queue's item is not message")
            continue
        }
        message, ok := item.(MongoModule.Message)
        if ok {
            var wg = &sync.WaitGroup{}
            for _, subscriber := range subscribers {
                go func() {
                    defer wg.Done()
                    var req *http.Request
                    var err error
                    switch strings.ToUpper(subscriber.Method) {
                    case http.MethodGet:
                        req, err = methodGetRequest(message, subscriber)
                    case http.MethodPost:
                        req, err = methodPostRequest(message, subscriber)
                    }
                    if err != nil {
                        log.Errorf("err:%+v", errors.WithMessage(err, "new request fail"))
                        return
                    }
                    c := &http.Client{}
                    _, err = c.Do(req)
                }()
            }
            wg.Wait()
        }
    }
}

func methodGetRequest(message MongoModule.Message, subscriber MongoModule.Subscriber) (*http.Request, error) {
    req, err := http.NewRequest(http.MethodGet, subscriber.Host+subscriber.Api, nil)
    if err != nil {
        return nil, err
    }
    req.URL.Query().Add("message", message.Value)
    req.URL.Query().Add("message", message.Value)
    return req, err
}

func methodPostRequest(message MongoModule.Message, subscriber MongoModule.Subscriber) (*http.Request, error) {
    req, err := http.NewRequest(http.MethodPost, subscriber.Host+subscriber.Api, nil)
    if err != nil {
        return nil, err
    }
    req.Form.Add("message", message.Value)
    req.Form.Add("auth_key", subscriber.Key)
    return req, err
}
