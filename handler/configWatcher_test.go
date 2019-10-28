package handler

import (
"testing"
)

func TestConfigWatcherService(t *testing.T) {
	ConfigWatcher("./totalConfig.json")
}
