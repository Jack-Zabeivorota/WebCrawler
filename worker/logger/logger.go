package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	stdlog "log"
	"net"
	"net/http"
	"os"
	"time"
)

type LoggerConfig struct {
	ServiceName  string
	Level        string
	LogsDir      string
	LogstashHost string
}

type LogMsg struct {
	Level   string
	Message string
}

type Logger struct {
	queue    chan *LogMsg
	levels   map[string]int
	config   LoggerConfig
	logsFile *os.File
}

var instance *Logger

func Instance() *Logger {
	if instance == nil {
		stdlog.Fatalln("Logs in not initialized")
	}
	return instance
}

func openLogsFile(config LoggerConfig) *os.File {
	if config.LogsDir == "" {
		return nil
	}

	err := os.MkdirAll(config.LogsDir, os.ModePerm)

	if err != nil {
		stdlog.Fatalln("Create directories on '", config.LogsDir, "' for logs error: ", err)
	}

	now := time.Now()

	path := fmt.Sprintf(
		"%s%s-%d.%d.%d.log",
		config.LogsDir, config.ServiceName, now.Year(), now.Month(), now.Day(),
	)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		stdlog.Fatalln("Open file on '", path, "' for logs error: ", err)
	}

	return file
}

func Init(config LoggerConfig) {
	if instance != nil {
		stdlog.Fatalln("Logger is already initialized")
	}

	if config.ServiceName == "" {
		stdlog.Fatalln("Service name not specified")
	}

	levels := map[string]int{
		"FATAL": 3,
		"ERROR": 2,
		"INFO":  1,
		"DEBUG": 0,
	}

	if _, ok := levels[config.Level]; !ok {
		stdlog.Fatalln("Incorrect level ", config.Level)
	}

	instance = &Logger{
		queue:    make(chan *LogMsg, 10),
		levels:   levels,
		config:   config,
		logsFile: openLogsFile(config),
	}

	go instance.logHandler()
}

// Hooks

func (log *Logger) sendlogToTerminal(entity map[string]any) {
	stdlog.Println(entity["timestamp"], " ", entity["status"], ": ", entity["message"])
}

func (log *Logger) sendlogToLogstash(entity map[string]any) {
	if log.config.LogstashHost == "" {
		return
	}

	data, err := json.Marshal(entity)

	if err != nil {
		stdlog.Fatalln("Parsing log to JSON error: ", err)
	}

	req, err := http.NewRequest("POST", log.config.LogstashHost, bytes.NewBuffer(data))

	if err != nil {
		stdlog.Fatalln("Create request for Logstash error: ", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	var resp *http.Response

	for i := 0; i < 3; i++ {
		resp, err = client.Do(req)

		if nErr, ok := err.(net.Error); !(ok && nErr.Timeout()) {
			break
		}

		if i == 2 {
			stdlog.Println("Sending log to Logstash error: ", err)
			return
		}
	}

	if err != nil {
		stdlog.Println("Sending log to Logstash error: ", err)
		return
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		stdlog.Println("Sending log to Logstash not OK: ", resp.Status)
	}
}

func (log *Logger) sendlogToFile(entity map[string]any) {
	if log.logsFile == nil {
		return
	}

	data, err := json.Marshal(entity)

	if err != nil {
		stdlog.Fatalln("Parsing log to JSON error: ", err)
	}

	_, err = log.logsFile.Write(data)

	if err != nil {
		stdlog.Println("Write log to file error: ", err)
	}
}

func (log *Logger) logHandler() {
	for {
		logMsg := <-log.queue

		if log.levels[logMsg.Level] < log.levels[log.config.Level] {
			continue
		}

		entity := map[string]any{
			"level":   logMsg.Level,
			"service": log.config.ServiceName,
			"message": logMsg.Message,
		}

		log.sendlogToTerminal(entity)
		log.sendlogToLogstash(entity)
		log.sendlogToFile(entity)

		if logMsg.Level == "FATAL" {
			if log.logsFile != nil {
				log.logsFile.Close()
			}
			os.Exit(1)
		}
	}
}

// Public functions

func (log *Logger) Debug(message string, args ...any) {
	log.queue <- &LogMsg{"DEBUG", fmt.Sprintf(message, args...)}
}

func (log *Logger) Info(message string, args ...any) {
	log.queue <- &LogMsg{"INFO", fmt.Sprintf(message, args...)}
}

func (log *Logger) Error(message string, args ...any) {
	log.queue <- &LogMsg{"ERROR", fmt.Sprintf(message, args...)}
}

func (log *Logger) Fatal(message string, args ...any) {
	log.queue <- &LogMsg{"FATAL", fmt.Sprintf(message, args...)}
}
