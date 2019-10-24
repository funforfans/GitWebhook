package main

import (
	"GitWebhook/handler"
	"fmt"
	"github.com/micro/go-micro/web"
	"log"
	"os"
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
	err := os.Chdir("/Users/tugame/gitPRJ/goku-api-gateway")
	if err != nil{
		fmt.Println("chdir: ", err)
	}

	//调用shell指令测试
	//var commands = getCommands()
	//excuteShellCommands(commands)
	//fmt.Println("-----> ", commands)
	//md5.Sum()
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}

