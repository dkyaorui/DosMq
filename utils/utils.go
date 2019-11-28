package utils

import (
    "github.com/pkg/errors"
    "github.com/spf13/viper"
)

// United return data format
type RequestResult struct {
    Code int // http's status code
    Data interface{}
}

func CheckRegisterTopicKey(key string) (err error) {
    if key != viper.Get("topic.key") {
        return errors.Wrap(err, "the key with register topic is wrong")
    }
    return nil
}
