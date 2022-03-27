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

func loadJobs() map[string]*scheduler.Job {
	jobs := map[string]*scheduler.Job{
		"say_lala_every_3_secs": {
			//from hour 1->19, from sec 0-50, run job every 3 seconds
			CronTime: "0-50/3 1-19,20-41,42-59 0-19 * * SUN-SAT",
			Do: func() {
				scheduler.Log("lala every 3 secs")
			},
		},
		"say_hehe": {
			CronTime: "*/6 * * * * *",
			Do: func() {
				scheduler.Log("hehe every 6 secs")
			},
		},
	}
	absPath, _ := filepath.Abs("examples/jobs.yml")
	ymlJobs := ymlToJob(absPath)
	for _, job := range ymlJobs {
		scheduler.Log("%+v", job)
		jobs[job.Name] = &scheduler.Job{
			CronTime: job.Time,
			Do: func() {
				scheduler.Log(job.Msg)
			}}
	}
	scheduler.Log("jobs: %+v", jobs)
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
