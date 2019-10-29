package handler

import (
	"path"
	"path/filepath"
	"strings"
	"testing"
)
func TestProtoc(t *testing.T)  {
	cloneURL:="http://git.touch4.me/xuyiwen/proto.git"
	projectName :=strings.Split(path.Base(cloneURL),".git")[0]
	t.Log(projectName)
	if err := filepath.Walk(path.Join(gitAbsDir, projectName), protoc);err!=nil{
		return
	}
}
