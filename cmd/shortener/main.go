package main

import (
	"context"
	"fmt"
	"github.com/moh-osman3/shortener"
	"github.com/moh-osman3/shortener/managers/def"
)


func main() {
	// create a start a urlManager
	urlManager := def.NewUrlManager()
	ctx := context.Background()
	err := urlManager.Start(ctx)
	if err != nil {
		//log error
	}
	defer urlManager.End()

	// create and start server
	server := shortener.NewServer(urlManager)
	server.AddDefaultRoutes()
	err = server.Serve()

	if err != nil {
		fmt.Println("error setting up server", err)
	}
	fmt.Println("server gracefully shutdown")
}