package main

import (
	"GitWebhook/handler"
	"github.com/micro/go-micro/util/log"
	"github.com/micro/go-micro/web"
	"github.com/micro/cli"
)

func main() {
	service := web.NewService(
		web.Name("tuyoo.micro.web.tools"),
		web.Version("latest"),
		web.Address(":8010"),
		)
	log.Log("web.NewService")
	if err := service.Init(
		web.Action(
			func(c *cli.Context) {
				// 初始化handler
				handler.Init()
			}),
			); err != nil {
		log.Fatal(err)
	}
	log.Log("service.Init()")
	service.HandleFunc("/gitupdate", handler.GetPush)
	log.Log("service.Run()")
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}

