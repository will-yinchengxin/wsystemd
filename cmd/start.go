package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
	"wsystemd/cmd/cluster"
	srv "wsystemd/cmd/http"
	"wsystemd/cmd/http/consts"
	"wsystemd/cmd/http/core"
	"wsystemd/cmd/log"
	"wsystemd/cmd/process"
	"wsystemd/cmd/task"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/common/version"
)

func init() {
	log.InitLog()
	level.Info(log.Logger).Log("msg", " built with "+runtime.Version())
}

func Run() int {
	var (
		serverPort = kingpin.Flag("server-port", "The vsw server port").Short('p').Default(consts.ServerPort).String()
		term       = make(chan os.Signal, 1)
		signals    = []os.Signal{syscall.SIGKILL, syscall.SIGSTOP,
			syscall.SIGINT, syscall.SIGQUIT, syscall.SIGILL,
			syscall.SIGABRT, syscall.SIGSYS, syscall.SIGTERM}
	)
	kingpin.Version(version.Print("wsystemd"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	signal.Notify(term, signals...)

	process.PManager = process.NewProcManager()
	level.Info(log.Logger).Log("msg", "NewProcManager Success")

	srv, cleanFun, err := srv.SetupRouter()
	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
	defer shutdownCancel()
	gracefulShutdown := make(chan struct{})
	if err != nil {
		level.Error(log.Logger).Log("msg", "Init Server Fail", "err", err)
		return 1
	}

	if val, ok := core.CoreConfig["singlemode"]; ok && !val.(bool) {
		wId := core.CoreConfig["workerid"].(string)
		cAddr := cluster.GetEtcdAddr()
		if wId == "" || cAddr == "" {
			panic("workerId/etcdAddr is empty")
		}
		manager, err := cluster.NewWorkerManager([]string{cAddr}, wId)
		if err != nil {
			level.Error(log.Logger).Log("msg", "Failed to create worker manager", "err", err)
			return 1
		}

		if err := manager.Register(shutdownCtx); err != nil {
			level.Error(log.Logger).Log("msg", "Failed to register worker", "err", err)
			return 1
		}
	}

	go func() {
		level.Info(log.Logger).Log("msg", "Start HTTP Server Success!!! ", "port", *serverPort)
		if err := http.ListenAndServe(":"+*serverPort, srv); err != nil {
			level.Error(log.Logger).Log("msg", "Error starting HTTP server", "err", err)
			time.Sleep(time.Second * 2)
			shutdownCancel()
		}
	}()

	go func() {
		task.CheckClientTask(shutdownCtx)
	}()

	go func() {
		<-term
		level.Info(log.Logger).Log("msg", "Received shutdown signal, starting graceful shutdown...")

		timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		shutdownCancel()
		CleanFunc(cleanFun)

		select {
		case <-timeoutCtx.Done():
			level.Error(log.Logger).Log("msg", "Graceful shutdown timed out")
		default:
		}

		close(gracefulShutdown)
	}()

	select {
	case <-shutdownCtx.Done():
		level.Info(log.Logger).Log("msg", "Server shutdown initiated")
		<-gracefulShutdown
		level.Info(log.Logger).Log("msg", "Graceful shutdown completed")
		return 0
	}
}

func CleanFunc(cleanFun []func()) {
	for _, clearFunc := range cleanFun {
		func() {
			defer func() {
				if r := recover(); r != nil {
					level.Error(log.Logger).Log("msg", "Panic in cleanup function", "error", r)
				}
			}()
			clearFunc()
		}()
	}
}
