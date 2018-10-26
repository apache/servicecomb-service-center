scctl
========

`scctl` is a command line client for service center.

## Global options

- `addr` the http host and port of service center, can be overrode by env `HOSTING_SERVER_IP`.
- `ca` the CA file path  to access service center, can be overrode by `$SSL_ROOT`/trust.cer.
- `cert` the certificate file path to access service center, can be overrode by `$SSL_ROOT`/server.cer.
- `key` the key file path to access service center, can be overrode by `$SSL_ROOT`/server_key.pem.
- `pass` the passphase string to decrypt key file.
- `pass-file` the passphase file path to decrypt key file, can be overrode by `$SSL_ROOT`/cert_pwd.
- `timeout` the maximum time allowed for the request.

## Get commands

The `get` command is the root command for getting any type resources from service center. 

### service [options]

Get the microservices list from service center. `service` command can be instead of `svc`.

#### Options

- `domain`(d) print the information under the specified domain in service center, print under the `default` domain by default.
- `output`(o) support mode `wide`, output the complete microservices information(e.g., framework, endpoints).
- `all-domains` print the information under all domains in service center.

#### Examples

```bash
./scctl get svc
#       NAME      |   APPID   | VERSIONS |     ENV     |               FRAMEWORK                | AGE  
# +---------------+-----------+----------+-------------+----------------------------------------+-----+
#   SERVICECENTER | default   | 0.0.1    | development |                                        | 9d   
#   Client        | springmvc | 1.0.0    |             |                                        | 9d   
#   consumer      | springmvc | 0.0.1    |             | servicecomb-java-chassis-CSE:2.3.35... | 9d   
#   provider      | springmvc | 0.0.1    |             | servicecomb-java-chassis-CSE:2.3.35... | 9d

./scctl get svc -owide
#   DOMAIN  |     NAME      |   APPID   | VERSIONS |     ENV     |                         FRAMEWORK                          |        ENDPOINTS        | AGE  
# +---------+---------------+-----------+----------+-------------+------------------------------------------------------------+-------------------------+-----+
#   default | Client        | springmvc | 1.0.0    |             |                                                            |                         | 9d   
#   default | consumer      | springmvc | 0.0.1    |             | servicecomb-java-chassis-CSE:2.3.35;ServiceComb:1.1.0.B006 |                         | 9d   
#   default | provider      | springmvc | 0.0.1    |             | servicecomb-java-chassis-CSE:2.3.35;ServiceComb:1.1.0.B006 |                         | 9d   
#   default | SERVICECENTER | default   | 0.0.1    | development |                                                            | rest://127.0.0.1:30100/ | 9d

./scctl get svc -d test
#    NAME  |  APPID  |  VERSIONS   | ENV | FRAMEWORK | AGE  
# +--------+---------+-------------+-----+-----------+-----+
#   Server | default | 1.0.0,1.0.1 |     |           | 1d
```

### instance [options]

Get the instances list from service center. `instance` command can be instead of `inst`.

#### Options

- `domain`(d) domain name, return `default` domain microservices list by default.
- `output`(o) support mode `wide`, return the complete microservices information(e.g., framework, endpoints).
- `all-domains` return all domains microservices information.

#### Examples
```bash
./scctl get inst
#       HOST     |        ENDPOINTS        | VERSION |    SERVICE    |  APPID  | LEASE | AGE  
# +--------------+-------------------------+---------+---------------+---------+-------+-----+
#   desktop-0001 | rest://127.0.0.1:30100/ | 0.0.1   | SERVICECENTER | default | 2m    | 11m 

./scctl get inst -owide
#   DOMAIN  |     HOST     |        ENDPOINTS        | VERSION |    SERVICE    |  APPID  |     ENV     | FRAMEWORK | LEASE | AGE  
# +---------+--------------+-------------------------+---------+---------------+---------+-------------+-----------+-------+-----+
#   default | desktop-0001 | rest://127.0.0.1:30100/ | 0.0.1   | SERVICECENTER | default | development |           | 2m    | 17m

./scctl get inst -d default
#       HOST     |        ENDPOINTS        | VERSION |    SERVICE    |  APPID  | LEASE | AGE  
# +--------------+-------------------------+---------+---------------+---------+-------+-----+
#   desktop-0001 | rest://127.0.0.1:30100/ | 0.0.1   | SERVICECENTER | default | 2m    | 18m
```


