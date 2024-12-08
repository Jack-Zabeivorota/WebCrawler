package logger

import (
	"fmt"
	stdlog "log"
	"os"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	cmdLog  *logrus.Logger
	fileLog *logrus.Logger
	fields  logrus.Fields
}

var instance *Logger

func Instance() *Logger {
	if instance == nil {
		stdlog.Fatalln("Logs in not initialized")
	}
	return instance
}

func Init(logsDir string) {
	os.MkdirAll(logsDir, os.ModePerm)
	file, err := os.OpenFile(logsDir+"controller.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		stdlog.Fatalln("Open file for logs error: ", err)
	}

	level := logrus.DebugLevel

	fileLog := logrus.New()
	fileLog.SetOutput(file)
	fileLog.SetFormatter(&logrus.JSONFormatter{})
	fileLog.SetLevel(level)

	cmdLog := logrus.New()
	cmdLog.SetOutput(os.Stdout)
	cmdLog.SetFormatter(&logrus.TextFormatter{})
	cmdLog.SetLevel(level)

	instance = &Logger{
		cmdLog:  cmdLog,
		fileLog: fileLog,
		fields: logrus.Fields{
			"service": "Main",
		},
	}
}

func (log *Logger) Debug(message string, args ...any) {
	msg := fmt.Sprintf(message, args...)
	log.cmdLog.Debugln(msg)
	log.fileLog.WithFields(log.fields).Debugln(msg)
}

func (log *Logger) Info(message string, args ...any) {
	msg := fmt.Sprintf(message, args...)
	log.cmdLog.Infoln(msg)
	log.fileLog.WithFields(log.fields).Infoln(msg)
}

func (log *Logger) Error(message string, args ...any) {
	msg := fmt.Sprintf(message, args...)
	log.cmdLog.Errorln(msg)
	log.fileLog.WithFields(log.fields).Errorln(msg)
}

func (log *Logger) Fatal(message string, args ...any) {
	msg := fmt.Sprintf(message, args...)
	log.cmdLog.Fatalln(msg)
	log.fileLog.WithFields(log.fields).Fatalln(msg)
}
