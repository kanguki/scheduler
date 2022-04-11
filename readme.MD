# Issues

Lib leader-election of Comcast is slow to react to changes in zookeeper. I must create one myself before using it in real life :(

# Why?

When running jobs, it has to decide which node is leader first. I think it's useful when I have to run jobs in a distributed system

# How to use?

##### ENV:

envs are declared in [env.go](env.go)

##### Code:

```
import (
"github.com/kanguki/scheduler"
"github.com/kanguki/leader-election"
)
func main() {
	var scheduler = &scheduler.Driver{
		Jobs: loadJobs(),
	}
	opts := scheduler.Opts{LeOpts: leaderelection.LeOpts{Base: "NATS", Name: "test", Size: 3}}
	go scheduler.Run(opts)
	//do other stuff
}
func loadJobs() map[string]*scheduler.Job {}
```

##### Force a job to run now:

By default, it has a handler that we can call http request to do this:

```
curl http://localhost:port/runNow?job=job_name
```

To disable this feature, set `SCHEDULER_DISABLE_HTTP_HANDLER=true`

# Run

```
chmod 765 test.sh ./examples/run.sh
```

Run examples

```
./examples/run.sh
```

Run tests

```
./test.sh
```
