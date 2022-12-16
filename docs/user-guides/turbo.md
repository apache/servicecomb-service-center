
# ServiceComb Turbo(experimental)
High performance service center running mode, 
it leverages high performance codec and http implemantation etc to gain better performance.
## How to enable
edit conf/chassis.yaml
```yaml
servicecomb:
  codec:
    plugin: gccy/go-json
  protocols:
    rest:
      listenAddress: 127.0.0.1:30106
```
edit conf/app.yaml
```yaml
server:
  turbo: true

```
## service center and etcd deployed in local host
#### Resource Consumption:
- 2 cpu cores, 4 threads
- 8 GB memory
- SSD
- concurency 10
- virtual box, ubuntu 20.04

#### Topology:
service center and etcd deployed in local host, 
even run benchmark tool in same host, 
so the performance is affected by benchmark tool

## Report
| API                       | No Turbo | Turbo  |
|---------------------------|----------|--------|
| register growing instance | 603/s    | 826/s  |
| register same instance    | 4451/s   | 7178/s | 
| heartbeat one instance    | 6121/s   | 9013/s |
| find one instance         | 6295/s   | 8748/s | 
| find 100 instance         | 2519/s   | 3751/s |  
| find 1000 instance        | 639/s    | 871/s  | 