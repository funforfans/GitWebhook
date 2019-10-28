package handler

import (
	"fmt"
	"github.com/micro/go-micro/config"
	"github.com/micro/go-micro/config/encoder/json"
	"github.com/micro/go-micro/config/source"
	"github.com/micro/go-micro/config/source/file"
	"github.com/micro/go-micro/util/log"
)

var (
	err                    error
	consulConfigCenterAddr string
	configer map[string]string
	//config
)

// Init 初始化配置
func ConfigWatcher(configPath string) {
	//现在先默认使用一个配置
	configer = make(map[string]string)
	e := json.NewEncoder()
	fileSource := file.NewSource(
		file.WithPath(configPath),
		source.WithEncoder(e),
	)
	conf := config.NewConfig()
	// 加载micro.yml文件
	if err = conf.Load(fileSource); err != nil {
		panic(err)
	}

	// 侦听文件变动
	watcher, err := conf.Watch()
	if err != nil {
		log.Fatalf("[Init] 开始侦听应用配置文件变动 异常，%s", err)
		panic(err)
	}
	oldStrMap := make(map[string]string)
	oldStrMap = conf.Get().StringMap(oldStrMap)
	configer = oldStrMap
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
			strMap := make(map[string]string)
			newMapConf := v.StringMap(strMap)
			findConfDif(oldStrMap, newMapConf)
			oldStrMap = deepCopy(newMapConf)
			fmt.Println("newMapConf: ", newMapConf)
			configer = newMapConf
		}
	}()
	return
}

func findConfDif(oldConf map[string]string, newConf map[string]string)(addConf map[string]string, subConf map[string]string, changeConf map[string]string) {
	//遍历旧配置一遍查看减少的配置,和改变的配置
	addConf = make(map[string]string)
	subConf = make(map[string]string)
	changeConf = make(map[string]string)
	for key, value := range oldConf {
		if newData, ok := newConf[key]; ok{
			if newData != value{
				//在旧配置中存在却不相等的配置  changeConf
				changeConf[string(key)] = string(value)
			}
		}else{
			//旧配置中不存在的配置  subConf
			subConf[string(key)] = string(value)
		}
	}
	//遍历新配置  查看增加的配置
	for key, value := range newConf {
		//log.Log(key, ":", value)
		if _, ok := oldConf[key]; !ok{
			addConf[string(key)] = string(value)
		}
	}
	log.Log("add---------->", addConf)
	log.Log("sub---------->", subConf)
	log.Log("change------->", changeConf)
	return addConf, subConf, changeConf
}

func deepCopy(oldMap map[string]string)(newMap map[string]string ){
	//map[string]string只使用一层拷贝即可
	newMap = make(map[string]string)
	for key, value := range oldMap {
		newMap[key] = value
	}
	return newMap
}

