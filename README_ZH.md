# DosMq

[English](README.md)

基于`golang`开发的消息队列中间件，可作为独立项目部署。



**版本**：![Dosmq](https://img.shields.io/badge/Dosmq-1.0.0-blue)

**环境依赖**：

![go-1.13.4](https://img.shields.io/badge/go-1.13.4-green)
![Mongo-latest](https://img.shields.io/badge/Mongo-latest-red)
![Redis-latest](https://img.shields.io/badge/Redis-latest-red)

## 部署方式

准备好`redis`和`mongodb`数据库，安装`go-1.13.4`环境

如果数据库有密码需要在环境变量中设置以下字段：
- `REDIS_PASSWORD`
- `MONGO_PASSWORD`

### 部署步骤

- 安装依赖

```bash
$ go mod download
```

- 修改配置文件 `/config/config.yaml`

- 编译项目

```bash
$ go build -o dosmq
```

- 运行项目

```bash
$ ./dosmq
```


## API 文档

`/api/consumer/get`

说明：`topic_process_type`为`get`时，消费者可调用的接口，用于获取一条信息。

方法：`POST`

请求头：`applications/json`

参数：

```json
{	
    "sub_server_name": "subscriber_name",
    "sub_host": "subscriber_name",
    "sub_key": "subscriber_secret",
    "sub_api": "/...",
    "sub_method": "get|post|...|none"
}
```

返回值：
```json
{
    "Code": "Http.StatusCode",
    "Data": "data | msg"
}
```

`/api/producer/send`

说明：消息发送方发送消息的接口

方法：`POST`

请求头：`applications/json`

参数：

```json
{	
    "auth_key": "auth_key",
    "owner_server_name": "producer_name",
    "owner_key": "producer_secret_key",
    "owner_host": "producer_host",
    "value": "massage"
}
```

返回值：

```json
{
    "Code": "Http.StatusCode",
    "Data": "data | msg"
}
```

`/api/topic/sign_up`

说明：消息发送方注册主题的接口

方法：`POST`

请求头：`applications/json`

参数：

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

返回值：

```json
{
    "Code": "Http.StatusCode",
    "Data": "data | msg"
}
```

`/api/topic/del`

说明：消息发送方删除主题的接口

方法：`POST`

请求头：`applications/json`

参数：

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

返回值：

```json
{
    "Code": "Http.StatusCode",
    "Data": "data | msg"
}
```

`/api/topic/subscribe`

说明：消息接收方订阅主题的接口

方法：`POST`

请求头：`applications/json`

参数：

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

返回值：

```json
{
    "Code": "Http.StatusCode",
    "Data": "data | msg"
}
```

`/api/topic/cancel_subscribe`

说明：消息接收方取消订阅主题的接口

方法：`POST`

请求头：`applications/json`

参数：

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

返回值：

```json
{
    "Code": "Http.StatusCode",
    "Data": "data | msg"
}
```