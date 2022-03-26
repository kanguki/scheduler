/**
* enable log to file
 */
package scheduler

import (
	"fmt"
	log "log"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

func Log(format string, args ...interface{}) {
	log.Printf("%s\n", fmt.Sprintf(format, args...))
}

func init() {
	if path := os.Getenv("LOG_PATH"); path != "" {
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
