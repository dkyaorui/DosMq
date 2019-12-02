package redis

import (
    "encoding/json"
    "fmt"
    "github.com/gomodule/redigo/redis"
    "github.com/pkg/errors"
    log "github.com/sirupsen/logrus"
    "github.com/spf13/viper"
    "strconv"
    "time"
)

var RDbClient *RDbPool

type RDbPool struct {
    pool *redis.Pool
}

func Init() {
    redisConfig := viper.GetStringMap("redis")
    redisHost := redisConfig["host"].(string) + ":" + strconv.Itoa(redisConfig["port"].(int))
    redisPassword := viper.GetString("redisPassword")
    RDbClient.pool = &redis.Pool{
        MaxIdle:     redisConfig["maxidle"].(int),
        MaxActive:   redisConfig["maxactive"].(int),
        IdleTimeout: 3 * time.Minute,
        Dial: func() (conn redis.Conn, err error) {
            conn, err = redis.Dial(
                "tcp",
                redisHost,
                redis.DialPassword(redisPassword))
            if err != nil {
                err = errors.Wrap(err, "Redis connect error")
                fmt.Println(err)
                log.Panicf("%+v", err)
            }
            _, _ = conn.Do("SELECT", redisConfig["db"])
            return conn, err
        },
    }
}

func (p *RDbPool) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
    coon := p.pool.Get()
    defer coon.Close()
    return coon.Do(commandName, args...)
}

func (p *RDbPool) encode(val interface{}) (interface{}, error) {
    var value interface{}
    switch v := val.(type) {
    case string, int, uint, int8, int16, int32, int64, float32, float64, bool:
        value = v
    default:
        b, err := json.Marshal(v)
        if err != nil {
            return nil, err
        }
        value = string(b)
    }
    return value, nil
}

func (p *RDbPool) decode(reply interface{}, err error, val interface{}) error {
    str, err := redis.String(reply, err)
    if err != nil {
        return err
    }
    return json.Unmarshal([]byte(str), val)
}

func (p *RDbPool) Insert(key string, value interface{}) (reply interface{}, err error) {
    reply, err = p.Do("SET", key, value)
    return reply, err
}

func (p *RDbPool) Get(key string) (reply interface{}, err error) {
    reply, err = p.Do("GET", key)
    return reply, err
}

func (p *RDbPool) Update(key string, value interface{}) (reply interface{}, err error) {
    reply, err = p.Do("SET", key, value)
    return reply, err
}

func (p *RDbPool) Delete(key string, value interface{}) (reply interface{}, err error) {
    reply, err = p.Do("DEL", key, value)
    return reply, err
}

func (p *RDbPool) LPush(key string, value interface{}) (reply interface{}, err error) {
    reply, err = p.Do("LPUSH", key, value)
    return reply, err
}

func (p *RDbPool) RPush(key string, value interface{}) (reply interface{}, err error) {
    reply, err = p.Do("RPUSH", key, value)
    return reply, err
}

func (p *RDbPool) RPop(key string) (reply interface{}, err error) {
    reply, err = p.Do("RPOP", key)
    return reply, err
}
