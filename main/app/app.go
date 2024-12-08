package app

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"main/cache"
	"main/database"
	dbm "main/database/models"
	"main/logger"
	"main/models"
	"main/msg_broker"
)

type App struct {
	id      int
	isPause bool
	echo    *echo.Echo
	broker  msg_broker.IMessageBroker
	cache   cache.ICache
	db      database.IDataBase
	log     *logger.Logger
	group   string
}

func New() *App {
	log := logger.Instance()

	db := database.NewPostgreSQL(os.Getenv("POSTGRES_CONN"))
	err := db.Init()

	if err != nil {
		log.Fatal("Initialization DB error: %v", err)
	}

	id, err := strconv.Atoi(os.Getenv("ID"))

	if err != nil {
		log.Fatal("ID is not specified")
	}

	app := &App{
		id:     id,
		echo:   echo.New(),
		broker: msg_broker.NewKafkaBroker(os.Getenv("KAFKA_HOSTS")),
		cache:  cache.NewRedisCache(os.Getenv("REDIS_HOST")),
		db:     db,
		log:    log,
		group:  "Mains",
	}
	app.setRoutes()

	return app
}

func (app *App) signsHandler(data []byte) {
	msg := &models.SignMsg{}
	err := json.Unmarshal(data, &msg)

	if err != nil {
		app.log.Error("Incorrect signal error: %v", err)
		return
	}

	if id, ok := msg.Services["main"]; ok {

		if msg.Sign == models.Signals.ScaleUpdate {
			if app.id <= id && app.isPause {
				app.isPause = false
				app.log.Info("Start with signal 'ScaleUpdate'")

			} else if app.id > id && !app.isPause {
				app.isPause = true
				app.log.Info("Pause with signal 'ScaleUpdate'")
			}
			return
		}

		if (msg.Sign == models.Signals.Shutdown || msg.Sign == models.Signals.Kill) && id == app.id {
			app.log.Info("Closed with signal '%s'", msg.Sign)
			os.Exit(0)
		}

	}
}

func (app *App) Run() {
	ctx := context.Background()
	go app.broker.Receiving(ctx, models.SignalsTopic, app.group, app.signsHandler)

	app.echo.Start(":8000")
}

// APIs

func (app *App) setRoutes() {
	app.echo.POST("/request", app.addRequest)
	app.echo.GET("/request", app.getRequest)
	app.echo.DELETE("/request", app.deleteRequest)
	app.echo.Static("/static", "static")
	app.echo.GET("/", app.index)
	app.echo.Use(app.pauseMiddleware)
}

func (app *App) index(c echo.Context) error {
	return c.File("static/index.html")
}

func (app *App) addRequest(c echo.Context) error {
	app.log.Info("Start 'add request'")

	// parsing request body

	urlData := models.URLData{}
	err := c.Bind(&urlData)

	var validWords []string

	if err == nil {
		urlData.URL = strings.TrimRight(urlData.URL, "/")

		validWords = make([]string, 0, len(urlData.Words))

		for _, word := range urlData.Words {
			if word != "" {
				validWords = append(validWords, strings.ToLower(word))
			}
		}
	}

	if err != nil || urlData.URL == "" || len(validWords) == 0 {
		app.log.Debug(
			"Incorrect parameters: wordsLen: %d, URL: %s, error: %v",
			len(urlData.Words), urlData.URL, err,
		)
		return c.JSON(400, jsonMsg("Incorrect parameters"))
	}

	// additing request data in DB and cache

	requestData := &models.RequestData{
		Words:          urlData.Words,
		SameDomainOnly: urlData.SameDomainOnly,
	}

	requestID, err := app.db.AddRequest(requestData, urlData.URL)

	if err != nil {
		return internalError(c)
	}

	requestData.Words = validWords

	err = app.cache.SetRequestData(requestID, requestData)

	if err != nil {
		app.db.DeleteRequest(requestID)
		return internalError(c)
	}

	err = app.cache.SetURLToAllURLs(requestID, urlData.URL)

	if err != nil {
		app.db.DeleteRequest(requestID)
		return internalError(c)
	}

	// sending message to workers

	msg := models.FindWordsMsg{
		RequestID: requestID,
		URL:       urlData.URL,
	}

	err = app.broker.Send(models.FindWordsTopic, msg)

	if err != nil {
		return internalError(c)
	}

	app.log.Info("Completed 'add request'")

	return c.JSON(200, map[string]any{
		"message":    "Your request accesed and in process",
		"request_id": requestID,
	})
}

