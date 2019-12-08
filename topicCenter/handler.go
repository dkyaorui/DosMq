package topicCenter

import (
    Mongodb "DosMq/db/mongo"
    "DosMq/modules"
    "DosMq/mq"
    "DosMq/utils"
    "github.com/gin-gonic/gin"
    "github.com/gin-gonic/gin/binding"
    log "github.com/sirupsen/logrus"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "net/http"
)

func TopicRegisterHandler(c *gin.Context) {
    var owner modules.Owner
    var topic modules.Topic
    var err error
    // check data
    if err = c.ShouldBindBodyWith(&owner, binding.JSON); err != nil {
        log.Infof("binding data error,struct:%s", "owner")
        c.AbortWithStatusJSON(http.StatusPartialContent, utils.RequestResult{
            Code: http.StatusPartialContent,
            Data: err.Error(),
        })
        return
    }
    if err = c.ShouldBindBodyWith(&topic, binding.JSON); err != nil {
        log.Infof("binding data error,struct:%s", "topic")
        c.AbortWithStatusJSON(http.StatusPartialContent, utils.RequestResult{
            Code: http.StatusPartialContent,
            Data: err.Error(),
        })
        return
    }

    // open mongo client
    mongoUtils := Mongodb.Utils
    mongoUtils.OpenConn()
    mongoUtils.SetDB(mongoUtils.DBName)
    defer mongoUtils.CloseConn()

    // curd
    topic.HashCode = topic.GetHashCode()
    topic.Id = primitive.NewObjectID()
    owner.TopicID = topic.Id
    owner.Id = primitive.NewObjectID()
    owner.HashCode = owner.GetHashCode()
    topic.Owner = owner
    // check is data repeat
    var isRepeat bool
    if isRepeat, err = topic.CheckIsRepeat(&mongoUtils); err != nil {
        log.Error(err)
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }
    if isRepeat {
        c.AbortWithStatusJSON(http.StatusCreated, utils.RequestResult{
            Code: http.StatusCreated,
            Data: "topic is repeat",
        })
        return
    }
    byteTopic, _ := bson.Marshal(&topic)
    doc := bson.M{}
    _ = bson.Unmarshal(byteTopic, &doc)
    _, err = mongoUtils.InsertOne(modules.DB_TOPIC, doc)
    if err != nil {
        log.Infof("insert data error. data:%v", topic)
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }
    log.Infof("[insert]topic:%v", topic)

    if isRepeat, err = owner.CheckIsRepeat(&mongoUtils); err != nil {
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }
    if isRepeat {
        c.AbortWithStatusJSON(http.StatusCreated, utils.RequestResult{
            Code: http.StatusCreated,
            Data: "owner is repeat",
        })
        return
    }
    byteOwner, _ := bson.Marshal(&owner)
    doc = bson.M{}
    _ = bson.Unmarshal(byteOwner, &doc)
    _, err = mongoUtils.InsertOne(modules.DB_OWNER, doc)
    if err != nil {
        log.Infof("insert data error. data:%v", topic)
        _, _ = mongoUtils.DelOne("topic", bson.M{"_id": owner.Id})
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }
    log.Infof("[insert]owner:%v", owner)
    queChannel := &mq.NewQueueChannel
    *queChannel <- topic.Id.Hex()
    c.JSON(http.StatusOK, utils.RequestResult{
        Code: http.StatusOK,
        Data: "register topic success",
    })
}

/*
   receive the owner info(key, server_name, host), then find the topic_id.
   Before delete the topic, we need del the subscribers first.
*/
func TopicDelHandler(c *gin.Context) {
    var owner modules.Owner
    var err error
    if err = c.ShouldBindBodyWith(&owner, binding.JSON); err != nil {
        log.Infof("binding data error,struct:%s", "owner")
        c.AbortWithStatusJSON(http.StatusPartialContent, utils.RequestResult{
            Code: http.StatusPartialContent,
            Data: err.Error(),
        })
    }
    owner.HashCode = owner.GetHashCode()
    // open mongo client
    mongoUtils := Mongodb.Utils
    mongoUtils.OpenConn()
    mongoUtils.SetDB(mongoUtils.DBName)
    defer mongoUtils.CloseConn()

    // main
    spc := bson.M{"hash_code": owner.HashCode}
    findResult, err := mongoUtils.FindOne(modules.DB_OWNER, spc)
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
    // del subscriber
    delResult, err := mongoUtils.DelMany(modules.DB_SUBSCRIBER, bson.M{"topic_id": owner.TopicID})
    if err != nil {
        log.Errorf("[delete error] subscriber's topic id: %v", owner.TopicID)
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }
    log.Infof("[delete] delete %d rows", delResult.DeletedCount)
    // del topic
    _, err = mongoUtils.DelOne(modules.DB_TOPIC, bson.M{"_id": owner.TopicID})
    if err != nil {
        log.Errorf("[delete error] topic id: %v", owner.TopicID)
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }
    log.Infof("[delete] topic id: %v", owner.TopicID)
    // del owner
    _, err = mongoUtils.DelOne(modules.DB_OWNER, bson.M{"_id": owner.Id})
    if err != nil {
        log.Errorf("[delete error] owner id: %v", owner.Id)
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }
    log.Infof("[delete] owner id: %v", owner.Id)
    mq.DelTopicChannel <- owner.TopicID.Hex()
    c.JSON(http.StatusOK, utils.RequestResult{
        Code: http.StatusOK,
        Data: "del success",
    })
}

