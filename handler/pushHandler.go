package handler

import (
"crypto/hmac"
"crypto/sha256"
"fmt"
"github.com/micro/go-micro/util/log"
"io"
"io/ioutil"
"net/http"
"os"
"os/exec"
"path"
"errors"
)

var secret = "abcdefghi"  //gogs中的secret

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
	strBody, _ := GetStrBody(r)
	myhs256 := ComputeHmacSha256(strBody, secret)
	if myhs256 != hs256{
		fmt.Println("sha256 hmac不一致！")
		return
	}
	log.Log(strBody)
	curPath, _ := os.Getwd()

	dirPath :=path.Join(curPath, "target")
	if exist, err :=PathExists(dirPath);err!=nil && exist{
		os.RemoveAll("target/proto_go")
	}

	os.MkdirAll("target/proto_go", 0777)

}

func getCommands()[]string{
	var commands = make([]string, 2)
	commands = append(commands, "-c")
	commands = append(commands, "tree -L 1")
	//commands = append(commands, "rm test")
	return commands
}

func excuteShellCommands(commands []string){
	for _, i := range(commands){
		if i != ""{
			fmt.Println("command: ", i)
			cmd := exec.Command("/bin/bash", "-c", i)//commands[2:4]...)
			bytes,err := cmd.Output()
			if err != nil {
				log.Log(err)
			}
			resp := string(bytes)
			log.Log("\n", resp)
		}
	}
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
