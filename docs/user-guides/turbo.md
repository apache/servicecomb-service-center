
# ServiceComb Turbo(experimental)
High performance service center running mode, it leverage high perfomance codec and http implemantation etc to gain better performance.
## How to enable
edit conf/chassis.yaml
```yaml
servicecomb:
  codec:
    plugin: gccy/go-json
```
## service center and etcd deployed in local host
#### Resource Consumption:
- 2 cpu cores, 4 threads
- 8 GB memory
- concurency 10

#### Topology:
service center and etcd deployed in local host

## Report
|  API| No Turbo|  Turbo|
|---|---|---|
| register growing instance  |  797/s|  861/s |
| register same instance | 5582/s |  5505/s  | 
| heartbeat one instance  |  6801/s  | 7471/s  |
| find one instance  | 7288/s| 8351/s  | 
| find 100 instance  | 3316/s  | 4106s  |  
| find 1000 instance  |  656/s  | 1038/s  | 