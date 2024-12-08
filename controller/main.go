package main

import (
	"os"

	"main/controller"
	"main/logger"
)

func main() {
	logger.Init(logger.LoggerConfig{
		ServiceName:  "Controller-" + os.Getenv("ID"),
		Level:        "INFO",
		LogsDir:      os.Getenv("LOGS_DIR"),
		LogstashHost: os.Getenv("LOGSTASH_HOST"),
	})
	controller.New().Run()
}
