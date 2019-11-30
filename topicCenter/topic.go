package topicCenter

import (
    Mongodb "DosMq/db/mongo"
    MongoModule "DosMq/modules/mongo"
    "DosMq/utils"
    "fmt"
    "github.com/gin-gonic/gin"
    "github.com/gin-gonic/gin/binding"
    log "github.com/sirupsen/logrus"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "net/http"
)

func TopicRegister(c *gin.Context) {
    var owner MongoModule.Owner
    var topic MongoModule.Topic
    var err error
    // check data
    if err = c.ShouldBindBodyWith(&owner, binding.JSON); err != nil {
        log.Infof("binding data error,struct:%s", "owner")
        c.AbortWithStatusJSON(http.StatusPartialContent, utils.RequestResult{
            Code: http.StatusPartialContent,
            Data: err.Error(),
        })
    }
    if err = c.ShouldBindBodyWith(&topic, binding.JSON); err != nil {
        log.Infof("binding data error,struct:%s", "topic")
        c.AbortWithStatusJSON(http.StatusPartialContent, utils.RequestResult{
            Code: http.StatusPartialContent,
            Data: err.Error(),
        })
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
    _, err = mongoUtils.InsertOne("topic", doc)
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
    _ = bson.Unmarshal(byteOwner, &doc)
    _, err = mongoUtils.InsertOne("owner", doc)
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

    c.JSON(http.StatusOK, utils.RequestResult{
        Code: http.StatusOK,
        Data: "register topic success",
    })
}

/*
   receive the owner info(key, server_name, host), then find the topic_id.
   Before delete the topic, we need del the subscribers first.
*/
func TopicDel(c *gin.Context) {
    var owner MongoModule.Owner
    var err error
    if err = c.ShouldBindBodyWith(&owner, binding.JSON); err != nil {
        log.Infof("binding data error,struct:%s", "owner")
        c.AbortWithStatusJSON(http.StatusPartialContent, utils.RequestResult{
            Code: http.StatusPartialContent,
            Data: err.Error(),
        })
    }
    // open mongo client
    mongoUtils := Mongodb.Utils
    mongoUtils.OpenConn()
    mongoUtils.SetDB(mongoUtils.DBName)
    defer mongoUtils.CloseConn()
    fmt.Println("del topic")
    // main
    spc := bson.M{}
    byteOwner, _ := bson.Marshal(&owner)
    _ = bson.Unmarshal(byteOwner, &spc)
    findResult, err := mongoUtils.FindOne("owner", spc)
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
    delResult, err := mongoUtils.DelMany("subscriber", bson.M{"topic_id": owner.TopicID,})
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
    _, err = mongoUtils.DelOne("topic", bson.M{"_id": owner.TopicID,})
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
    _, err = mongoUtils.DelOne("owner", bson.M{"_id": owner.Id,})
    if err != nil {
        log.Errorf("[delete error] owner id: %v", owner.Id)
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
        return
    }
    log.Infof("[delete] owner id: %v", owner.Id)

    c.JSON(http.StatusOK, utils.RequestResult{
        Code: http.StatusOK,
        Data: "del success",
    })
}

func SubscribeNews(c *gin.Context) {
    var owner MongoModule.Owner
    var subscriber MongoModule.Subscriber
    var topic MongoModule.Topic
    var err error

    if err = c.ShouldBindBodyWith(&owner, binding.JSON); err != nil {
        log.Infof("binding data error,struct:%s", "owner")
        c.AbortWithStatusJSON(http.StatusPartialContent, utils.RequestResult{
            Code: http.StatusPartialContent,
            Data: err.Error(),
        })
    }
    if err = c.ShouldBindBodyWith(&subscriber, binding.JSON); err != nil {
        log.Infof("binding data error,struct:%s", "subscriber")
        c.AbortWithStatusJSON(http.StatusPartialContent, utils.RequestResult{
            Code: http.StatusPartialContent,
            Data: err.Error(),
        })
    }
    // open mongo client
    mongoUtils := Mongodb.Utils
    mongoUtils.OpenConn()
    mongoUtils.SetDB(mongoUtils.DBName)
    defer mongoUtils.CloseConn()
    fmt.Println("del topic")
    // main
    spc := bson.M{}
    byteOwner, _ := bson.Marshal(&owner)
    _ = bson.Unmarshal(byteOwner, &spc)
    findResult, err := mongoUtils.FindOne("owner", spc)
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

    findResult, err = mongoUtils.FindOne("topic", bson.M{"_id": owner.TopicID,})
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
    if topic.Subscribers == nil {
        topic.Subscribers = []MongoModule.Subscriber{subscriber}
    } else {
        topic.Subscribers = append(topic.Subscribers, subscriber)
    }
    _, err = mongoUtils.UpdateOne(
        "topic",
        bson.M{"_id": topic.Id,},
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
    byteSubscriber, _ := bson.Marshal(subscriber)
    _ = bson.Unmarshal(byteSubscriber, &doc)
    _, err = mongoUtils.InsertOne("subscriber", doc)
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
