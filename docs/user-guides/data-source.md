# Data Source

## etcd

Download the etcd according to your own
environment. [Installation package address](https://github.com/etcd-io/etcd/releases).

Configure app.yaml according to your needs.

```YAML
registry:
  # buildin, etcd, embeded_etcd, mongo
  kind: etcd
  # registry cache, if this option value set 0, service center can run
  # in lower memory but no longer push the events to client.
  cache:
    mode: 1
    # the cache will be clear after X, if not set cache will be never clear
    ttl:
  # enabled if registry.kind equal to etcd or embeded_etcd
  etcd:
    # the interval of etcd health check, aggregation conflict check and sync loop
    autoSyncInterval: 30s
    compact:
      # indicate how many revision you want to keep in etcd
      indexDelta: 100
      interval: 12h
    cluster:
      # if registry_plugin equals to 'embeded_etcd', then
      # name: sc-0
      # managerEndpoints: http://127.0.0.1:2380"
      # endpoints: sc-0=http://127.0.0.1:2380
      # if registry_plugin equals to 'etcd', then
      # endpoints: 127.0.0.1:2379
      endpoints: 127.0.0.1:2379
    # the timeout for failing to establish a connection
    connect:
      timeout: 10s
    # the timeout for failing to read response of registry
    request:
      timeout: 30s
```

|  field  | description  | required  | value |
|  :----  | :----  |  :----  | :---- |
| registry.kind  | database type (etcd or mongo) | yes | etcd / mongo |
| registry.cache.mode  | open cache (1 is on, 0 is off)| yes | 1 / 0 |
| registry.cache.ttl  | cache timeout (if not set cache will be never clear)| no | an integer time, like 30s/20m/10h |
| registry.etcd.autoSyncInterval | synchronization interval | yes | an integer time, like 30s/20m/10h |
| registry.etcd.compact.indexDelta | version retained in etcd | yes | a 64 bit integer, like 100 |
| registry.etcd.compact.interval | compression interval | yes | an integer time, like 30s/20m/10h |
| registry.etcd.cluster.endpoints | endpoints address | yes | string, like 127.0.0.1:2379 |
| registry.etcd.connect.timeout | the timeout for establishing a connection | yes | an integer time, like 30s/20m/10h |
| registry.etcd.request.timeout | request timeout | yes | an integer time, like 30s/20m/10h |

**Download the installation package according to the environment information**

1. Download etcd package.
2. Unzip, modify the configuration and start etcd.
3. Download the latest release from [ServiceComb Website](http://servicecomb.apache.org/release/).
4. Decompress, modify /conf/app.yaml.
5. Execute the start script to run service center

## mongo

Download the mongodb according to your own
environment.[Installation package address](https://www.mongodb.com/try/download/community).

Configure app.yaml according to your needs.

```YAML
registry:
  # buildin, etcd, embeded_etcd, mongo
  kind: mongo
  # registry cache, if this option value set 0, service center can run
  # in lower memory but no longer push the events to client.
  cache:
    mode: 1
    # the cache will be clear after X, if not set cache will be never clear
    ttl:
  mongo:
    heartbeat:
      # Mongo's heartbeat plugin
      # heartbeat.kind="checker or cache"
      # if heartbeat.kind equals to 'cache', should set cacheCapacity,workerNum and taskTimeout
      # capacity = 10000
      # workerNum = 10
      # timeout = 10
      kind: cache
      cacheCapacity: 10000
      workerNum: 10
      timeout: 10
    cluster:
      uri: mongodb://localhost:27017
      sslEnabled: false
      rootCAFile: /opt/ssl/ca.crt
      verifyPeer: false
      certFile: /opt/ssl/client.crt
      keyFile: /opt/ssl/client.key
```

|  field  | description  | required  | value |
|  :----  | :----  |  :----  | :---- |
| registry.kind  | database type (etcd or mongo) | yes | mongo / etcd |
| registry.cache.mode  | open cache (1 is on, 0 is off)| yes | 1 / 0 |
| registry.cache.ttl  | cache timeout (if not set cache will be never clear)| no | an integer time, like 30s/20m/10h |
| registry.mongo.heartbeat.kind | there are two types of heartbeat plug-ins. With cache and without cache. | yes | cache/checker |
| registry.mongo.heartbeat.cacheCapacity | cache capacity | yes | a integer, like 10000 |
| registry.mongo.heartbeat.workerNum | the number of working cooperations | yes | a integer, like 10 |
| registry.mongo.heartbeat.timeout | processing task timeout (default unit: s)| yes | a integer, like 10 |
| registry.mongo.cluster.uri | mongodb server address | yes | string, like mongodb://localhost:27017 |
| registry.mongo.cluster.sslEnabled | ssl enabled / not enabled | yes | false / true |
| registry.mongo.cluster.rootCAFile | if sslEnabled equal true, should set CA file path | yes | string, like /opt/ssl/ca.crt |
| registry.mongo.cluster.verifyPeer | insecure skip verify | yes | false / true |
| registry.mongo.cluster.certFile | the cert file path need to be set according to the configuration of mongodb server | no | string, like /opt/ssl/client.crt |
| registry.mongo.cluster.keyFile | the key file path need to be set according to the configuration of mongodb server | no | string, like /opt/ssl/client.key |

**For example, I want to run the service-center in linux environment, the database is Mongo, and the SSL/TLS
authentication is enabled for the database.**

1. Download mongodb package.
2. Unzip, modify the configuration and start mongodb.

> [configure-ssl](https://docs.mongodb.com/v4.0/tutorial/configure-ssl/)

3. Download the latest release from [ServiceComb Website](http://servicecomb.apache.org/release/).
4. Decompress, modify /conf/app.yaml.
5. Execute the start script to run service center