package mongo

import (
    myMongo "DosMq/db/mongo"
    "bytes"
    "crypto/md5"
    "encoding/hex"
    "fmt"
    "github.com/pkg/errors"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "strings"
    "time"
)

const (
    DB_SUBSCRIBER = "subscriber"
    DB_TOPIC      = "topic"
    DB_OWNER      = "owner"
    DB_MESSAGE    = "message"
)

type ModuleUtils interface {
    GetHashCode() []byte
    GetRedisKey() string
    // if not exist false
    CheckIsRepeat(mongoUtils *myMongo.DbMongoUtils) (bool, error)
}

type RequestMessage struct {
    RecvRequest []byte    `bson:"receive_request"`
    Created     time.Time `bson:"created"`
}

/*
   订阅者：
   1. 当订阅方式为pull时，通过subscriber的topic_id校验订阅权限
   2. 当订阅方式为push时，通过topic的subscribers向订阅者发送消息
   3. 订阅者订阅消息时，需要在topic中的subscribers和subscriber中同时记录，方便数据读取和后续功能添加
   发布者：
   1. 当发布者注册一个主题时，需要携带配置文件中topic.key，防止恶意请求注册无效topic
   2. 当发布者注册一个主题时，需要在topic表中和owner表同时记录数据，方便数据读取和后续功能添加
   3. 发布者持有的key作为安全检查的一个方法：
       （1）当订阅方式为pull时，订阅者携带的key需要和发布者持有的key一致
       （2）当订阅方式为push时，订阅者可以选择在自身服务中对key做校验处理，防止恶意请求导致服务器处理无效请求
*/

// collection_name: topic
type Topic struct {
    Id                 primitive.ObjectID `bson:"_id" json:"id" binding:"-"`
    Name               string             `bson:"name" json:"topic_name" binding:"required"`
    ProcessMessageType string             `bson:"process_message_type" json:"topic_process_type" binding:"required"` // pull or push
    Subscribers        []Subscriber       `bson:"subscribers" json:"topic_subscribers" binding:"-"`
    Owner              Owner              `bson:"owner" json:"topic_owner" binding:"-"`
    HashCode           []byte             `bson:"hash_code" json:"hash_code" binding:"-"`
}

/*
消息
collection_name: message

push the message to their topic's message queue.If the queue is full,
save the message into redis.

when consumer use the message,query in redis first and if no data in redis then
get message in queue.when the message is used then save it in mongodb.
*/
type Message struct {
    Id        primitive.ObjectID `bson:"_id" json:"id" binding:"-"`
    TopicId   primitive.ObjectID `bson:"topic_id" json:"topic_id" binding:"-"`
    Value     string             `bson:"value" json:"value" binding:"required"`
    Timestamp int64              `bson:"create_time" json:"timestamp" binding:"-"`
    HashCode  []byte             `bson:"hash_code" json:"hash_code" binding:"-"`
}

// 订阅者
// collection_name: subscriber
type Subscriber struct {
    Id         primitive.ObjectID `bson:"_id" json:"id" binding:"-"`
    ServerName string             `bson:"sub_server_name" json:"sub_server_name" binding:"required"`
    Host       string             `bson:"sub_Host" json:"sub_host" binding:"required"`
    Key        string             `bson:"sub_key" json:"sub_key" binding:"required"`
    Api        string             `bson:"sub_api" json:"sub_api" binding:"required"`
    Method     string             `bson:"sub_method" json:"sub_method" binding:"required"`
    TopicId    primitive.ObjectID `bson:"topic_id" json:"sub_topic_id" binding:"-"`
    HashCode   []byte             `bson:"hash_code" json:"hash_code" binding:"-"`
}

// 发布者
// collection_name: owner
type Owner struct {
    Id         primitive.ObjectID `bson:"_id" json:"id" binding:"-"`
    ServerName string             `bson:"owner_server_name" json:"owner_server_name" binding:"required"`
    Key        string             `bson:"owner_key" json:"owner_key" binding:"required"` // Owner's key
    Host       string             `bson:"owner_host" json:"owner_host" binding:"required"`
    TopicID    primitive.ObjectID `bson:"topic_id" json:"owner_topic_id" binding:"-"`
    HashCode   []byte             `bson:"hash_code" json:"hash_code" binding:"-"`
}

func (t *Topic) GetHashCode() []byte {
    // hash_code = name + process_message_type + owner.host + owner.key
    hashCode := md5.Sum([]byte(t.Name + t.ProcessMessageType + t.Owner.Key + t.Owner.Host))
    return hashCode[:]
}

func (m *Message) GetHashCode() []byte {
    // hash_code = topicA_id + value
    var buffer bytes.Buffer
    buffer.Write(m.TopicId[:])
    buffer.Write([]byte(m.Value))
    hashCode := md5.Sum(buffer.Bytes())
    return hashCode[:]
}

func (s *Subscriber) GetHashCode() []byte {
    // hash_code = server_name + host + key + api + method
    hashCode := md5.Sum([]byte(s.ServerName + s.Host + s.Key + s.Api + s.Method))
    fmt.Print(hashCode)
    return hashCode[:]
}

func (o *Owner) GetHashCode() []byte {
    // hash_code = host + key + ServerName
    var buffer bytes.Buffer
    buffer.Write([]byte(o.Host + o.Key + o.ServerName))
    hashCode := md5.Sum(buffer.Bytes())
    return hashCode[:]
}

