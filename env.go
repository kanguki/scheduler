/**
* declare all available env for this repo
 */
package scheduler

var (
	//scheduler
	SCHEDULER_DISABLE_HTTP_HANDLER = "SCHEDULER_DISABLE_HTTP_HANDLER" //bool. default false, enable forcing run now
	SCHEDULER_HTTP_PORT            = "SCHEDULER_HTTP_PORT"            //int. default 8000

	//elector
	//nats
	CLUSTER_SIZE = "CLUSTER_SIZE"
	LE_BASE      = "LE_BASE"     //enum: ("NATS")
	NATS_QUORUM  = "NATS_QUORUM" //comma separated. default: nats://127.0.0.1:4222. reference: https://github.com/nats-io/go-nats/blob/master/example_test.go

	//log
	LOG_PATH = "LOG_PATH" //string. log path. default log to stdout

)
