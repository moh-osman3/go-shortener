package main

import (
	"context"

	"github.com/syndtr/goleveldb/leveldb"
	"go.uber.org/zap"

	"github.com/moh-osman3/shortener"
	"github.com/moh-osman3/shortener/managers/def"
)

func main() {
	logger := zap.Must(zap.NewDevelopment())
	client, err := leveldb.OpenFile(".", nil)
	if err != nil {
		logger.Error("unable to create leveldb database", zap.Error(err))
	}
	// create and start a urlManager
	urlManager := def.NewDefaultUrlManager(logger, client)
	ctx := context.Background()
	err = urlManager.Start(ctx)
	if err != nil {
		logger.Error("error starting url manager", zap.Error(err))
		return
	}
	defer urlManager.End()

	// create and start server
	server := shortener.NewServer(urlManager, logger, "3030")
	server.AddDefaultRoutes()
	err = server.Serve()
	defer func() {
		logger.Info("shutting down server")
		server.Shutdown()
	}()

	if err != nil {
		logger.Error("error setting up server", zap.Error(err))
	}
}
