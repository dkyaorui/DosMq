package broker

import (
    Mongodb "DosMq/db/mongo"
    MongoModule "DosMq/modules/mongo"
    "DosMq/utils"
    "context"
    "encoding/json"
    "fmt"
    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "net/http"
    "time"
)

func TopicRegister(c *gin.Context) {
    var requestResult utils.RequestResult
    var subscribers []MongoModule.Subscriber
    var owner MongoModule.Owner
    topicName := c.PostForm("topicName")
    err := json.Unmarshal([]byte(c.PostForm("subscribers")), &subscribers)
    if err != nil {
        fmt.Println(err)
    }
    err = json.Unmarshal([]byte(c.PostForm("owner")), &owner)
    if err != nil {
        fmt.Println(err)
    }
    processMessageType := c.PostForm("processMessageType")
    fmt.Println(topicName, owner, processMessageType)
    for _, v := range subscribers {
        fmt.Println(v)
        fmt.Println(v.ServerName)
    }
    var topic = MongoModule.Topic{
        Name:               topicName,
        Subscribers:        subscribers,
        Owner:              owner,
        ProcessMessageType: processMessageType,
    }
    mongoUtils := Mongodb.Utils
    mongoUtils.OpenConn()
    mongoUtils.SetDB(mongoUtils.DBName)
    defer mongoUtils.CloseConn()
    collection := mongoUtils.Database.Collection("topic")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    var result MongoModule.Topic
    if err := collection.FindOne(ctx, bson.M{"name": topic.Name}).Decode(&result); err == nil {
        requestResult = utils.RequestResult{
            Code: http.StatusOK,
            Data: err,
        }
        c.JSON(http.StatusOK, bson.M{"res": "已存在"})
        return
    }
    //fmt.Println(result)
    //result, _ := collection.InsertOne(ctx, topic)
    //fmt.Println(result.InsertedID)
    requestResult = utils.RequestResult{
        Code: http.StatusOK,
        Data: result,
    }
    c.JSON(http.StatusOK, requestResult)
}
