package topicCenter

import (
    MongoModule "DosMq/modules/mongo"
    "DosMq/utils"
    "fmt"
    "github.com/gin-gonic/gin"
    "net/http"
)

func TopicRegister(c *gin.Context) {
    var requestResult utils.RequestResult
    var err error

    key := c.PostForm("auth_key")
    if err = utils.CheckRegisterTopicKey(key); err != nil {
        requestResult = utils.RequestResult{
            Code: http.StatusNonAuthoritativeInfo,
            Data: err,
        }
        c.JSON(http.StatusNonAuthoritativeInfo, requestResult)
        return
    }
    var owner MongoModule.Owner
    //var topic MongoModule.Topic
    fmt.Println(c.PostForm("server_name"))
    if err = c.BindJSON(&owner); err != nil {
        requestResult = utils.RequestResult{
            Code: http.StatusPartialContent,
            Data: err.Error(),
        }
        c.JSON(http.StatusPartialContent, requestResult)
        return
    }
    fmt.Println(owner)

    //if err = c.BindJSON(&topic); err!= nil {
    //    requestResult = utils.RequestResult{
    //        Code: http.StatusPartialContent,
    //        Data: err,
    //    }
    //    c.JSON(http.StatusPartialContent, requestResult)
    //    return
    //}

    //topicName := c.PostForm("topicName")
    //err = json.Unmarshal([]byte(c.PostForm("owner")), &owner)
    //if err != nil {
    //    fmt.Println(err)
    //}
    //processMessageType := c.PostForm("processMessageType")
    //fmt.Println(topicName, owner, processMessageType)
    //var topic = MongoModule.Topic{
    //    Name:               topicName,
    //    Subscribers:        make([]MongoModule.Subscriber, 0),
    //    Owner:              owner,
    //    ProcessMessageType: processMessageType,
    //}
    //mongoUtils := Mongodb.Utils
    //mongoUtils.OpenConn()
    //mongoUtils.SetDB(mongoUtils.DBName)
    //defer mongoUtils.CloseConn()
    //collection := mongoUtils.Database.Collection("topic")
    //ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    //defer cancel()
    //var result MongoModule.Topic
    //if err := collection.FindOne(ctx, bson.M{"name": topic.Name}).Decode(&result); err == nil {
    //    requestResult = utils.RequestResult{
    //        Code: http.StatusOK,
    //        Data: err,
    //    }
    //    c.JSON(http.StatusOK, bson.M{"res": "已存在"})
    //    return
    //}
    ////fmt.Println(result)
    ////result, _ := collection.InsertOne(ctx, topic)
    ////fmt.Println(result.InsertedID)
    //requestResult = utils.RequestResult{
    //    Code: http.StatusOK,
    //    Data: result,
    //}
    //c.JSON(http.StatusOK, requestResult)
}
