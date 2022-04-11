package main

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	leaderelection "github.com/kanguki/leader-election"
	"github.com/kanguki/log"
	sc "github.com/kanguki/scheduler"
	"gopkg.in/yaml.v3"
)

func main() {
	flag.Parse()
	var scheduler = &sc.Driver{
		Jobs: loadJobs(),
	}
	clusterSize, err := strconv.Atoi(os.Getenv(sc.CLUSTER_SIZE))
	if err != nil {
		log.Log("Error parsing %v", sc.CLUSTER_SIZE)
		os.Exit(1)
	}
	leBase := leaderelection.Base(os.Getenv(sc.LE_BASE))
	opts := sc.Opts{LeOpts: leaderelection.LeOpts{Base: leBase, Name: "test", Size: clusterSize, TimeoutDecideLeader: 10}}
	go scheduler.Run(opts)
	select {}
}

func loadJobs() map[string]*sc.Job {
	jobs := map[string]*sc.Job{
		"say_lala_every_3_secs": {
			//from hour 1->19, from sec 0-50, run job every 3 seconds
			CronTime: "0-50/3 1-19,20-41,42-59 0-19 * * SUN-SAT",
			Do: func() {
				log.Log("lala every 3 secs")
			},
		},
		"say_hehe": {
			CronTime: "*/6 * * * * *",
			Do: func() {
				log.Log("hehe every 6 secs")
			},
		},
	}

	absPath, _ := filepath.Abs("examples/jobs.yml")
	ymlJobs := ymlToJob(absPath)
	log.Debug("yamlJobs %v", ymlJobs)
	for _, v := range ymlJobs {
		msg := v.Msg
		jobs[v.Name] = &sc.Job{
			CronTime: v.Time,
			Do: func() {
				log.Log("%v", msg) //if we directly use v.Msg instead of setting msg and pass here, it will use msg of last item. maybe because goroutine effect here upon array :D
			},
		}
	}
	log.Log("jobs: %+v", jobs)
	return jobs

}
func ymlToJob(path string) []YamlJob {
	yfile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Log(err.Error())
	}
	var data YamlJobs
	err2 := yaml.Unmarshal(yfile, &data)
	if err2 != nil {
		log.Log(err2.Error())
	}
	return data.Jobs
}

type YamlJobs struct {
	Jobs []YamlJob `yaml:"jobs"`
}
type YamlJob struct {
	Name string `yaml:"name"`
	Time string `yaml:"time"`
	Msg  string `yaml:"msg"`
}
