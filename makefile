dev: 
	@ air -c .air.toml

test-bench:
	@ go test -bench . -benchmem -count=1000

test-watch: 
	@ gotestsum --watch --format testname -- -short -v