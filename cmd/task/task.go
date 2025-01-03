package task

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/log/level"
	"time"
	"wsystemd/cmd/http/core"
	"wsystemd/cmd/http/service"
	"wsystemd/cmd/log"
)

func CheckClientTask(ctx context.Context) {
	timerTicker, ok := core.CoreConfig["scheduletimeticker"].(int64)
	if ok && timerTicker > 0 {
		timerTicker = core.CoreConfig["scheduletimeticker"].(int64)
	} else {
		timerTicker = 20
	}
	for {
		select {
		case <-ctx.Done():
			level.Info(log.Logger).Log("msg", "Received SIGTERM, CheckClientAlive Task Exit")
			return
		case <-time.Tick(time.Second * time.Duration(timerTicker)):
			err := service.CheckClientAlive()
			if err != nil {
				fmt.Println("")
				level.Error(log.Logger).Log("Err", fmt.Sprintf("Schedule CheckClientAlive Err: %s"))
				fmt.Println("")
				time.Sleep(time.Second * 10)
			}
		}
	}
}
