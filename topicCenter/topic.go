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
    var requestResult utils.RequestResult
    var err error
    // check data
    if err = c.ShouldBindBodyWith(&owner, binding.JSON); err != nil {
        log.Infof("binding data error,struct:%s", "owner")
        requestResult = utils.RequestResult{
            Code: http.StatusPartialContent,
            Data: err.Error(),
        }
        c.AbortWithStatusJSON(http.StatusPartialContent, requestResult)
    }
    if err = c.ShouldBindBodyWith(&topic, binding.JSON); err != nil {
        log.Infof("binding data error,struct:%s", "topic")
        requestResult = utils.RequestResult{
            Code: http.StatusPartialContent,
            Data: err.Error(),
        }
        c.AbortWithStatusJSON(http.StatusPartialContent, requestResult)
    }

    // open mongo client
    mongoUtils := Mongodb.Utils
    mongoUtils.OpenConn()
    mongoUtils.SetDB(mongoUtils.DBName)
    defer mongoUtils.CloseConn()

    // curd
    byteTopic, _ := bson.Marshal(&topic)
    doc := bson.M{}
    _ = bson.Unmarshal(byteTopic, &doc)
    insertResult, err := mongoUtils.InsertOne("topic", doc)
    if err != nil {
        log.Infof("insert data error. data:%v", topic)
        c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
            Code: http.StatusBadGateway,
            Data: err.Error(),
        })
    } else {
        log.Infof("[insert]topic:%v", topic)
        owner.TopicID = insertResult.InsertedID.(primitive.ObjectID)
        topic.Owner = owner
        _, err = mongoUtils.UpdateOne(
            "topic",
            bson.M{"_id": insertResult.InsertedID,},
            bson.M{
                "$set": bson.M{
                    "owner": owner,
                },
            })
        if err != nil {
            log.Errorf("[update error] topic's _id:%s", insertResult.InsertedID)
            _, _ = mongoUtils.DelOne("topic", bson.M{"_id": insertResult.InsertedID})
            c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
                Code: http.StatusBadGateway,
                Data: err.Error(),
            })
        }
        log.Infof("[update]topic:%v", topic)
        byteOwner, _ := bson.Marshal(&owner)
        _ = bson.Unmarshal(byteOwner, &doc)
        _, err = mongoUtils.InsertOne("owner", doc)
        if err != nil {
            log.Infof("insert data error. data:%v", topic)
            _, _ = mongoUtils.DelOne("topic", bson.M{"_id": insertResult.InsertedID})
            c.AbortWithStatusJSON(http.StatusBadGateway, utils.RequestResult{
                Code: http.StatusBadGateway,
                Data: err.Error(),
            })
        }
        log.Infof("[insert]owner:%v", owner)
    }
    requestResult = utils.RequestResult{
        Code: http.StatusOK,
        Data: "register topic success",
    }
    c.JSON(http.StatusOK, requestResult)
}

/*
   receive the owner info(key, server_name, host), then find the topic_id.
   Before delete the topic, we need del the subscribers first.
*/
func TopicDel(c *gin.Context) {
    var owner MongoModule.Owner
    //var topic MongoModule.Topic
    var requestResult utils.RequestResult
    var err error
    if err = c.ShouldBindBodyWith(&owner, binding.JSON); err != nil {
        log.Infof("binding data error,struct:%s", "owner")
        requestResult = utils.RequestResult{
            Code: http.StatusPartialContent,
            Data: err.Error(),
        }
        c.AbortWithStatusJSON(http.StatusPartialContent, requestResult)
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
