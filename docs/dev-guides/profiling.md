# Profiling
service center integrated pprof 
## Configuration
```yaml
server:
  pprof:
    mode: 1
```

## Run pprof
```sh
go tool pprof http://localhost:30100/debug/pprof/profile?seconds=30
```