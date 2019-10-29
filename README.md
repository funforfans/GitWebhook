# GitWebhook
```
一个通用的git钩子服务器,通过将git的webhook钩子URL设定为服务器的监听地址
```
[钩子配置教程](https://jingyan.baidu.com/article/5d6edee228c88899ebdeec47.html)

## 项目结构
```
.
├── README.md
├── config              存放配置文件 config.json
├── gits                项目运行后自动创建的目录，存放所有中间目录
├── go.mod
├── go.sum
├── handler             项目的处理方法模块
└── main.go
```

## 项目功能
```
1.监听绑定至本服务URL的git项目
2.接受到git项目推送后，拉取最新的内容至gits文件夹下
3.将该项目中的.proto文件生成对应的micro.go、pb.go文件，按原来的目录层级存放在gits/generate_protocol项目下，generate_protocol为配置中的targetUrl对应的项目名
```

## 使用说明
1.填写配置文件 config/config.json
```
{
  "secret" : "abcdefghi",
  "_secret" : "本地与github钩子界面协议校验的秘钥，可自行设置",
  "targetUrl": "http://git.touch4.me/your_git/generate_protocol.git",
  "_targetUrl": "将原项目处理后的文件打包后，转存的统一git项目"
}
```

2.启动项目  
① 可自行设置服务名、监听端口
```
service := web.NewService(
	web.Name("tuyoo.micro.web.tools"),
	web.Version("latest"),
	web.Address(":8010"),
	)
```
②可自行设置服务名、监听端口以及路由绑定
```
service.HandleFunc("/gitupdate", handler.GetPush)
```

## 注意事项
```
1.正式使用项目前需要确保json文件中targetUrl项目拥有 拉取 上传权限以及配置完整，即config.json文件中targetUrl
2.git hook中设置的URL能够路由至本服务
```

## TODO
```
1.将本项目加入微服务集合
2.统一使用consul配置模块
3.拓展类似项目中protoc方法、自动执行、编译、配置检测等
```