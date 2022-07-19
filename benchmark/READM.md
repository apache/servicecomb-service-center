# Benchmark
Use [k6](https://k6.io/) to call service center API
## service center and etcd deployed in local host
#### virtual user 10
|  scene | TPS  |  Latency | sc mem  | etcd mem  |
|---|---|---|---|---|
| register growing instance  |  348.244462/s  | p(90)=76.64ms p(95)=84.73ms  |  95 | 256  |
| register same instance | 1937.723719/s  |  p(90)=5.04ms  p(95)=22.23ms |   |   |
| find one instance  | 1374.776952/s  | p(90)=13.93ms p(95)=43.8ms  |  76m | 232  |
| find 100 instance  | 1838.609097/s  | p(90)=5.49ms   p(95)=21.32ms  | 72  | 196  |
| find 1000 instance  |  267.533417/s  | p(90)=75.23ms p(95)=90.71ms  |  106 | 234  |
| heartbeat one instance  |  3430.479538/s  | p(90)=1.75ms  p(95)=4.28ms  | 75m  |  195m |
## service center with embedded etcd in local host
#### virtual user 10
|  scene | TPS  |  Latency | sc mem  | etcd mem  |
|---|---|---|---|---|
| register growing instance  |  478.132249/s | p(90)=63.74ms  p(95)=74.8ms  |  206m |   |
| register same instance |2773.890508/s   |  p(90)=3.03ms  p(95)=7.62ms |  156m |   |
| find one instance  | 4182.78487/s  | p(90)=1.76ms  p(95)=3.7ms  |  175m |   |
| find 100 instance  | 1531.041088/s  | p(90)=8.62ms   p(95)=33.72ms  |  171m |   |
| find 1000 instance  | 253.041503/s  | p(90)=81.54ms p(95)=97.55ms  |  240m |   |
| heartbeat one instance  | 3232.36232/s  | p(90)=2.25ms p(95)=5.34ms  |  156m |   |