package modules

import (
	myMongo "DosMq/db/mongo"
	"bytes"
	"crypto/md5"
	"strings"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	// DbSubscriber const
	DbSubscriber = "subscriber"
	// DbTopic const
	DbTopic = "topic"
	// DbOwner const
	DbOwner = "owner"
	// DbMessage const
	DbMessage = "message"
)

// AuthKey The required params of requests
type AuthKey struct {
	Value string `json:"auth_key" binding:"required"`
}

// ModuleUtils some tools of Module
type ModuleUtils interface {
	GetHashCode() []byte
	GetRedisKey() string
	// return false if not exist
	CheckIsRepeat(mongoUtils *myMongo.DbMongoUtils) (bool, error)
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

// Topic collection_name: topic
type Topic struct {
	ID                 [12]byte
	Name               [16]byte
	ProcessMessageType [4]byte // pull or push
	HashCode           [16]byte
	Owner              Owner
	Subscribers        []Subscriber
}

/*
Message collection_name: message

push the message to their topic's message queue.If the queue is full,
save the message into redis.

when consumer use the message,query in redis first and if no data in redis then
get message in queue.when the message is used then save it in mongodb.
*/
type Message struct {
	ID          [12]byte
	TopicID     [12]byte
	HashCode    [16]byte
	Timestamp   int64
	OffSet      int64
	MessageSize int32
	Value       string
}

// Subscriber collection_name: subscriber
type Subscriber struct {
	ID         [12]byte
	ServerName [16]byte
	HashCode   [16]byte
	TopicID    [12]byte
	Method     [4]byte
	Host       string
	API        string
	Key        string
}

// Owner collection_name: owner
type Owner struct {
	ID         [12]byte
	TopicID    [12]byte
	HashCode   [16]byte
	ServerName [16]byte
	Key        string // Owner's key
	Host       string
}

// GetHashCode generate topic hashcode
// hash_code = name + process_message_type + owner.host + owner.key
func (t *Topic) GetHashCode() []byte {
	var buffer bytes.Buffer
	buffer.Write(t.Name[:])
	buffer.Write(t.ProcessMessageType[:])
	buffer.Write([]byte(t.Owner.Key + t.Owner.Host))
	hashCode := md5.Sum(buffer.Bytes())
	return hashCode[:]
}

// GetHashCode generate Message hashcode
// hash_code = topicA_id + value
func (m *Message) GetHashCode() []byte {
	var buffer bytes.Buffer
	buffer.Write(m.TopicID[:])
	buffer.Write([]byte(m.Value))
	hashCode := md5.Sum(buffer.Bytes())
	return hashCode[:]
}

// GetHashCode generate Subscriber hashcode
// hash_code = server_name + host + key + api + method
func (s *Subscriber) GetHashCode() []byte {
	var buffer bytes.Buffer
	buffer.Write(s.ServerName[:])
	buffer.Write([]byte(s.Host + s.Key + s.API))
	buffer.Write(s.Method[:])
	hashCode := md5.Sum(buffer.Bytes())
	return hashCode[:]
}

// GetHashCode generate Owner hashcode
// hash_code = host + key + ServerName
func (o *Owner) GetHashCode() []byte {
	var buffer bytes.Buffer
	buffer.Write([]byte(o.Host + o.Key))
	buffer.Write(o.ServerName[:])
	hashCode := md5.Sum(buffer.Bytes())
	return hashCode[:]
}

// Equal equal two topic struct
func (t *Topic) Equal(target Topic) bool {
	if bytes.Compare(t.HashCode[:], target.HashCode[:]) == 0 {
		if bytes.Compare(t.Name[:], target.Name[:]) == 0 &&
			bytes.Compare(t.ProcessMessageType[:], target.ProcessMessageType[:]) == 0 &&
			strings.Compare(t.Owner.Host, target.Owner.Host) == 0 &&
			strings.Compare(t.Owner.Key, target.Owner.Key) == 0 {
			return true
		}
		return false
	}
	return false
}

// Equal equal two Message struct
func (m *Message) Equal(target Message) bool {
	if bytes.Compare(m.HashCode[:], target.HashCode[:]) == 0 {
		if bytes.Compare(m.TopicID[:], target.TopicID[:]) == 0 &&
			strings.Compare(m.Value, target.Value) == 0 {
			return true
		}
		return false
	}
	return false
}

// Equal equal two Subscriber struct
func (s *Subscriber) Equal(target Subscriber) bool {
	if bytes.Compare(s.HashCode[:], target.HashCode[:]) == 0 {
		if bytes.Compare(s.ServerName[:], target.ServerName[:]) == 0 &&
			strings.Compare(s.Host, target.Host) == 0 &&
			strings.Compare(s.Key, target.Key) == 0 &&
			strings.Compare(s.API, target.API) == 0 &&
			bytes.Compare(s.Method[:], target.Method[:]) == 0 {
			return true
		}
		return false
	}
	return false

}

// Equal equal two Owner struct
func (o *Owner) Equal(target Owner) bool {
	if bytes.Compare(o.HashCode[:], target.HashCode[:]) == 0 {
		if strings.Compare(o.Host, target.Host) == 0 &&
			strings.Compare(o.Key, target.Key) == 0 &&
			bytes.Compare(o.ServerName[:], target.ServerName[:]) == 0 {
			return true
		}
		return false
	}
	return false

}

// CheckIsRepeat check the topic is repeat
func (t *Topic) CheckIsRepeat(mongoUtils *myMongo.DbMongoUtils) (bool, error) {
	spc := bson.M{"hash_code": t.HashCode}
	findResult, err := mongoUtils.FindOne(DbTopic, spc)
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

// CheckIsRepeat check the topic is repeat
func (m *Message) CheckIsRepeat(mongoUtils *myMongo.DbMongoUtils) (bool, error) {
	spc := bson.M{"hash_code": m.HashCode}
	findResult, err := mongoUtils.FindOne(DbMessage, spc)
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

// CheckIsRepeat check the topic is repeat
func (s *Subscriber) CheckIsRepeat(mongoUtils *myMongo.DbMongoUtils) (bool, error) {
	spc := bson.M{"hash_code": s.HashCode}
	findResult, err := mongoUtils.FindOne(DbSubscriber, spc)
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

// CheckIsRepeat check the topic is repeat
func (o *Owner) CheckIsRepeat(mongoUtils *myMongo.DbMongoUtils) (bool, error) {
	spc := bson.M{"hash_code": o.HashCode}
	findResult, err := mongoUtils.FindOne(DbOwner, spc)
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

// FindIndexInSubscribers find the topic's subscribe by hashCode
func (t *Topic) FindIndexInSubscribers(target Subscriber) int {
	for index, sub := range t.Subscribers {
		if bytes.Compare(sub.HashCode[:], target.HashCode[:]) == 0 {
			return index
		}
	}
	return -1
}

// DelSubscribe del one subscribe of topic
func (t *Topic) DelSubscribe(target Subscriber) {
	subIndex := t.FindIndexInSubscribers(target)
	if subIndex != -1 {
		t.Subscribers = append(t.Subscribers[:subIndex], t.Subscribers[subIndex+1:]...)
	}
}

// GetAllSubscribers return all subscribers of topic
func (t *Topic) GetAllSubscribers() ([]Subscriber, error) {
	mongoUtils := myMongo.Utils
	mongoUtils.OpenConn()
	mongoUtils.SetDB(mongoUtils.DBName)
	defer mongoUtils.CloseConn()

	findResult, err := mongoUtils.FindMore(DbSubscriber, bson.M{"topic_id": t.ID})
	if err != nil {
		return nil, err
	}
	resultLength := len(findResult)
	subscribers := make([]Subscriber, resultLength)
	for index, item := range findResult {
		var subscriber Subscriber
		itemByte, _ := bson.Marshal(item)
		_ = bson.Unmarshal(itemByte, &subscriber)
		subscribers[index] = subscriber
	}
	return subscribers, nil
}

// func (t *Topic) GetRedisKey() string {
// 	return "topic_" + hex.EncodeToString(t.HashCode)
// }

// func (m *Message) GetRedisKey() string {
// 	return "message_item_" + hex.EncodeToString(m.ID[:])
// }

// func (s *Subscriber) GetRedisKey() string {
// 	return "subscriber_" + hex.EncodeToString(s.HashCode)
// }

// func (o *Owner) GetRedisKey() string {
// 	return "owner_" + hex.EncodeToString(o.HashCode)
// }

// func (m *Message) GetMqKey() string {
// 	return "message_" + hex.EncodeToString(m.TopicId[:])
// }

// func GetMqKeyByTopicId(topicID string) string {
// 	return "message_" + topicID
// }
