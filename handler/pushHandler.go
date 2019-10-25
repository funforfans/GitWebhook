package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/micro/go-micro/util/log"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"encoding/json"
	"strings"
	"path"
)

var secret = "abcdefghi"  //gogs中的secret
var rawPWD = ""


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
	curPath, _ := os.Getwd()
	if rawPWD=="" {
		rawPWD = curPath
	}
	//一定要加上，将工作目录切换回去
	defer func(){
		os.Chdir(rawPWD)
	}()
	cloneURL, err := GetMapContent(body, "repository", "clone_url")
	if err != nil{
		panic(err)
	}
	fmt.Println(curPath, cloneURL)
	//拉取代码开始进行操作
	gitPull(cloneURL.(string))

	//dirPath :=path.Join(curPath, "target")
	//if exist, err :=PathExists(dirPath);err!=nil && exist{
	//	os.RemoveAll("target/proto_go")
	//}
	//os.MkdirAll("target/proto_go", 0777)
	//
}

//通过cloneURL找到本地对应的仓库，并在没有此路径时git clone cloneURL， 如果冲突则删除再clone，否则直接pull
func gitPull(cloneURL string){
	s :=strings.Split(cloneURL, "/")
	name := s[len(s)-1]
	name = name[0:len(name)-4]
	fmt.Println(name)
	pwd, _ := os.Getwd()
	gitsPath := path.Join(pwd, "gits")
	err := CheckDirOrCreate(gitsPath)
	if err!=nil{
		panic(err)
	}
	gitPrjPath := path.Join(gitsPath, name)
	ifExistGit, _ := PathExists(gitPrjPath)
	if !ifExistGit{
		//如果本地不存在仓库
		resp := excuteShellCommand("git clone " + cloneURL)
		os.Chdir(gitPrjPath)
		ifErr := strings.Index(resp, "error")
		if ifErr != -1{
			fmt.Println("发送错误")
		}
		fmt.Println(resp)
	}else {
		//如果本地已经存在仓库
		os.Chdir(gitPrjPath)
		resp := excuteShellCommand("git pull ")
		ifErr := strings.Index(resp, "error")
		if ifErr != -1{
			fmt.Println("发送错误")
		}
		fmt.Println(resp)
	}

}

func getCommands()[]string{
	var commands = make([]string, 2)
	commands = append(commands, "-c")
	commands = append(commands, "tree -L 1")
	//commands = append(commands, "rm test")
	return commands
}

func excuteShellCommands(commands []string)[]string{
	var resps = make([]string, 1)
	for _, i := range(commands){
		if i != ""{
			resp := excuteShellCommand(i)
			resps = append(resps, resp)
		}
	}
	return resps
}

func excuteShellCommand(command string)string{
	fmt.Println("command: ", command)
	cmd := exec.Command("/bin/bash", "-c", command)
	bytes,err := cmd.Output()
	if err != nil {
		log.Log(err)
		}
	resp := string(bytes)
	log.Log("\n", resp)
	return resp
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
	fmt.Println(r.Method)
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