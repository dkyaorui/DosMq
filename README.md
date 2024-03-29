# DosMq


Message queue middleware based on `Golang` that can be deployed  as an independent project.

[中文](README_ZH.md)

**Version**: ![Dosmq](https://img.shields.io/badge/Dosmq-1.0.0-blue)

**Dependency**:

![go-1.13.4](https://img.shields.io/badge/go-1.13.4-green)
![Mongo-latest](https://img.shields.io/badge/Mongo-latest-red)
![Redis-latest](https://img.shields.io/badge/Redis-latest-red)

##  Deployment

We need a linux server with `redis` and `mongodb`, and install `go-1.13.4` first.

If you set password for database, you should set variable in environment:

- `REDIS_PASSWORD`
- `MONGO_PASSWORD`

### Deploy step

- Install dependency

```bash
$ go mod download
```

- Edit the config file: `/config/config.yaml`

- Build the project

```bash
$ go build -o dosmq
```

- Run the Project

```bash
$ ./dosmq
```


## API Document

`/api/consumer/get`

Note：if `topic_process_type` is`get`，consumer can get one message.

Method：`POST`

Headers：`applications/json`

Parameter：

```json
{	
    "sub_server_name": "subscriber_name",
    "sub_host": "subscriber_name",
    "sub_key": "subscriber_secret",
    "sub_api": "/...",
    "sub_method": "get|post|...|none"
}
```

return：
```json
{
    "Code": "Http.StatusCode",
    "Data": "data | msg"
}
```

`/api/producer/send`

Note：An api for producer send one message.

Method：`POST`

Headers：`applications/json`

Parameter：

```json
{	
    "auth_key": "auth_key",
    "owner_server_name": "producer_name",
    "owner_key": "producer_secret_key",
    "owner_host": "producer_host",
    "value": "massage"
}
```

return：

```json
{
    "Code": "Http.StatusCode",
    "Data": "data | msg"
}
```

`/api/topic/sign_up`

Note：An api for producer sign up a topic.

Method：`POST`

Headers：`applications/json`

Parameter：

```json
{	
    "auth_key": "auth_key",
    "owner_server_name": "producer_name",
    "owner_key": "producer_secret",
    "owner_host": "producer_host",
    "topic_name": "topic_name",
    "topic_process_type": "push | pull"
}
```

return：

```json
{
    "Code": "Http.StatusCode",
    "Data": "data | msg"
}
```

`/api/topic/del`

Note：An api for producer del one topic.

Method：`POST`

Headers：`applications/json`

Parameter：

```json
{	
    "auth_key": "auth_key",
    "owner_server_name": "producer_name",
    "owner_key": "producer_secret",
    "owner_host": "producer_host",
    "topic_name": "topic_name",
    "topic_process_type": "push | pull"
}
```

return：

```json
{
    "Code": "Http.StatusCode",
    "Data": "data | msg"
}
```

`/api/topic/subscribe`

Note：An api for consumer subscribe one topic.

Method：`POST`

Headers：`applications/json`

Parameter：

```json
{	
    "auth_key": "auth_key",
    "owner_server_name": "producer_name",
    "owner_key": "producer_secret",
    "owner_host": "producer_host",
    "sub_server_name": "subscriber_name",
    "sub_host": "subscriber_name",
    "sub_key": "subscriber_secret",
    "sub_api": "/...",
    "sub_method": "get|post|...|none"
}
```

return：

```json
{
    "Code": "Http.StatusCode",
    "Data": "data | msg"
}
```

`/api/topic/cancel_subscribe`

Note：An api for consumer unsubscribe one topic.

Method：`POST`

Headers：`applications/json`

Parameter：

```json
{	 
    "auth_key": "auth_key",
    "owner_server_name": "producer_name",
    "owner_key": "producer_secret",
    "owner_host": "producer_host",
    "sub_server_name": "subscriber_name",
    "sub_host": "subscriber_name",
    "sub_key": "subscriber_secret",
    "sub_api": "/...",
    "sub_method": "get|post|...|none"
}
```

return：

```json
{
    "Code": "Http.StatusCode",
    "Data": "data | msg"
}
```