func SubscribeNewsHandler(c *gin.Context) {
    var owner modules.Owner
    var subscriber modules.Subscriber
    var topic modules.Topic
    var err error

    if err = c.ShouldBindBodyWith(&owner, binding.JSON); err != nil {
        log.Infof("binding data error,struct:%s", "owner")
        c.AbortWithStatusJSON(http.StatusPartialContent, utils.RequestResult{
            Code: http.StatusPartialContent,
            Data: err.Error(),
        })
        return
    }
    if err = c.ShouldBindBodyWith(&subscriber, binding.JSON); err != nil {
        log.Infof("binding data error,struct:%s", "subscriber")
        c.AbortWithStatusJSON(http.StatusPartialContent, utils.RequestResult{
            Code: http.StatusPartialContent,
            Data: err.Error(),
        })
        return
    }
    // open mongo client
    mongoUtils := Mongodb.Utils
    mongoUtils.OpenConn()
    mongoUtils.SetDB(mongoUtils.DBName)
    defer mongoUtils.CloseConn()
    // main
    owner.HashCode = owner.GetHashCode()
    spc := bson.M{"hash_code": owner.HashCode}
    findResult, err := mongoUtils.FindOne(modules.DB_OWNER, spc)
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
    log.Infof("topic id:%v", owner.TopicID)

    subscriber.TopicId = owner.TopicID
    subscriber.HashCode = subscriber.GetHashCode()

    var isRepeat bool
    if isRepeat, err = subscriber.CheckIsRepeat(&mongoUtils); err != nil {
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }
    if isRepeat {
        c.AbortWithStatusJSON(http.StatusCreated, utils.RequestResult{
            Code: http.StatusCreated,
            Data: "subscriber is repeat",
        })
        return
    }

    findResult, err = mongoUtils.FindOne(modules.DB_TOPIC, bson.M{"_id": owner.TopicID})
    if err != nil {
        log.Errorf("[find error] owner:%v", owner)
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }
    if err = findResult.Decode(&topic); err != nil {
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }
    // update topic's subscribers
    if topic.Subscribers == nil {
        topic.Subscribers = []modules.Subscriber{subscriber}
    } else {
        topic.Subscribers = append(topic.Subscribers, subscriber)
    }
    _, err = mongoUtils.UpdateOne(
        modules.DB_TOPIC,
        bson.M{"_id": topic.Id},
        bson.M{
            "$set": bson.M{
                "subscribers": topic.Subscribers,
            },
        })
    if err != nil {
        log.Errorf("[update error] topic's _id:%s", topic.Id)
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }
    doc := bson.M{}
    subscriber.Id = primitive.NewObjectID()
    byteSubscriber, _ := bson.Marshal(subscriber)
    _ = bson.Unmarshal(byteSubscriber, &doc)
    _, err = mongoUtils.InsertOne(modules.DB_SUBSCRIBER, doc)
    if err != nil {
        log.Infof("[insert error]data: %v", subscriber)
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }
    c.JSON(http.StatusOK, utils.RequestResult{
        Code: http.StatusOK,
        Data: "subscribe success",
    })
}

func CancelSubscribeHandler(c *gin.Context) {
    var subscribe modules.Subscriber
    var err error

    if err = c.ShouldBindBodyWith(&subscribe, binding.JSON); err != nil {
        log.Info("binding data error, struct:%s", "subscriber")
        c.AbortWithStatusJSON(http.StatusPartialContent, utils.RequestResult{
            Code: http.StatusPartialContent,
            Data: err.Error(),
        })
        return
    }
    subscribe.HashCode = subscribe.GetHashCode()

    mongoUtils := Mongodb.Utils
    mongoUtils.OpenConn()
    mongoUtils.SetDB(mongoUtils.DBName)
    defer mongoUtils.CloseConn()

    spc := bson.M{"hash_code": subscribe.HashCode}
    findResult, err := mongoUtils.FindOne(modules.DB_SUBSCRIBER, spc)
    if err != nil {
        log.Errorf("[find error] subscriber:%v", subscribe)
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }
    if err = findResult.Decode(&subscribe); err != nil {
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }
    log.Infof("[find] subscribe:%v", subscribe)

    var topic modules.Topic
    spc = bson.M{"_id": subscribe.TopicId}
    findResult, err = mongoUtils.FindOne(modules.DB_TOPIC, spc)
    if err != nil {
        log.Errorf("[find error] topic:%+v", err)
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }
    if err = findResult.Decode(&topic); err != nil {
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }
    log.Infof("[find] topic:%v", topic)

    topic.DelSubscribe(subscribe)
    _, err = mongoUtils.UpdateOne(
        modules.DB_TOPIC,
        bson.M{"_id": topic.Id},
        bson.M{
            "$set": bson.M{
                "subscribers": topic.Subscribers,
            },
        })
    if err != nil {
        log.Errorf("[update error] topic's _id:%s", topic.Id)
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }

    spc = bson.M{"_id": subscribe.Id}
    _, err = mongoUtils.DelOne(modules.DB_SUBSCRIBER, spc)
    if err != nil {
        log.Errorf("[delete error] subscriber id: %v", subscribe.Id)
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }
    log.Infof("[delete] subscriber id: %v", subscribe.Id)

    c.JSON(http.StatusOK, utils.RequestResult{
        Code: http.StatusOK,
        Data: "delete subscriber success",
    })

}
