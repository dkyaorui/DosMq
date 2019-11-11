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
    Name       string       `bson:"name"`
    Subscriber []Subscriber `bson:"subscriber"`
    Owner      Owner        `bson:"owner"`
    Target     string       `bson:"target"`
    Type       bool         `bson:"type"`
}

// 消息
type Message struct {
    TopicId   primitive.ObjectID `bson:"topic_id"`
    Value     []byte             `bson:"value"`
    Timestamp int64              `bson:"create_time"`
}

// 订阅者
type Subscriber struct {
    Name   string
    Host   string
    Key    string
    Api    string
    Method string
}

// 发布者
type Owner struct {
    Name string
    Key  string
}
