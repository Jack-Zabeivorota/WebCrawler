package main

import (
	"os"

	"main/logger"
	"main/worker"
)

func main() {
	logger.Init(logger.LoggerConfig{
		ServiceName:  "Worker-" + os.Getenv("ID"),
		Level:        "INFO",
		LogsDir:      os.Getenv("LOGS_DIR"),
		LogstashHost: os.Getenv("LOGSTASH_HOST"),
	})
	worker.New().Run()
}
