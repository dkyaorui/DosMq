package mongo

import (
    "go.mongodb.org/mongo-driver/bson/primitive"
    "time"
)

type RequestMessage struct {
    RecvRequest []byte    `bson:"receive_request"`
    Created     time.Time `bson:"created"`
}

// 主题
type Topic struct {
    Name               string       `bson:"name"`
    Subscribers        []Subscriber `bson:"subscribers"`
    Owner              Owner        `bson:"owner"`
    ProcessMessageType string       `bson:"process_message_type"` // pull or push
}

// 消息
type Message struct {
    TopicId   primitive.ObjectID `bson:"topic_id"`
    Value     []byte             `bson:"value"`
    Timestamp int64              `bson:"create_time"`
}

// 订阅者
type Subscriber struct {
    ServerName string `bson:"sub_server_name"`
    Host       string `bson:"sub_Host"`
    Key        string `bson:"sub_key"`
    Api        string `bson:"sub_api"`
    Method     string `bson:"sub_method"`
    Protocol   string `bson:"sub_protocol"`
}

// 发布者
type Owner struct {
    ServerName string `bson:"owner_server_name"`
    Key        string `bson:"owner_key"` // 发布者上传key
    Host       string `bson:"owner_host"`
}
