package tools

import (
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"main/logger"
)

func RetryCycle(operation func() error, errMessage string, interceptAllErrors bool) error {
	period, trys := time.Minute, 0
	log := logger.Instance()

	for {
		err := operation()

		if interceptAllErrors {
			if err == nil {
				return nil
			}
		} else {
			isConnErr := false

			if nErr, ok := err.(net.Error); ok && nErr.Timeout() {
				isConnErr = true
			}

			if _, ok := err.(*net.OpError); ok {
				isConnErr = true
			}

			if _, ok := err.(*pgconn.ConnectError); ok {
				isConnErr = true
			}

			if !isConnErr {
				return err
			}
		}

		log.Error("%s: %v", errMessage, err)

		if trys == 5 {
			trys = 0
			period *= 2
		}

		trys++
		time.Sleep(period)
	}
}

func Select[Tin any, Tout any](arr []Tin, selector func(Tin) Tout) []Tout {
	result := make([]Tout, len(arr))

	for i := 0; i < len(arr); i++ {
		result[i] = selector(arr[i])
	}

	return result
}
