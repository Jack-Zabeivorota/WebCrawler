package main

import (
	"os"

	"main/app"
	"main/logger"
)

func main() {
	logger.Init(os.Getenv("LOGS_DIR"))
	app.New().Run()
}
