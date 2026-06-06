package logger

import (
    "os"
    "time"
    
    "github.com/sirupsen/logrus"
)

var log = logrus.New()

func Init(level string, format string) {
    // Уровень логирования
    lvl, err := logrus.ParseLevel(level)
    if err != nil {
        lvl = logrus.InfoLevel
    }
    log.SetLevel(lvl)
    
    // Формат вывода
    if format == "json" {
        log.SetFormatter(&logrus.JSONFormatter{
            TimestampFormat: time.RFC3339,
        })
    } else {
        log.SetFormatter(&logrus.TextFormatter{
            FullTimestamp:   true,
            TimestampFormat: time.RFC3339,
        })
    }
    
    log.SetOutput(os.Stdout)
}

func Debug(args ...interface{}) {
    log.Debug(args...)
}

func Info(args ...interface{}) {
    log.Info(args...)
}

func Warn(args ...interface{}) {
    log.Warn(args...)
}

func Error(args ...interface{}) {
    log.Error(args...)
}

func Fatal(args ...interface{}) {
    log.Fatal(args...)
}

func WithField(key string, value interface{}) *logrus.Entry {
    return log.WithField(key, value)
}

func WithFields(fields logrus.Fields) *logrus.Entry {
    return log.WithFields(fields)
}