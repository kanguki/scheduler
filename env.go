/**
* declare all available env for this repo
 */
package scheduler

const (
	//scheduler
	SCHEDULER_DISABLE_HTTP_HANDLER = "SCHEDULER_DISABLE_HTTP_HANDLER" //bool. default false, enable forcing run now
	SCHEDULER_HTTP_PORT            = "SCHEDULER_HTTP_PORT"            //int. default 8000

	//elector
	SCHEDULER_ELECTOR = "SCHEDULER_ELECTOR" //string. enum: "zk". leave empty if this is ran as a single node.
	ZOOKEEPER_URLS    = "ZOOKEEPER_URLS"    //string. used for zookeeper-based elector. comma separated. eg: 127.0.0.1:2181,127.0.0.1:2182

	//log
	LOG_PATH = "LOG_PATH" //string. log path. default log to stdout
)
