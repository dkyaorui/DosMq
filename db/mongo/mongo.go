package mongo

import (
    "context"
    "fmt"
    log "github.com/sirupsen/logrus"
    "github.com/spf13/viper"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "os"
)

var mongoOptions *options.ClientOptions
var client *mongo.Client
var database *mongo.Database

func Init() {
    var err error
    mongoConfig := viper.GetStringMap("mongo")
    mongoUri := fmt.Sprintf("mongodb://%s:%s@%s:%d", mongoConfig["user"], os.Getenv("MONGO_PASSWORD"),
        mongoConfig["host"].(string), mongoConfig["port"].(int))
    mongoOptions = options.Client().ApplyURI(mongoUri)
    if client, err = mongo.Connect(context.TODO(), mongoOptions); err != nil {
        log.Fatalf("%+v", err)
    }
    err = client.Ping(context.TODO(), nil)
    if err != nil {
        log.Fatalf("%+v", err)
    }
    database = client.Database(viper.GetString("mongo.db"))
}

func GetMongoDataBase() *mongo.Database {
    return database
}