### schema [options]

Get the schemas content from service center.

#### Options

- `domain`(d) domain name, return all microservice schema contents from `default` domain by default.
- `app` the application name of microservice.
- `name` the name of microservice.
- `version` the semantic version of microservice.
- `save-dir`(s) the directory to save the schema content,
the schema file path structure follows the rule:
`{save-dir}/schemas/[{domain}/][{project}/][{env}/]{app}/{microservice}.{version}/{schemaId}.yaml` 
- `all-domains` return all microservice schema contents from all domains.

#### Examples
```bash
# save schemas to files
./scctl get schema -s .
#  2 / 2 [============================================================] 100.00% 0s
# Finished.
ls -l schemas/springmvc/provider.v0.0.1
# total 4
# -rw-r----- 1 ubuntu ubuntu 610 Sep 15 23:05 say.yaml

# print schema to console
./scctl get schema --name provider
# ---
# swagger: "2.0"
# info:
#   version: "1.0.0"
#   title: "swagger definition for com.service.provider.controller.ProviderImpl"
#   x-java-interface: "cse.gen.springmvc.provider.provider.ProviderImplIntf"
# basePath: "/say"
# consumes:
# - "application/json"
# produces:
# - "application/json"
# paths:
#   /helloworld:
#     get:
#       operationId: "helloworld"
#       produces:
#       - "application/json"
#       parameters:
#       - name: "name"
#         in: "query"
#         required: false
#         type: "string"
#       responses:
#         200:
#           description: "response of 200"
#           schema:
#             type: "string"
```

### cluster [options]

Get the registry clusters managed by service center.

#### Examples
```bash
./scctl get cluster
#   CLUSTER |        ENDPOINTS         
# +---------+-------------------------+
#   sc-1    | http://172.0.1.32:30100  
#   sc-0    | http://172.0.1.29:30100
```

## Diagnose commands

The `diagnose` command can output the service center health report. 
If the service center is isolated from etcd, the diagnosis will print wrong information.

#### Options

- `etcd-addr` the http addr and port of etcd endpoints
- `etcd-ca` the CA file path  to access etcd, can be overrode by env `$SSL_ROOT`/trust.cer.
- `etcd-cert` the certificate file path to access etcd, can be overrode by env `$SSL_ROOT`/server.cer.
- `etcd-key` the key file path to access etcd, can be overrode by env `$SSL_ROOT`/server_key.pem.
- `etcd-pass` the passphase string to decrypt key file.
- `etcd-pass-file` the passphase file path to decrypt key file, can be overrode by env `$SSL_ROOT`/cert_pwd.

#### Examples
```bash
./scctl diagnose
echo exit $?
# exit 0

./scctl diagnose
echo exit $?
# 1. found in etcd but not in cache, details:
#   service: [springmvc/Client/1.0.0 springmvc/consumer/0.0.1 springmvc/provider/0.0.1 default/SERVICECENTER/0.0.1 default/Server/1.0.0 default/Server/1.0.1]
#   instance: [[rest://127.0.0.1:30100/]]
# error: 1. found in etcd but not in cache
# exit 1
```

## Health Check commands

The `health` command can check the service center health. 

#### Exit codes

- `0` the service center is healthy.
- `1` an internal error occurred in scctl.
- `2` the service center is unavailable.
- `3` the service center is abnormal.

#### Examples
```bash
./scctl health
# Registry service is unavailable(invoke request failed: Get http://127.0.0.1:30100/v4/default/registry/health: dial tcp 127.0.0.1:30100: getsockopt: connection refused)

echo exit $?
# exit 2
```