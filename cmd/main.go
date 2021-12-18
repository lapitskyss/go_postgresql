package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/lapitskyss/go_postgresql/internal/http"
	"github.com/lapitskyss/go_postgresql/internal/store"
	"github.com/lapitskyss/go_postgresql/internal/tracing"
)

func main() {
	// init logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Panicf("%v", err)
	}

	// init tracer
	tracer, closer := tracing.InitJaeger("yt", logger)

	// init db connection
	str, err := store.Connect(context.Background(), os.Getenv("POSTGRES_URL"), logger)
	if err != nil {
		logger.Panic("Can not connect to database", zap.NamedError("err", err))
	}

	// start server
	handler := http.NewServer(str, logger, tracer)
	handler.Start()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case x := <-interrupt:
		logger.Info("Received a signal: " + x.String())
	case err = <-handler.Notify():
		logger.Info("Received an error from server", zap.NamedError("err", err))
	}

	// Stop http server gracefully
	err = handler.Stop()
	if err != nil {
		logger.Error("Fail to stop http server", zap.NamedError("err", err))
	}

	// close db connection
	str.Close()

	// close tracing
	err = closer.Close()
	if err != nil {
		logger.Error("Fail to stop tracer", zap.NamedError("err", err))
	}

	// sync logger
	_ = logger.Sync()
}
