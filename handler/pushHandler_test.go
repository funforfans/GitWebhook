package handler

import (
	"github.com/go-log/log"
	"testing"
)

func TestService(t *testing.T) {
	log.Log()
	GetFilelist(gitBaseDir)
}
