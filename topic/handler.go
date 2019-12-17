package topic

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

func RegisterTopicHandler(c *gin.Context) {
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
    topic.ID = primitive.NewObjectID()
    owner.TopicID = topic.ID
    owner.ID = primitive.NewObjectID()
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
    _, err = mongoUtils.InsertOne(modules.DbTopic, doc)
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
    _, err = mongoUtils.InsertOne(modules.DbOwner, doc)
    if err != nil {
        log.Infof("insert data error. data:%v", topic)
        _, _ = mongoUtils.DelOne("topic", bson.M{"_id": owner.ID})
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }
    log.Infof("[insert]owner:%v", owner)
    queChannel := &mq.NewQueueChannel
    *queChannel <- topic.ID.Hex()
    c.JSON(http.StatusOK, utils.RequestResult{
        Code: http.StatusOK,
        Data: "register topic success",
    })
}

/*
   receive the owner info(key, server_name, host), then find the topic_id.
   Before delete the topic, we need del the subscribers first.
*/
func DelTopicHandler(c *gin.Context) {
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
    findResult, err := mongoUtils.FindOne(modules.DbOwner, spc)
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
    delResult, err := mongoUtils.DelMany(modules.DbSubscriber, bson.M{"topic_id": owner.TopicID})
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
    _, err = mongoUtils.DelOne(modules.DbTopic, bson.M{"_id": owner.TopicID})
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
    _, err = mongoUtils.DelOne(modules.DbOwner, bson.M{"_id": owner.ID})
    if err != nil {
        log.Errorf("[delete error] owner id: %v", owner.ID)
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }
    log.Infof("[delete] owner id: %v", owner.ID)
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
    findResult, err := mongoUtils.FindOne(modules.DbOwner, spc)
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

    findResult, err = mongoUtils.FindOne(modules.DbTopic, bson.M{"_id": owner.TopicID})
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
        modules.DbTopic,
        bson.M{"_id": topic.ID},
        bson.M{
            "$set": bson.M{
                "subscribers": topic.Subscribers,
            },
        })
    if err != nil {
        log.Errorf("[update error] topic's _id:%s", topic.ID)
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }
    doc := bson.M{}
    subscriber.ID = primitive.NewObjectID()
    byteSubscriber, _ := bson.Marshal(subscriber)
    _ = bson.Unmarshal(byteSubscriber, &doc)
    _, err = mongoUtils.InsertOne(modules.DbSubscriber, doc)
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
    findResult, err := mongoUtils.FindOne(modules.DbSubscriber, spc)
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
    findResult, err = mongoUtils.FindOne(modules.DbTopic, spc)
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
        modules.DbTopic,
        bson.M{"_id": topic.ID},
        bson.M{
            "$set": bson.M{
                "subscribers": topic.Subscribers,
            },
        })
    if err != nil {
        log.Errorf("[update error] topic's _id:%s", topic.ID)
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }

    spc = bson.M{"_id": subscribe.ID}
    _, err = mongoUtils.DelOne(modules.DbSubscriber, spc)
    if err != nil {
        log.Errorf("[delete error] subscriber id: %v", subscribe.ID)
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }
    log.Infof("[delete] subscriber id: %v", subscribe.ID)

    c.JSON(http.StatusOK, utils.RequestResult{
        Code: http.StatusOK,
        Data: "delete subscriber success",
    })

}
