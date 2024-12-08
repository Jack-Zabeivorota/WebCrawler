package main

import (
	"os"

	"main/app"
	"main/logger"
)

func main() {
	logger.Init(logger.LoggerConfig{
		ServiceName:  "Main-" + os.Getenv("ID"),
		Level:        "INFO",
		LogsDir:      os.Getenv("LOGS_DIR"),
		LogstashHost: os.Getenv("LOGSTASH_HOST"),
	})
	app.New().Run()
}
