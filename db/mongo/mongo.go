package mongo

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type DbMongoUtils struct {
	Client     *mongo.Client
	Database   *mongo.Database
	DBName     string
	ServerIP   string
	Port       int
	isAuth     bool
	user       string
	password   string
	authSource string
}

// Utils mongodb tool in global
var Utils DbMongoUtils

// Init mongodb init func
func Init() {
	mongoConfig := viper.GetStringMap("mongo")
	Utils = DbMongoUtils{
		Client:     nil,
		Database:   nil,
		DBName:     mongoConfig["db_name"].(string),
		ServerIP:   mongoConfig["host"].(string),
		Port:       mongoConfig["port"].(int),
		isAuth:     mongoConfig["auth"].(bool),
		user:       mongoConfig["user"].(string),
		password:   viper.GetString("mongoPassword"),
		authSource: mongoConfig["auth_source"].(string),
	}
	Utils.OpenConn()
	Utils.CloseConn()
	Utils.SetDB(Utils.DBName)
	Utils.CloseConn()
}

func (m *DbMongoUtils) Bson2Obj(val interface{}, obj interface{}) error {
	data, err := bson.Marshal(val)
	if err != nil {
		return err
	}
	if err := bson.Unmarshal(data, obj); err != nil {
		return err
	}
	return nil
}

func GetCtx() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	return ctx, cancel
}

func (m *DbMongoUtils) OpenConn() {
	var client *mongo.Client
	var err error
	mongoUri := fmt.Sprintf("mongodb://%s:%d", m.ServerIP, m.Port)
	if _, err := url.Parse(mongoUri); err != nil {
		log.Panicf("%+v", err)
		panic(err)
	}
	opts := &options.ClientOptions{}
	opts.SetAuth(options.Credential{AuthSource: m.authSource, Username: m.user, Password: m.password})
	opts.SetMaxPoolSize(10)
	opts.ApplyURI(mongoUri)

	ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
	if client, err = mongo.Connect(ctx, opts); err != nil {
		log.Panicf("%+v", err)
		panic(err)
	}
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		log.Panicf("%+v", err)
		panic(err)
	}
	m.Client = client
}

func (m *DbMongoUtils) CloseConn() error {
	err := m.Client.Disconnect(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (m *DbMongoUtils) SetDB(db string) {
	if m.Client == nil {
		log.Panicf("mongo's client is nil")
		panic("mongo's client is nil")
	}
	m.Database = m.Client.Database(db)
}

func (m *DbMongoUtils) FindOne(col string, spc bson.M) (*mongo.SingleResult, error) {
	if m.Database == nil || m.Client == nil {
		return nil, errors.New("there is no database or client")
	}
	table := m.Database.Collection(col)
	ctx, cancel := GetCtx()
	defer cancel()
	findResult := table.FindOne(ctx, spc)
	if findResult.Err() != nil {
		return nil, errors.WithMessage(findResult.Err(), "[find error] db error")
	}

	return findResult, nil
}

func (m *DbMongoUtils) FindMore(col string, spc bson.M) ([]bson.M, error) {
	if m.Database == nil || m.Client == nil {
		return nil, fmt.Errorf("there is no database or client")
	}
	table := m.Database.Collection(col)
	ctx, cancel := GetCtx()
	defer cancel()
	cur, err := table.Find(ctx, spc)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var resultArray []bson.M
	for cur.Next(ctx) {
		var result bson.M
		if err := cur.Decode(&result); err != nil {
			return nil, err
		}
		resultArray = append(resultArray, result)
	}
	return resultArray, nil
}

func (m *DbMongoUtils) InsertOne(col string, document bson.M) (result *mongo.InsertOneResult, err error) {
	if m.Database == nil || m.Client == nil {
		return nil, fmt.Errorf("there is no database or client")
	}
	table := m.Database.Collection(col)
	ctx, cancel := GetCtx()
	defer cancel()
	if result, err = table.InsertOne(ctx, document); err != nil {
		return nil, err
	}
	return result, nil
}

func (m *DbMongoUtils) InsertMany(col string, spc bson.M) ([]bson.M, error) {
	return nil, nil
}

func (m *DbMongoUtils) DelOne(col string, spc bson.M) (result *mongo.DeleteResult, err error) {
	if m.Database == nil || m.Client == nil {
		return nil, fmt.Errorf("there is no database or client")
	}
	table := m.Database.Collection(col)
	ctx, cancel := GetCtx()
	defer cancel()
	if result, err = table.DeleteOne(ctx, spc); err != nil {
		return nil, err
	}
	return result, nil
}

func (m *DbMongoUtils) DelMany(col string, spc bson.M) (result *mongo.DeleteResult, err error) {
	if m.Database == nil || m.Client == nil {
		return nil, fmt.Errorf("there is no database or client")
	}
	table := m.Database.Collection(col)
	ctx, cancel := GetCtx()
	defer cancel()
	if result, err = table.DeleteMany(ctx, spc); err != nil {
		return nil, err
	}
	return result, nil
}

func (m *DbMongoUtils) UpdateOne(col string, spc bson.M, doc bson.M) (result *mongo.UpdateResult, err error) {
	if m.Database == nil || m.Client == nil {
		return nil, fmt.Errorf("there is no database or client")
	}
	table := m.Database.Collection(col)
	ctx, cancel := GetCtx()
	defer cancel()
	if result, err = table.UpdateOne(ctx, spc, doc); err != nil {
		return nil, err
	}
	return result, nil
}

func (m *DbMongoUtils) UpdateMany(col string, spc bson.M) ([]bson.M, error) {
	return nil, nil
}

func (m *DbMongoUtils) Count(col string, spc bson.M) (result int64, err error) {
	if m.Database == nil || m.Client == nil {
		return 0, fmt.Errorf("there is no database or client")
	}
	table := m.Database.Collection(col)
	ctx, cancel := GetCtx()
	defer cancel()

	if result, err = table.CountDocuments(ctx, spc); err != nil {
		return 0, err
	}
	return 0, nil
}

func (m *DbMongoUtils) Distinct(col string, filed string, spc bson.M) (result interface{}, err error) {
	if m.Database == nil || m.Client == nil {
		return nil, fmt.Errorf("there is no database or client")
	}
	table := m.Database.Collection(col)
	ctx, cancel := GetCtx()
	defer cancel()
	if result, err = table.Distinct(ctx, filed, spc); err != nil {
		return nil, err
	}
	return result, nil
}
