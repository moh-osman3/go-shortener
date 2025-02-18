package main

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"github.com/redis/go-redis/v9"

	"github.com/moh-osman3/shortener"
	"github.com/moh-osman3/shortener/managers/def"
)


func main() {
	// create db from redis quickstart docs
	client := redis.NewClient(&redis.Options{
        Addr:	  "localhost:6379",
        Password: "", // No password set
        DB:		  0,  // Use default DB
        Protocol: 2,  // Connection protocol
    })
	// create a start a urlManager
	logger := zap.Must(zap.NewDevelopment())
	urlManager := def.NewDefaultUrlManager(logger, client)
	ctx := context.Background()
	err := urlManager.Start(ctx)
	if err != nil {
		//log error
	}
	defer urlManager.End()

	// create and start server
	server := shortener.NewServer(urlManager, logger)
	server.AddDefaultRoutes()
	err = server.Serve()

	if err != nil {
		fmt.Println("error setting up server", err)
	}
	fmt.Println("server gracefully shutdown")
}