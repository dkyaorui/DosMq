package db

import (
    "fmt"
    "github.com/gomodule/redigo/redis"
    "github.com/pkg/errors"
    log "github.com/sirupsen/logrus"
    "github.com/spf13/viper"
    "strconv"
    "time"
)

var RedisClient *redis.Pool

func RedisInit() {
    redisConfig := viper.GetStringMap("redis")
    redisHost := redisConfig["host"].(string) + ":" + strconv.Itoa(redisConfig["port"].(int))
    redisPassword := viper.GetString("redisPassword")
    RedisClient = &redis.Pool{
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
