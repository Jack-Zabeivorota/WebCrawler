package main

import (
	"os"

	"main/aggregator"
	"main/logger"
)

func main() {
	logger.Init(logger.LoggerConfig{
		ServiceName:  "Aggregator-" + os.Getenv("ID"),
		Level:        "INFO",
		LogsDir:      os.Getenv("LOGS_DIR"),
		LogstashHost: os.Getenv("LOGSTASH_HOST"),
	})
	aggregator.New().Run()
}
