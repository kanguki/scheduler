package scheduler

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	le "github.com/kanguki/leader-election"
	log "github.com/kanguki/log"
	cron "github.com/robfig/cron/v3"
)

const (
	defaultPort = ":8000"
)

type Driver struct {
	Cron *cron.Cron
	*httpHandler
	Jobs map[string]*Job
	le.LE
}

type Opts struct {
	le.LeOpts
}

type httpHandler struct {
	port string
	mux  *http.ServeMux
}
type Job struct {
	Name     string
	CronTime string
	Do       func()
}

func (j *Job) AddCron(id string, cron *cron.Cron, le le.LE) {
	cron.AddFunc(j.CronTime, func() {
		if le.AmILeader() {
			log.Debug("I am leader for job %v", id)
			j.Do()
		}
	})
}

func (s *Driver) Run(opts Opts) {
	log.Log("Start scheduler")
	//init leader elector
	leaderElector, err := le.NewLE(opts.LeOpts)
	if err != nil {
		log.Log("Creating Leader elector failed. Exiting")
		os.Exit(1)
	}
	s.LE = leaderElector
	// defer s.LE.CleanResource()

	//init cron
	if s.Cron == nil {
		s.Cron = cron.New(cron.WithSeconds())
	}
	for id, j := range s.Jobs {
		j.AddCron(id, s.Cron, s.LE)
	}
	s.Cron.Start()
	// defer s.cron.Stop()

	//init handler
	disableHttpHandler, _ := strconv.ParseBool(os.Getenv(SCHEDULER_DISABLE_HTTP_HANDLER))
	if !disableHttpHandler {
		if s.httpHandler == nil {
			s.httpHandler = s.newHttpHandler()
		}
		err := http.ListenAndServe(s.port, s.mux)
		if err != nil {
			log.Log("error serving http handler: ", err.Error())
		}
	}
}
func (s *Driver) newHttpHandler() *httpHandler {
	port := os.Getenv(SCHEDULER_HTTP_PORT)
	if port != "" {
		if _, e := strconv.Atoi(port); e == nil {
			port = ":" + port
		} else {
			log.Log("invalid %v. use default port %v\n", SCHEDULER_HTTP_PORT, defaultPort)
			port = defaultPort
		}
	} else {
		log.Log("empty %v. use default port %v\n", SCHEDULER_HTTP_PORT, defaultPort)
		port = defaultPort
	}
	log.Log("Start http cmd handler at %v\n", port)
	mux := http.NewServeMux()
	mux.HandleFunc("/runNow", s.handleCmd) //?job=
	return &httpHandler{port: port, mux: mux}
}

//run in this server. no need to decide who is leader
func (s *Driver) handleCmd(w http.ResponseWriter, r *http.Request) {
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
