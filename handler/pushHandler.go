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
	rawPWD, _ = os.Getwd()
	gitAbsDir= path.Join(rawPWD, "gits") // 总目录名称
	gitMiddleDir = "template"
	configPath = "config/config.json"
)

func init()  {
	log.Log("init...", rawPWD)
	ConfigWatcher(path.Join(rawPWD, configPath))
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
	myhs256 := ComputeHmacSha256(strBody, configer["secret"].(string))
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
	gitPull(cloneURL.(string))
	projectName :=strings.Split(path.Base(cloneURL.(string)), ".git")[0]
	targetName :=projectName+"_auto_push"
	pushUrl := path.Dir(cloneURL.(string))+"/"+targetName +".git"
	log.Log(pushUrl)
	if err:=ExecMethods(gitAbsDir);err!=nil{
		return
	}
	gitPull(configer["targetUrl"].(string))
	os.Chdir(gitAbsDir)
	path.Join(rawPWD, gitMiddleDir)
	cmd := fmt.Sprintf("cp -R %s/* %s/", path.Join(rawPWD, gitMiddleDir), targetName)
	if _, err :=excuteShellCommand(cmd);err!=nil{
		panic(err)
		return
	}

	gitPusher(pushUrl)
}

func cmdProtoc(targetPath string)  {
	var commands = make([]string, 2)
	commands = append(commands, "protoc")
	commands = append(commands, "protoc")
}

//通过cloneURL找到本地对应的仓库，并在没有此路径时git clone cloneURL， 如果冲突则删除再clone，否则直接pull
func gitPull(cloneURL string){
	defer func(){
		err:= os.Chdir(rawPWD)
		if err!=nil{
			panic(err)
		}
	}()
	// 缺人项目已经创建
	CheckDirOrCreate(gitAbsDir)
	gitPullDir :=strings.Split(path.Base(cloneURL), ".git")[0]
	gitsPath := path.Join(gitAbsDir, gitPullDir)
	ifExistGit, _ := PathExists(gitsPath)
	if !ifExistGit{
		fmt.Println("git pull clone")
		if err:=os.Chdir(gitAbsDir);err!=nil{
			log.Log("no proto?")
			panic(err)
		}
		//如果本地不存在仓库
		resp, err := excuteShellCommand("git clone " + cloneURL)
		if err != nil{
			fmt.Println("----->err : ", err)
		}
		log.Log(gitsPath)
		os.Chdir(gitsPath)
		ifErr := strings.Index(resp, "error")
		if ifErr != -1{
			fmt.Println("发送错误")
		}
		fmt.Println(resp)
	}else {
		//如果本地已经存在仓库
		os.Chdir(gitsPath)
		log.Log(gitsPath)
		resp, _ := excuteShellCommand("git pull ")
		ifErr := strings.Index(resp, "error")
		if ifErr != -1{
			fmt.Println("发送错误")
		}
		fmt.Println(resp)
	}

}

func gitPusher(pushUrl string){
	pushProjectName := strings.Split(path.Base(pushUrl), ".git")[0]
	pushPath := path.Join(rawPWD, gitAbsDir, pushProjectName)
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
	if path.Ext(curPath) == ".proto"{
		cmd := fmt.Sprintf("protoc --proto_path=%s --micro_out=. --go_out=. %s", path.Dir(curPath), curPath)
		if _, err:=excuteShellCommand(cmd);err!=nil{
			log.Log(err)
			return err
		}
	}
	return nil
}

func ExecMethods(path string) error{
	if err := filepath.Walk(path, protoc);err!=nil{
		return err
	}
	return nil
}