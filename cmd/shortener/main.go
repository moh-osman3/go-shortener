package main

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"github.com/syndtr/goleveldb/leveldb"

	"github.com/moh-osman3/shortener"
	"github.com/moh-osman3/shortener/managers/def"
)


func main() {
	client, err := leveldb.OpenFile(".", nil)
	if err != nil {
		panic(err)
	}
	// create a start a urlManager
	logger := zap.Must(zap.NewDevelopment())
	urlManager := def.NewDefaultUrlManager(logger, client)
	ctx := context.Background()
	err = urlManager.Start(ctx)
	if err != nil {
		//log error
	}
	defer urlManager.End()

	// create and start server
	server := shortener.NewServer(urlManager, logger, "3030")
	server.AddDefaultRoutes()
	err = server.Serve()

	if err != nil {
		fmt.Println("error setting up server", err)
	}
	fmt.Println("server gracefully shutdown")
}