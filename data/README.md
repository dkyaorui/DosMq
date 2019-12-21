# The File System in DosMq

## The data type in file

Every row data in file split by "\n"

topic.dosmq

```go
type Topic struct {
  ID                 [12]byte     `json:"id" binding:"-"`
  Name               [16]byte     `json:"topic_name" binding:"required"`
  ProcessMessageType [4]byte      `json:"topic_process_type" binding:"required"` // pull or push
  HashCode           [16]byte     `json:"hash_code" binding:"-"`
  Owner              Owner        `json:"topic_owner" binding:"-"`
  Subscribers        []Subscriber `json:"topic_subscribers" binding:"-"`
}
```

message.dosmq

```go
/*
Message collection_name: message

push the message to their topic's message queue.If the queue is full,
save the message into redis.

when consumer use the message,query in redis first and if no data in redis then
get message in queue.when the message is used then save it in mongodb.
*/
type Message struct {
  ID        [12]byte `json:"id" binding:"-"`
  TopicID   [12]byte `json:"topic_id" binding:"-"`
  HashCode  [16]byte `json:"hash_code" binding:"-"`
  Timestamp int64    `json:"timestamp" binding:"-"`
  Value     string   `json:"value" binding:"required"`
}
```

subscriber.dosmq

```go
// Subscriber collection_name: subscriber
type Subscriber struct {
  ID         [12]byte `json:"id" binding:"-"`
  ServerName [16]byte `json:"sub_server_name" binding:"required"`
  HashCode   [16]byte `json:"hash_code" binding:"-"`
  TopicID    [12]byte `json:"sub_topic_id" binding:"-"`
  Method     [4]byte  `json:"sub_method" binding:"required"`
  Host       string   `json:"sub_host" binding:"required"`
  API        string   `json:"sub_api" binding:"required"`
  Key        string   `json:"sub_key" binding:"required"`
}
```

owner.dosmq

```go
// Owner collection_name: owner
type Owner struct {
  ID          [12]byte `json:"id" binding:"-"`
  TopicID     [12]byte `json:"owner_topic_id" binding:"-"`
  HashCode    [16]byte `json:"hash_code" binding:"-"`
  ServerName  [16]byte `json:"owner_server_name" binding:"required"`
  Key         string `json:"owner_key" binding:"required"` // Owner's key
  Host        string `json:"owner_host" binding:"required"`
}
```

## The logic to deal with file

> The message.dosmq's max size is 0.5GB, and if the file size of more than 0.5GB, we will use new file to save data.
>
> we will not delete the data file initiativly, so wo find data by index.