func (t *Topic) Equal(target Topic) bool {
    if bytes.Compare(t.HashCode, target.HashCode) == 0 {
        if strings.Compare(t.Name, target.Name) == 0 &&
            strings.Compare(t.ProcessMessageType, target.ProcessMessageType) == 0 &&
            strings.Compare(t.Owner.Host, target.Owner.Host) == 0 &&
            strings.Compare(t.Owner.Key, target.Owner.Key) == 0 {
            return true
        }
        return false
    } else {
        return false
    }
}

func (m *Message) Equal(target Message) bool {
    if bytes.Compare(m.HashCode, target.HashCode) == 0 {
        if bytes.Compare(m.TopicId[:], target.TopicId[:]) == 0 &&
            strings.Compare(m.Value, target.Value) == 0 {
            return true
        }
        return false
    } else {
        return false
    }
}

func (s *Subscriber) Equal(target Subscriber) bool {
    if bytes.Compare(s.HashCode, target.HashCode) == 0 {
        if strings.Compare(s.ServerName, target.ServerName) == 0 &&
            strings.Compare(s.Host, target.Host) == 0 &&
            strings.Compare(s.Key, target.Key) == 0 &&
            strings.Compare(s.Api, target.Api) == 0 &&
            strings.Compare(s.Method, target.Method) == 0 {
            return true
        }
        return false
    } else {
        return false
    }
}

func (o *Owner) Equal(target Owner) bool {
    if bytes.Compare(o.HashCode, target.HashCode) == 0 {
        if strings.Compare(o.Host, target.Host) == 0 &&
            strings.Compare(o.Key, target.Key) == 0 &&
            strings.Compare(o.ServerName, target.ServerName) == 0 {
            return true
        }
        return false
    } else {
        return false
    }
}

func (t *Topic) CheckIsRepeat(mongoUtils *myMongo.DbMongoUtils) (bool, error) {
    spc := bson.M{"hash_code": t.HashCode}
    findResult, err := mongoUtils.FindOne(DB_TOPIC, spc)
    if err != nil {
        if strings.Contains(err.Error(), "no documents in result") {
            return false, nil
        }
        return true, errors.Wrap(err, "no data")
    }
    var topic Topic
    if err = findResult.Decode(&topic); err != nil {
        return true, err
    }
    if t.Equal(topic) {
        return true, nil
    }
    return false, nil
}
func (m *Message) CheckIsRepeat(mongoUtils *myMongo.DbMongoUtils) (bool, error) {
    spc := bson.M{"hash_code": m.HashCode}
    findResult, err := mongoUtils.FindOne(DB_MESSAGE, spc)
    if err != nil {
        if strings.Contains(err.Error(), "no documents in result") {
            return false, nil
        }
        return true, errors.Wrap(err, "no data")
    }
    var message Message
    if err = findResult.Decode(&message); err != nil {
        return true, err
    }
    if m.Equal(message) {
        return true, nil
    }
    return false, nil
}

func (s *Subscriber) CheckIsRepeat(mongoUtils *myMongo.DbMongoUtils) (bool, error) {
    spc := bson.M{"hash_code": s.HashCode}
    findResult, err := mongoUtils.FindOne(DB_SUBSCRIBER, spc)
    if err != nil {
        if strings.Contains(err.Error(), "no documents in result") {
            return false, nil
        }
        return true, errors.Wrap(err, "no data")
    }
    var subscriber Subscriber
    if err = findResult.Decode(&subscriber); err != nil {
        return true, err
    }
    if s.Equal(subscriber) {
        return true, nil
    }
    return false, nil
}
func (o *Owner) CheckIsRepeat(mongoUtils *myMongo.DbMongoUtils) (bool, error) {
    spc := bson.M{"hash_code": o.HashCode}
    findResult, err := mongoUtils.FindOne(DB_OWNER, spc)
    if err != nil {
        if strings.Contains(err.Error(), "no documents in result") {
            return false, nil
        }
        return true, errors.Wrap(err, "no data")
    }
    var owner Owner
    if err = findResult.Decode(&owner); err != nil {
        return true, err
    }
    if o.Equal(owner) {
        return true, nil
    }
    return false, nil
}

func (t *Topic) FindIndexInSubscribers(target Subscriber) int {
    for index, sub := range t.Subscribers {
        if bytes.Compare(sub.HashCode, target.HashCode) == 0 {
            return index
        }
    }
    return -1
}

func (t *Topic) DelSubscribe(target Subscriber) {
    subIndex := t.FindIndexInSubscribers(target)
    if subIndex != -1 {
        t.Subscribers = append(t.Subscribers[:subIndex], t.Subscribers[subIndex+1:]...)
    }
}

func (t *Topic) GetRedisKey() string {
    return "topic_" + hex.EncodeToString(t.HashCode)
}

func (m *Message) GetRedisKey() string {
    return "message_item_" + hex.EncodeToString(m.Id[:])
}

func (s *Subscriber) GetRedisKey() string {
    return "subscriber_" + hex.EncodeToString(s.HashCode)
}

func (o *Owner) GetRedisKey() string {
    return "owner_" + hex.EncodeToString(o.HashCode)
}

func (m *Message) GetMqKey() string {
    return "message_" + hex.EncodeToString(m.TopicId[:])
}

func GetMqKeyByTopicId(topicID string) string {
    return "message_" + topicID
}
