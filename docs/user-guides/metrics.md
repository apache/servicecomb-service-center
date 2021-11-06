# Metrics

---

## How to export the metrics

Service-Center is compatible with the [Prometheus](https://prometheus.io/) standard.
By default, the full metrics can be collected by accessing the `/metrics` API through the `30100` port.

If you want to customize the metrics configuration.
```yaml
metrics:
  enable: true # enable to start metrics gather
  interval: 30s # the duration of collection
  exporter: prometheus # use the prometheus exporter
  prometheus:
    # optional, listen another ip-port and path if set, e.g. http://127.0.0.1:80/other
    listenURL:
```


## Summary

**FamilyName**: service_center 

### Server

|metric|type|description|
|:---|:---:|:---|
|http_request_total|counter|The total number of received service requests.|
|http_success_total|counter|Total number of requests responding to status code 2xx or 3xx.|
|http_request_durations_microseconds|summary|The latency of http requests.|
|http_query_per_seconds|gauge|TPS of http requests.|

### Pub/Sub

|metric|type|description|
|:---|:---:|:---|
|notify_publish_total|counter|The total number of instance events.|
|notify_publish_durations_microseconds|summary|The latency between the event generated in ServiceCenter and received by the client.|
|notify_pending_total|counter|The total number of pending instances events.|
|notify_pending_durations_microseconds|summary|The latency of pending instances events.|
|notify_subscriber_total|counter|The total number of subscriber, e.g. Websocket, gRPC.|

### Meta

|metric|type|description|
|:---|:---:|:---|
|db_heartbeat_total|counter|The total number of received instance heartbeats.|
|db_heartbeat_durations_microseconds|summary|The latency of received instance heartbeats.|
|db_domain_total|counter|The total number of domains.|
|db_service_total|counter|The total number of micro-services.|
|db_service_usage|counter|The usage percentage of service quota.|
|db_instance_total|counter|The total number of instances.|
|db_instance_usage|counter|The usage percentage of instances.|
|db_schema_total|counter|The total number of schemas.|
|db_framework_total|counter|The total number of SDK frameworks.|

### Backend

|metric|type|description|
|:---|:---:|:---|
|db_backend_event_total|counter|The total number of received backend events, e.g. etcd, Mongo.|
|db_backend_event_durations_microseconds|summary|The latency between received backend events and finish to build cache.|
|db_dispatch_event_total|counter|The total number of dispatch events to resource handlers.|
|db_dispatch_event_durations_microseconds|summary|The latency between received backend events and finish to dispatch.|

### System

|metric|type|description|
|:---|:---:|:---|
|db_sc_total|counter|The total number of ServiceCenter instances.|
|process_resident_memory_bytes|||
|process_cpu_seconds_total|||
|process_cpu_usage|||
|go_threads|||
|go_goroutines|||
