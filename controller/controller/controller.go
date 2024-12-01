package controller

import (
	"crypto/sha256"
	"encoding/hex"
	"os"

	"github.com/labstack/echo/v4"

	"main/logger"
	"main/models"
	"main/msg_broker"
	"main/planner"
	"main/trafic_analizer"
)

type Controller struct {
	echo   *echo.Echo
	broker msg_broker.IMessageBroker
	log    *logger.Logger
	group  string
}

func New() *Controller {
	ctrl := &Controller{
		echo:   echo.New(),
		broker: msg_broker.NewKafkaBroker(os.Getenv("KAFKA_HOSTS")),
		log:    logger.Instance(),
		group:  "Controllers",
	}

	ctrl.setRoutes()
	return ctrl
}

func (ctrl *Controller) Run() {
	if os.Getenv("ENABLE_PLANNER") == "yes" {
		go planner.New(ctrl.broker, os.Getenv("PLANNER_RULES")).Run()
	} else {
		go trafic_analizer.New(ctrl.broker, ctrl.group).Run()
	}
	ctrl.echo.Start(":8001")
}

func (ctrl *Controller) checkPassword(password string) bool {
	hashPass := sha256.Sum256([]byte(password))
	return os.Getenv("PASSWORD_HASH") == hex.EncodeToString(hashPass[:])
}

// API

func (ctrl *Controller) setRoutes() {
	ctrl.echo.POST("/sign", ctrl.sign)
}

func (ctrl *Controller) isSign(value string) bool {
	signs := map[string]struct{}{
		models.Signals.Shutdown:    {},
		models.Signals.Kill:        {},
		models.Signals.ScaleUpdate: {},
	}

	_, ok := signs[value]
	return ok
}

func (ctrl *Controller) sign(c echo.Context) error {
	ctrl.log.Info("Start 'sign'")

	data := models.SignData{}
	err := c.Bind(&data)

	if err != nil {
		ctrl.log.Error("Incorrect params error: %v", err)
		return c.JSON(400, jsonMsg("Incorrect parameters"))
	}

	if !ctrl.checkPassword(data.Password) {
		ctrl.log.Error("Incorrect password: %s", data.Password)
		return c.JSON(401, jsonMsg("Incorrect password"))
	}

	if !ctrl.isSign(data.Sign) {
		ctrl.log.Error("Incorrect sign: %s", data.Sign)
		return c.JSON(400, jsonMsg("Incorrect signal"))
	}

	err = ctrl.broker.Send(models.SignalsTopic, &models.SignMsg{
		Sign:     data.Sign,
		Services: data.Services,
	})

	if err != nil {
		return internalError(c)
	}

	ctrl.log.Info("Completed 'sign'")
	return c.JSON(200, jsonMsg("Signal sent"))
}

// Helpers

func jsonMsg(message string) map[string]string {
	return map[string]string{"message": message}
}

func internalError(c echo.Context) error {
	return c.JSON(500, jsonMsg("Ops..."))
}
