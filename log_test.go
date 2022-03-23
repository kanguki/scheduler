package scheduler

import (
	"testing"

	log "log"
)

func TestLog(t *testing.T) {
	InitLog("log/app.log")
	for i := 0; i < 10; i++ {
		log.Println("info")
	}
}
