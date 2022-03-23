export SCHEDULER_HTTP_PORT=8000 #default 8000
export SCHEDULER_DISABLE_HTTP_HANDLER=false #default false, enable forcing run now
export LOG_PATH=log/app.log #default log to stdout

rm log/* &&
go run examples/main.go