func (app *App) parseRequestID(c echo.Context) int64 {
	param := c.QueryParam("ID")

	if param == "" {
		app.log.Debug("Do not entered requestID")
		c.JSON(400, jsonMsg("Do not entered requestID"))
		return -1
	}

	requestID, err := strconv.ParseInt(param, 10, 64)

	if err != nil {
		app.log.Debug("Wrong parameter: requestID is '%s', must be integer", param)
		c.JSON(400, jsonMsg("Entered ID is not number"))
		return -1
	}

	return requestID
}

func (App) createRequestResultResponse(request *dbm.Request, urls []dbm.URL) *models.RequestResult {
	urlResults := make([]models.URLResult, len(urls))

	urlStatusDict := map[int]string{
		0: "Success",
		1: "Fail",
		2: "Unreaded",
	}

	for i, url := range urls {
		urlResults[i] = models.URLResult{
			URL:         url.URL,
			Status:      urlStatusDict[url.Status],
			FindedWords: strings.Split(url.FindedWords, ","),
		}
	}

	return &models.RequestResult{
		ID:             request.ID,
		StartURL:       request.StartURL,
		Words:          strings.Split(request.Words, ","),
		SameDomainOnly: request.SameDomainOnly,
		URLs:           urlResults,
	}
}

func (app *App) getRequest(c echo.Context) error {
	app.log.Info("Start 'get request'")

	// parsing and validation parameter

	requestID := app.parseRequestID(c)

	if requestID == -1 {
		return nil
	}

	// geting request from db and valid

	request, err := app.db.GetRequest(requestID)

	if err != nil {
		return internalError(c)
	}

	if request == nil {
		app.log.Info("Completed 'get request': not found request %d", requestID)
		return c.JSON(404, jsonMsg("Request not found"))
	}

	if !request.IsDone {
		app.log.Info("Completed 'get request': request %d in progress", requestID)
		return c.JSON(200, jsonMsg("Your request in progress"))
	}

	// geting urls from request and sending response

	urls, err := app.db.GetURLsFromRequest(requestID)

	if err != nil {
		return internalError(c)
	}

	results := app.createRequestResultResponse(request, urls)

	app.log.Info("Completed 'get request'")
	return c.JSON(200, results)
}

func (app *App) deleteRequest(c echo.Context) error {
	app.log.Info("Start 'delete request'")

	// parsing and validation parameter

	requestID := app.parseRequestID(c)

	if requestID == -1 {
		return nil
	}

	// deleting request and urls results

	isDeleted, err := app.db.DeleteRequestAndURLs(requestID)

	if err != nil {
		return internalError(c)
	}

	if !isDeleted {
		app.log.Debug("Did not deleted request and URLs from DB (request not found probably)")
		return c.JSON(404, jsonMsg("Request not found"))
	}

	err = app.cache.ClearRequest(requestID)

	if err != nil {
		return internalError(c)
	}

	app.log.Info("Completed 'delete request'")
	return c.JSON(200, jsonMsg("Request has been deleted"))
}

// Middlewares

func (app *App) pauseMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if app.isPause {
			return c.JSON(503, jsonMsg("Unworking service"))
		}
		return next(c)
	}
}

// Helpers

func jsonMsg(message string) map[string]string {
	return map[string]string{"message": message}
}

func internalError(c echo.Context) error {
	return c.JSON(500, jsonMsg("Ops..."))
}
