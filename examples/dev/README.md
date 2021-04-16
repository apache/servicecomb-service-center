# Intro

this guide will show you how to develop service-center in your local machine.

service-center depend on etcd or mongodb.

you can get more details about service-center configuration [`here`](https://github.com/apache/servicecomb-service-center/blob/master/docs/user-guides/data-source.rst).

in this guide, we will use mongodb launched by docker compose.


# Get started

1.Build

```bash
cd examples/dev
go build github.com/apache/servicecomb-service-center/cmd/scserver
```

2.Run mongodb and service-center

```bash
docker-compose up -d
./scserver 
```
