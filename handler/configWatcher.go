package handler

import (
	"fmt"
	"github.com/micro/go-micro/config"
	"github.com/micro/go-micro/config/encoder/json"
	"github.com/micro/go-micro/config/source"
	"github.com/micro/go-micro/config/source/file"
	"github.com/micro/go-micro/util/log"
	"errors"
)

var (
	configer map[string] interface{}
)

// Init 初始化配置
func ConfigWatcher(configPath string) {
	//现在先默认使用一个配置
	e := json.NewEncoder()
	fileSource := file.NewSource(
		file.WithPath(configPath),
		source.WithEncoder(e),
	)
	conf := config.NewConfig()
	// 加载micro.yml文件
	if err := conf.Load(fileSource); err != nil {
		panic(err)
	}

	// 侦听文件变动
	watcher, err := conf.Watch()
	if err != nil {
		log.Fatalf("[Init] 开始侦听应用配置文件变动 异常，%s", err)
		panic(err)
	}
	configer = conf.Map()
	fmt.Println("configer: ", configer)
	go func() {
		for {
			v, err := watcher.Next()
			if err != nil {
				log.Fatalf("[loadAndWatchConfigFile] 侦听应用配置文件变动 异常， %s", err)
				return
			}
			if err = conf.Load(fileSource); err != nil {
				panic(err)
			}
			log.Logf("[loadAndWatchConfigFile] 文件变动，%s", string(v.Bytes()))

			////本部分代码还有部分问题 1.对于底层修改、增删的部分只会认为是change
			configer=conf.Map()
		}
	}()
	return
}

//读取map的方法
func GetMapContent(m map[string]interface{}, path ...string) (interface{}, error){
	//本接口将获取一个map中，按path路径取值，返回一个interface
	var content interface{}
	var ok bool
	l := len(path)
	if l ==0 || (l == 1 && path[0]==""){  //当没有填入
		return m, nil
	}
	for k, v:= range path{
		if k ==l-1{
			content, ok = m[v]
			if !ok{
				return nil, errors.New(" 配置读取错误---> 	" + v)
			}
			return content,nil
		}
		if m, ok = m[v].(map[string]interface{}); !ok{
			return nil, errors.New(" 配置读取错误---> 	" + v)
		}
	}
	return nil, errors.New("missing map!")
}