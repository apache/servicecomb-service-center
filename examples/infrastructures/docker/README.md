# Integrate with Docker

A simple demo to deploy ServiceCenter in docker environment.

## Quick Start

```bash
cd examples/infrastructures/docker
docker-compose up
```
This will start up ServiceCenter listening on `:30100` for handling requests and Dashboard listening on `:30103`.

## Confirm ServiceCenter is Running

You can also point your browser to `http://${NODE}:30103` to view the dashboard of ServiceCenter.

## Next

ServiceCenter already integrate with Prometheus,
you can get more metrics of ServiceCenter in Grafana, [see](https://service-center.readthedocs.io/en/latest/user-guides/integration-grafana.html)