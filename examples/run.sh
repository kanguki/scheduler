# export SCHEDULER_HTTP_PORT=8000 
export SCHEDULER_DISABLE_HTTP_HANDLER= 
export LOG_PATH= 
export SCHEDULER_ELECTOR=zk 
export ZOOKEEPER_URLS=127.0.0.1:2181 

#rm log/* &&
go run examples/main.go