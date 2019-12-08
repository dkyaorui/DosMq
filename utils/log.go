package utils

import (
    rotatelogs "github.com/lestrrat-go/file-rotatelogs"
    "github.com/rifflock/lfshook"
    log "github.com/sirupsen/logrus"
    "github.com/spf13/viper"
    "os"
    "path"
    "time"
)

func LogInit() {
    logConfig := viper.GetStringMap("log")
    baseLogPath := path.Join(logConfig["path"].(string), logConfig["name"].(string))

    logFile, err := os.OpenFile(baseLogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)

    if err != nil {
        panic(err)
    }
    log.SetOutput(logFile)
    log.SetLevel(log.DebugLevel)

    writer, err := rotatelogs.New(
        baseLogPath+".%Y%m%d%h%M",
        rotatelogs.WithLinkName(baseLogPath),
        rotatelogs.WithRotationCount(5),
        rotatelogs.WithRotationTime(time.Hour),
    )

    if err != nil {
        log.Panic(err)
    }

    lfHook := lfshook.NewHook(lfshook.WriterMap{
        log.DebugLevel: writer,
        log.InfoLevel:  writer,
        log.WarnLevel:  writer,
        log.ErrorLevel: writer,
        log.FatalLevel: writer,
        log.PanicLevel: writer,
    }, &log.TextFormatter{})

    log.AddHook(lfHook)

    log.Info("log file load success……")
}

