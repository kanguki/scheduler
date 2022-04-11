# export SCHEDULER_HTTP_PORT=8000 
export SCHEDULER_DISABLE_HTTP_HANDLER= 
export LOG_PATH= 
export CLUSTER_SIZE=3 #must be an odd number
export LE_BASE="NATS"
# export NATS_QUORUM=

rm -f log/* &&
go run examples/main.go "$@"