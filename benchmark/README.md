# Benchmark
Use [k6](https://k6.io/) to call service center API
## How to use scripts
Install k6
```shell
k6 run --vus 10 --duration 30s register/register_random_service.js
```
## service center and etcd deployed in local host
Resource Consumption:
- 2 cpu cores
- 8 GB memory

k6, service center, etcd all in one VM
#### virtual user 10
|  scene | TPS  |  Latency | sc mem  | etcd mem  |
|---|---|---|---|---|
| register growing instance  |  699/s  | p(90)=24.02ms  p(95)=32.26ms  |  83 | 283  |
| register same instance | 5069/s  |  p(90)=3.72ms  p(95)=5.62ms | 70  | 123  |
| heartbeat one instance  |  7068/s  | p(90)=2.66ms  p(95)=4.28ms  | 52m  |  145m |
| find one instance  | 7577/s  | p(90)=2.42ms  p(95)=3.9ms  |  51m | 144  |
| find 100 instance  | 3242/s  | p(90)=6.7ms    p(95)=10.01ms  | 72  | 196  |
| find 1000 instance  |  544/s  | p(90)=75.23ms p(95)=90.71ms  |  106 | 234  |

## service center with embedded etcd in local host
#### virtual user 10
|  scene | TPS  |  Latency | sc mem  | etcd mem  |
|---|---|---|---|---|
| register growing instance  |  561/s | p(90)=19.11ms p(95)=26.7ms  |  330m |   |
| register same instance |5042/s   |  p(90)=3.69ms  p(95)=5.69ms |  156m |   |
| heartbeat one instance  | 6513/s  | p(90)=2.94ms  p(95)=4.73ms  |  156m |   |
| find one instance  | 6580/s  | p(90)=2.81ms  p(95)=4.61ms  |  175m |   |
| find 100 instance  | 2843/s  |  p(90)=7.82ms   p(95)=11.62ms  |  200m |   |
| find 1000 instance  | 546/s  | p(90)=28.69ms p(95)=36.78ms  |  200m |   |



