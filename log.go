package scheduler

import (
	"gopkg.in/natefinch/lumberjack.v2"
	log "log"
)

func InitLog(path string) {
	if path != "" {
		l := &lumberjack.Logger{
			Filename: path,
			MaxSize:  1,    // megabytes
			MaxAge:   7,    //days
			Compress: true, // disabled by default
		}

		l.Rotate()
		log.SetOutput(l)
	}
}
