package mongo

import (
    "go.mongodb.org/mongo-driver/bson/primitive"
    "time"
)

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
type Topic struct {
    Name               string       `bson:"name" json:"topic_name"`
    Subscribers        []Subscriber `bson:"subscribers" json:"topic_subscribers"`
    Owner              Owner        `bson:"owner" json:"topic_owner"`
    ProcessMessageType string       `bson:"process_message_type" json:"topic_process_message_type"` // pull or push
}

// 消息
type Message struct {
    TopicId   primitive.ObjectID `bson:"topic_id"`
    Value     []byte             `bson:"value"`
    Timestamp int64              `bson:"create_time"`
}

// 订阅者
type Subscriber struct {
    ServerName string             `bson:"sub_server_name"`
    Host       string             `bson:"sub_Host"`
    Key        string             `bson:"sub_key"`
    Api        string             `bson:"sub_api"`
    Method     string             `bson:"sub_method"`
    Protocol   string             `bson:"sub_protocol"`
    TopicId    primitive.ObjectID `bson:"topic_id"`
}

// 发布者
type Owner struct {
    ServerName string             `bson:"owner_server_name" json:"owner_server_name"`
    Key        string             `bson:"owner_key" json:"owner_key"` // 发布者上传key
    Host       string             `bson:"owner_host" json:"owner_host"`
    TopicID    primitive.ObjectID `bson:"topic_id" json:"owner_topic_id"`
}
