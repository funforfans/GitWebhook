package main

import (
	"GitWebhook/handler"
	"github.com/micro/go-micro/web"
	"log"
)

func main() {
	service := web.NewService(
		web.Name("tuyoo.micro.web.tools"),
		web.Version("latest"),
		web.Address(":8000"),
		)
	if err := service.Init(); err != nil {
		log.Fatal(err)
	}
	service.HandleFunc("/gitupdate", handler.GetPush)

	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}

