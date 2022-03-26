package scheduler

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	cron "github.com/robfig/cron/v3"
	log "log"
)

const (
	defaultPort                    = ":8000"
	SCHEDULER_DISABLE_HTTP_HANDLER = "SCHEDULER_DISABLE_HTTP_HANDLER"
	SCHEDULER_HTTP_PORT            = "SCHEDULER_HTTP_PORT"
)

type Driver struct {
	Cron *cron.Cron
	*httpHandler
	Jobs map[string]Job
}
type httpHandler struct {
	port string
	mux  *http.ServeMux
}
type Job struct {
	CronTime string
	Do       func()
}

func (s *Driver) Run() {
	Log("Start scheduler")
	if s.Cron == nil {
		s.Cron = cron.New(cron.WithSeconds())
	}
	for _, j := range s.Jobs {
		s.Cron.AddFunc(j.CronTime, j.Do)
	}
	s.Cron.Start()
	// defer s.cron.Stop()

	disableHttpHandler := os.Getenv(SCHEDULER_DISABLE_HTTP_HANDLER)
	disableHttpHandlerBool, _ := strconv.ParseBool(disableHttpHandler)
	if !disableHttpHandlerBool {
		if s.httpHandler == nil {
			s.httpHandler = s.NewHttpHandler()
		}
		err := http.ListenAndServe(s.port, s.mux)
		if err != nil {
			log.Fatal("error serving http handler: ", err.Error())
		}
	}
}
func (s *Driver) NewHttpHandler() *httpHandler {
	port := os.Getenv(SCHEDULER_HTTP_PORT)
	if port != "" {
		if _, e := strconv.Atoi(port); e == nil {
			port = ":" + port
		} else {
			log.Printf("invalid %v. use default port %v\n", SCHEDULER_HTTP_PORT, defaultPort)
			port = defaultPort
		}
	} else {
		log.Printf("empty %v. use default port %v\n", SCHEDULER_HTTP_PORT, defaultPort)
		port = defaultPort
	}
	log.Printf("Start http cmd handler at %v\n", port)
	mux := http.NewServeMux()
	mux.HandleFunc("/runNow", s.HandleCmd) //?job=
	return &httpHandler{port: port, mux: mux}
}
func (s *Driver) HandleCmd(w http.ResponseWriter, r *http.Request) {
	if name := r.FormValue("job"); name != "" {
		if job, ok := s.Jobs[name]; ok {
			job.Do()
			fmt.Fprintf(w, "Executed. Watch logs to check its status\n")

		} else {
			http.Error(w, "404 job name not found.", http.StatusNotFound)
		}
	} else {
		http.Error(w, "missing ?job=<job-name> in query", http.StatusBadRequest)
	}
}
