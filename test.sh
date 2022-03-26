export SCHEDULER_HTTP_PORT=
export SCHEDULER_DISABLE_HTTP_HANDLER=
export LOG_PATH=

#-short to run tests not connecting to network
#-coverprofile testCoverage.out
rm -f log/* && go clean -testcache && go test $(go list ./... | grep -v /examples) "$@"
