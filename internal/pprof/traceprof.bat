curl http://localhost:8000/debug/pprof/trace?seconds=40 -o trace.out
go tool trace -http ":8083" trace.out
