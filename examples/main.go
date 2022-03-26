package main

import (
	"io/ioutil"
	"path/filepath"

	"github.com/kanguki/scheduler"
	"gopkg.in/yaml.v3"
)

func main() {
	var scheduler = &scheduler.Driver{
		Jobs: loadJobs(),
	}
	go scheduler.Run()
	select {}
}

func loadJobs() map[string]scheduler.Job {
	jobs := map[string]scheduler.Job{
		"say_lala": {
			//from hour 1->19, from minute 20-40 and 42-59, from sec 0-50, run job every 3 seconds
			CronTime: "0-50/3 20-40,42-59 1-19 * * SUN-SAT",
			Do: func() {
				scheduler.Log("lala")
			},
		},
		"say_hehe": {
			CronTime: "*/5 * * * * *", //run every 5 second. (@every 5s) works the same way
			Do: func() {
				scheduler.Log("hehe")
			},
		},
	}
	absPath, _ := filepath.Abs("examples/jobs.yml")
	ymlJobs := ymlToJob(absPath)
	for _, job := range ymlJobs {
		jobs[job.Name] = scheduler.Job{
			CronTime: job.Time,
			Do: func() {
				scheduler.Log(job.Msg)
			}}
	}
	return jobs

}
func ymlToJob(path string) []YamlJob {
	yfile, err := ioutil.ReadFile(path)
	if err != nil {
		scheduler.Log(err.Error())
	}
	var data YamlJobs
	err2 := yaml.Unmarshal(yfile, &data)
	if err2 != nil {
		scheduler.Log(err2.Error())
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
