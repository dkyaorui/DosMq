package consumer

import (
    mongodb "DosMq/db/mongo"
    mongoModule "DosMq/modules/mongo"
    "context"
    "fmt"
    "github.com/gin-gonic/gin"
    log "github.com/sirupsen/logrus"
    "net/http"
    "net/http/httputil"
    "time"
)

func Hello(c *gin.Context) {
    var recvReq []byte
    var err error
    if recvReq, err = httputil.DumpRequest(c.Request, true); err != nil {
        log.Errorf("%+v", err)
        return
    }
    recvRequest := mongoModule.RequestMessage{
        RecvRequest: recvReq,
        Created:     time.Now(),
    }
    fmt.Printf(c.PostForm("name"))
    mongoUtils := mongodb.Utils
    mongoUtils.OpenConn()
    mongoUtils.SetDB(mongoUtils.DBName)
    defer mongoUtils.CloseConn()

    collection := mongoUtils.Database.Collection("recv_requests")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    result, err := collection.InsertOne(ctx, recvRequest)
    if err != nil {
        log.Error("insert error")
        return
    }
    log.Infof("insert %s", result.InsertedID)
    //coon := redis.RedisClient.Get()
    //defer coon.Close()
    //if _, err := redis.Insert(coon, "name", name); err != nil {
    //    log.Error(err)
    //}
    println("hahahahahha")
    c.JSON(http.StatusOK, gin.H{"receive": recvRequest})
}
