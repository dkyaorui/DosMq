package producer

import (
    "DosMq/db/mongo"
    mongoModule "DosMq/modules/mongo"
    "bufio"
    "bytes"
    "context"
    "fmt"
    "github.com/gin-gonic/gin"
    log "github.com/sirupsen/logrus"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "net/http"
    "strings"
    "time"
)

func Hello(c *gin.Context) {
    var ans mongoModule.RequestMessage
    dataBase := mongo.GetMongoDataBase()
    collection := dataBase.Collection("recv_requests")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    objId, _ := primitive.ObjectIDFromHex("5dc90044f17b859e3bbd87f0")
    err := collection.FindOne(ctx, bson.M{"_id": objId}).Decode(&ans)
    if err != nil {
        log.Error("find error")
        return
    }
    x, _ := http.ReadRequest(bufio.NewReader(bytes.NewReader(ans.RecvRequest)))

    reqUri := ""
    if strings.HasPrefix(x.Proto, "HTTPS") {
        reqUri = "https://" + x.Host + x.RequestURI
    } else {
        reqUri = "http://" + x.Host + x.RequestURI
    }

    req, err := http.NewRequest(x.Method, reqUri, x.Body)

    if err != nil {
        log.Error(err)
        return
    }
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        log.Error(err)
        return
    }
    defer resp.Body.Close()
    fmt.Println("req", resp)
    c.JSON(http.StatusOK, ans)
}
