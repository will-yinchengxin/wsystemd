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
		srvc       = make(chan struct{})
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
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		CleanFunc(cleanFun)
		close(term)
	}()
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

		if err := manager.Register(ctx); err != nil {
			level.Error(log.Logger).Log("msg", "Failed to register worker", "err", err)
			return 1
		}
	}

	go func() {
		level.Info(log.Logger).Log("msg", "Start HTTP Server Success!!! ", "port", *serverPort)
		if err := http.ListenAndServe(":"+*serverPort, srv); err != nil {
			level.Error(log.Logger).Log("msg", "Error starting HTTP server", "err", err)
			close(srvc)
		}
	}()

	go func() {
		task.CheckClientTask(ctx)
	}()

	go func() {
		<-term
		level.Info(log.Logger).Log("msg", "Received shutdown signal, starting graceful shutdown...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		done := make(chan struct{})
		go func() {
			cancel()
			CleanFunc(cleanFun)
			close(done)
		}()

		select {
		case <-done:
			level.Info(log.Logger).Log("msg", "Graceful shutdown completed")
		case <-shutdownCtx.Done():
			level.Error(log.Logger).Log("msg", "Graceful shutdown timed out")
		}

		os.Exit(0)
	}()

	for {
		select {
		case <-term:
			level.Info(log.Logger).Log("msg", "Received SIGTERM, Exiting Gracefully...")
			return 0
		case <-srvc:
			level.Error(log.Logger).Log("msg", "Server Exist With Error, Check it...")
			return 1
		}
	}
}

func CleanFunc(cleanFun []func()) {
	if len(cleanFun) > 0 {
		for _, clearFunc := range cleanFun {
			clearFunc()
		}
	}
}
