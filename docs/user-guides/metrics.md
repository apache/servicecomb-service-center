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
1. **http_request_total**: The total number of received service requests.
1. **http_success_total**: Total number of requests responding to status code 2xx or 3xx.
1. **http_request_durations_microseconds**: The latency of http requests.
1. **http_query_per_seconds**: TPS of http requests.

### Pub/Sub
1. **notify_publish_total**: The total number of instance events.
1. **notify_publish_durations_microseconds**: The latency between the event generated in ServiceCenter and received by the client.
1. **notify_pending_total**: The total number of pending instances events.
1. **notify_pending_durations_microseconds**: The latency of pending instances events.
1. **notify_subscriber_total**: The total number of subscriber, e.g. Websocket, gRPC.

### Meta
1. **db_heartbeat_total**: The total number of received instance heartbeats.
1. **db_heartbeat_durations_microseconds**: The latency of received instance heartbeats.
1. **db_service_total**: The total number of micro-services.
1. **db_domain_total**: The total number of domains.
1. **db_instance_total**: The total number of instances.
1. **db_schema_total**: The total number of schemas.
1. **db_framework_total**: The total number of SDK frameworks.

### Backend
1. **db_backend_event_total**: The total number of received backend events, e.g. etcd, Mongo.
1. **db_backend_event_durations_microseconds**: The latency between received backend events and finish to build cache.
1. **db_dispatch_event_total**: The total number of dispatch events to resource handlers.
1. **db_dispatch_event_durations_microseconds**: The latency between received backend events and finish to dispatch.
1. **db_backend_operation_total**: The total number of backend requests, e.g. etcd, mongo.
1. **db_backend_operation_durations_microseconds**: The latency of backend requests.
1. **db_backend_total**: The total number of backend instances.

### System
1. **db_sc_total**: The total number of ServiceCenter instances.
1. process_resident_memory_bytes
1. process_cpu_seconds_total
1. process_cpu_usage
1. go_threads
1. go_goroutines