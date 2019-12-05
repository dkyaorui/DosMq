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

var RDbClient RDbPool

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

func (p *RDbPool) getKey(key string) string {
    return "mq_" + key
}

func (p *RDbPool) Expire(key string, expire int64) error {
    _, err := redis.Bool(p.Do("EXPIRE", p.getKey(key), expire))
    return err
}

func (p *RDbPool) TTL(key string) (ttl int64, err error) {
    return redis.Int64(p.Do("TTL", p.getKey(key)))
}

func (p *RDbPool) Exists(key string) (bool, error) {
    return redis.Bool(p.Do("EXISTS", p.getKey(key)))
}

func (p *RDbPool) Set(key string, value interface{}, expire int64) (err error) {
    val, err := p.encode(value)
    if err != nil {
        return err
    }
    if expire > 0 {
        _, err := p.Do("SETEX", p.getKey(key), expire, val)
        return err
    }
    _, err = p.Do("SET", p.getKey(key), value)
    return err
}

func (p *RDbPool) Get(key string) (reply interface{}, err error) {
    reply, err = p.Do("GET", p.getKey(key))
    return reply, err
}

func (p *RDbPool) Delete(key string, value interface{}) (err error) {
    _, err = p.Do("DEL", p.getKey(key), value)
    return err
}

func (p *RDbPool) LPush(key string, value interface{}) (err error) {
    val, err := p.encode(value)
    if err != nil {
        return err
    }
    _, err = p.Do("LPUSH", p.getKey(key), val)
    return err
}

func (p *RDbPool) RPush(key string, value interface{}) (err error) {
    val, err := p.encode(value)
    if err != nil {
        return err
    }
    _, err = p.Do("RPUSH", p.getKey(key), val)
    return err
}

func (p *RDbPool) LPop(key string) (reply interface{}, err error) {
    reply, err = p.Do("LPOP", p.getKey(key))
    return reply, err
}

func (p *RDbPool) LPopObject(key string, value interface{}) (err error) {
    reply, err := p.LPop(key)
    return p.decode(reply, err, value)
}

func (p *RDbPool) RPop(key string) (reply interface{}, err error) {
    reply, err = p.Do("RPOP", p.getKey(key))
    return reply, err
}

func (p *RDbPool) RPopObject(key string, value interface{}) (err error) {
    reply, err := p.RPop(key)
    return p.decode(reply, err, value)
}

func (p *RDbPool) LLen(key string) (length int64, err error) {
    reply, err := p.Do("LLEN", p.getKey(key))
    return reply.(int64), err
}
