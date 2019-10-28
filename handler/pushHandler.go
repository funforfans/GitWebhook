package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/micro/go-micro/util/log"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)
var(
	secret = "abcdefghi"  //gogs中的secret
	rawPWD = ""
	gitBaseDir= "gits" // 总目录名称
	gitPullDir = "proto"
	gitMiddleDir = "template"
	gitPushDir = "target"
	fullPathMap = make(map[string] string, 3)
)

func init()  {
	rawPWD,_ = os.Getwd()
	log.Log("init...", rawPWD)
	gitBaseDir=path.Join(rawPWD, gitBaseDir)
	dirList := []string{gitPullDir, gitMiddleDir, gitPushDir}
	for i, dir:=range dirList{
		fullPathMap[dir] = path.Join(gitBaseDir, dir)
		dirList[i] = fullPathMap[dir]
	}
	log.Log("init: map: ", fullPathMap)
	fullPathDir := dirList
	checkSourceDir(fullPathDir...)
}
func checkSourceDir(fullPathDir ...string)  {
	for _, dir:=range fullPathDir{
		if dir != ""{
			if err:=CheckDirOrCreate(dir);err!=nil{
				panic(err)
			}
		}
	}
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func GetPush(w http.ResponseWriter, r *http.Request) {
	log.Log("init...", GetPush)
	hs256 :=r.Header.Get("X-Gogs-Signature")
	strBody,body, err := GetStrandMapBody(r)
	if err != nil{
		panic(err)
	}
	myhs256 := ComputeHmacSha256(strBody, secret)
	fmt.Println(myhs256, hs256)
	if myhs256 != hs256{
		fmt.Println("sha256 hmac不一致！")
		return
	}
	//一定要加上，将工作目录切换回去
	defer func(){
		err:= os.Chdir(rawPWD)
		if err!=nil{
			panic(err)
		}
	}()
	cloneURL, err := GetMapContent(body, "repository", "clone_url")
	if err != nil{
		panic(err)
	}
	//fmt.Println(curPath, cloneURL)
	//拉取代码开始进行操作
	gitPull(cloneURL.(string), gitPullDir)
	//projectName :=strings.Split(path.Base(cloneURL.(string)), ".git")[0]
	GetFilelist(gitBaseDir)
	pushUrl :="http://git.touch4.me/xuyiwen/generate_protocol.git"
	gitPull(pushUrl, gitPushDir)
	gitsPath := path.Join(rawPWD, "gits")
	os.Chdir(gitsPath)
	log.Log("gitsPath", gitsPath)
	cmd := fmt.Sprintf("cp -R %s %s", gitMiddleDir, gitPushDir)
	if _, err :=excuteShellCommand(cmd);err!=nil{
		pwd, _ := os.Getwd()
		log.Log("pwd: ", pwd)
		log.Log("pwd: ", gitMiddleDir)
		log.Log("pwd: ", gitPushDir)
		panic(err)
		return
	}

	gitPusher(pushUrl, gitPushDir)
}

func cmdProtoc(targetPath string)  {
	var commands = make([]string, 2)
	commands = append(commands, "protoc")
	commands = append(commands, "protoc")
}

//通过cloneURL找到本地对应的仓库，并在没有此路径时git clone cloneURL， 如果冲突则删除再clone，否则直接pull
func gitPull(cloneURL,gitPullDir  string){
	defer func(){
		err:= os.Chdir(rawPWD)
		if err!=nil{
			panic(err)
		}
	}()
	//projectName :=strings.Split(path.Base(cloneURL), ".git")[0]
	//pwd, _ := os.Getwd()
	gitsPath := path.Join(gitBaseDir, gitPullDir)
	log.Log("1: ", rawPWD)
	log.Log("2: ",gitBaseDir)
	log.Log("3: ",gitPullDir)
	//下面会自动clone
	//log.Log(gitsPath)
	//err := CheckDirOrCreate(gitsPath)
	//if err!=nil{
	//	panic(err)
	//}
	if err:=os.Chdir(gitsPath);err!=nil{
		log.Log("no proto?")
		panic(err)
	}
	gitPrjPath := gitsPath//path.Join(gitsPath, projectName)
	ifExistGit, _ := PathExists(gitPrjPath)
	if !ifExistGit{
		//如果本地不存在仓库
		resp, _ := excuteShellCommand("git clone " + cloneURL)
		log.Log(gitPrjPath)
		os.Chdir(gitPrjPath)
		ifErr := strings.Index(resp, "error")
		if ifErr != -1{
			fmt.Println("发送错误")
		}
		fmt.Println(resp)
	}else {
		//如果本地已经存在仓库
		os.Chdir(gitPrjPath)
		log.Log(gitPrjPath)
		resp, _ := excuteShellCommand("git pull ")
		ifErr := strings.Index(resp, "error")
		if ifErr != -1{
			fmt.Println("发送错误")
		}
		fmt.Println(resp)
	}

}

func gitPusher(cloneURL,gitPushDir string){
	pushProjectName := strings.Split(path.Base(cloneURL), ".git")[0]
	pushPath := path.Join(rawPWD, gitBaseDir, pushProjectName)
	fmt.Println("------->pushPath: ", pushPath)
	os.Chdir(pushPath)
	defer func(){
		err:= os.Chdir(rawPWD)
		if err!=nil{
			panic(err)
		}
	}()
	//commands := []string{}
	//commands = append(commands, "git add .")
	//commands = append(commands, "git commit -m \"自动编译，提交\"")
	//commands = append(commands, "git push")
	//resps, _ := excuteShellCommands(commands)
	//log.Log("----> push resps: ", resps)


}

func excuteShellCommands(commands []string)([]string, error){
	var resps = make([]string, 1)
	for _, i := range commands{
		if i != ""{
			resp , err := excuteShellCommand(i)
			if err!=nil{
				return nil , err
			}
			resps = append(resps, resp)
		}
	}
	return resps, nil
}

func excuteShellCommand(command string) (string, error){
	fmt.Println("command: ", command)
	cmd := exec.Command("/bin/bash", "-c", command)
	bytes,err := cmd.Output()
	if err != nil {
		log.Log(err)
		return "", err
		}
	resp := string(bytes)
	log.Log("\n", resp)
	return resp, nil
}

func ComputeHmacSha256(message string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	io.WriteString(h, message)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func GetStrBody(r *http.Request) (string, error) {
	ContType  := r.Header["Content-Type"]
	var strb string
	if ContType[0] == "application/json"{
		if err:=r.ParseForm();err!=nil{
			return "", errors.New("参数解析异常")
		}
		b, err := ioutil.ReadAll(r.Body)
		//fmt.Println("b: ",(string)(b))
		if err != nil {
			return "", errors.New("连接错误")
		}
		strb = string(b)
	}
	return strb, nil
}

//返回请求r 的string, map[string]interface{} 两种类型的body
func GetStrandMapBody(r *http.Request) (string, map[string]interface{}, error){
	//将参数解析为 map[string]interface{}型
	var strb string
	ContType  := r.Header["Content-Type"]
	if ContType[0] == "application/json"{
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return "", nil, errors.New("连接错误")
		}
		var webData interface{}
		if err := json.Unmarshal(b, &webData); err!=nil{
			return "", nil, errors.New("json解析异常")
		}
		strb = string(b)
		mapdata := webData.(map[string]interface{})
		return strb, mapdata, nil
	}
	return "", nil, errors.New("请求HEADER类型错误，请检查！")
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

//检查一个dir路径，没有则会创建
func CheckDirOrCreate(dirPath string) error{
	if ifExist,err :=PathExists(dirPath); err != nil{
		return err
	}else if !ifExist{
		err1 := os.MkdirAll(dirPath, 0777)
		if err1!=nil{
			return err1
		}
	}
	return nil
}

func protoc(curPath string, fileInfo os.FileInfo, err error)  error{
	for _,v := range fullPathMap{
		checkSourceDir(v)
	}

	if path.Ext(curPath) == ".proto"{
		log.Log(fullPathMap)
		cmd := fmt.Sprintf("protoc --proto_path=%s --micro_out=%s --go_out=%s %s", fullPathMap["proto"], fullPathMap["template"], fullPathMap["template"], curPath)
		excuteShellCommand(cmd)
	}
	return err
}

func GetFilelist(path string) {
	if err := filepath.Walk(path, protoc);err!=nil{
		fmt.Println("test")
	}
}