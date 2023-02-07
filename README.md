# Apache-ServiceComb-Service-Center
[![Build Status](https://www.travis-ci.org/apache/servicecomb-service-center.svg?branch=master)](https://www.travis-ci.org/apache/servicecomb-service-center)  [![Coverage Status](https://coveralls.io/repos/github/apache/servicecomb-service-center/badge.svg?branch=master)](https://coveralls.io/github/apache/servicecomb-service-center?branch=master)  [![Go Report Card](https://goreportcard.com/badge/github.com/apache/servicecomb-service-center)](https://goreportcard.com/report/github.com/apache/servicecomb-service-center) [![GoDoc](https://godoc.org/github.com/apache/servicecomb-service-center?status.svg)](https://godoc.org/github.com/apache/servicecomb-service-center)  [![HitCount](http://hits.dwyl.io/apache/servicecomb-service-center.svg)](http://hits.dwyl.io/apache/servicecomb-service-center) [![Gitter](https://img.shields.io/badge/ServiceComb-Gitter-ff69b4.svg)](https://gitter.im/ServiceCombUsers/Lobby)  

Apache ServiceComb Service-Center is a Restful based service-registry that provides micro-services discovery and micro-service management. 
It is based on Open API format and provides features like service-discovery, 
fault-tolerance, dynamic routing, notify subscription and scalable by design. 
It has high performance cache design and separate entity management for micro-services and their instances. 
It provides out of box support for metrics and tracing. It has a web portal to manage the micro-services.  

# Components
- server: a http server which provide restful API
- scctl: CLI of server
- syncer: alpha feature. multiple cluster synchronization tool, designed for large microservice architectures

## Features
 - **[`Open API`](/docs/openapi/v4.yaml)**: API doc(Open API format) management for microservice
 - **Metadata**: Metadata management for both microservice and microservice instance
 - **Separated**: Separated microservice and microservice instance entity management
 - **Domains**: Logical multiple domains management
 - **[`Security`](https://service-center.readthedocs.io/en/latest/user-guides/rbac.html)**: Role base access control for service discovery
 - **Discovery**: Support query instance by criteria
 - **Subscribe**: Use web socket to notify client about instance change events
 - **[`Portal`](/ux)**: Awesome web portal
 - **Fault tolerance**: Multiple fault tolerance mechanism and design in the architecture
 - **Performance**: Performance/Caching design
 - **[`Metrics`](https://service-center.readthedocs.io/en/latest/user-guides/integration-grafana.html)**: Able to expose Prometheus metric API automatically, [see](/docs/user-guides/metrics.md)
 - **[`Tracing`](https://service-center.readthedocs.io/en/latest/plugins-tracing/tracing.html)**: Able to report tracing data to Zipkin server
 - **[`Pluginable`](https://service-center.readthedocs.io/en/latest/design-guides/plugin.html)**: Able to load custom authentication, tls and other dynamic libraries
 - **[`CLI`](https://service-center.readthedocs.io/en/latest/intro/scctl.html)**: Easy to control service center
 - **[`Kubernetes`](https://service-center.readthedocs.io/en/latest/dev-guides/kubeclusters.html)**: Embrace kubernetes ecosystem and support multi cluster service discovery
 - **[`Datacenters`](https://service-center.readthedocs.io/en/latest/dev-guides/multidcs.html)**: Additional layer of abstraction to clusters deployed in multiple datacenters
 - **[`Aggregation`](https://service-center.readthedocs.io/en/latest/design-guides/design.html)**: Able to aggregate microservices from multiple registry platforms and
    support platform registry and client side registry at the same time
 - **[`FastRegister`](https://service-center.readthedocs.io/en/latest/user-guides/fast-registration.html)**: Fast register instance to service center

## Documentation

Project documentation is available on the [ServiceComb website][servicecomb-website]. You can also find full document [`here`](https://service-center.readthedocs.io/).

[servicecomb-website]: http://servicecomb.apache.org/

## Quick Start

### Getting Service Center

The easiest way to get Service Center is to use one of the pre-built release binaries which are available for Linux, Windows and Docker.

[github-release]: http://servicecomb.apache.org/release/

### Build docker image

```sh
sudo bash scripts/docker/build-image/build.sh
```
it builds an image servicecomb/service-center

## Get The Latest Release

[Download Service Center](http://servicecomb.apache.org/release/service-center-downloads/)

## Client

- [go](https://github.com/go-chassis/sc-client)

## Ecosystem

- [Go-Chassis](https://github.com/go-chassis/go-chassis): default registry
- [Kratos](https://github.com/go-kratos/kratos): [registry plugin](https://github.com/go-kratos/kratos/tree/main/contrib/registry/servicecomb)
- [CloudWeGo-kitex](https://github.com/cloudwego/kitex): [registry plugin](https://github.com/kitex-contrib/registry-servicecomb)

## Contact

Bugs: [issues](https://issues.apache.org/jira/browse/SCB)

## Contributing

See [Contribution guide](/CONTRIBUTING.md) for details on submitting patches and the contribution workflow.

## Reporting Issues

See reporting bugs for details about reporting any issues.

## License
Licensed under an [Apache 2.0 license](LICENSE).

## Export Notice

This distribution includes cryptographic software. The country in which you currently reside may have restrictions on the import, possession, use, and/or re-export to another country, of encryption software. BEFORE using any encryption software, please check your country's laws, regulations and policies concerning the import, possession, or use, and re-export of encryption software, to see if this is permitted. See <http://www.wassenaar.org/> for more information.

The Apache Software Foundation has classified this software as Export Commodity Control Number (ECCN) 5D002, which includes information security software using or performing cryptographic functions with asymmetric algorithms. The form and manner of this Apache Software Foundation distribution makes it eligible for export under the "publicly available" Section 742.15(b) exemption (see the BIS Export Administration Regulations, Section 742.15(b)) for both object code and source code.

The following provides more details on the included cryptographic software:
  * golang crypto https://github.com/golang/go/tree/master/src/crypto
  * The transport between server and client can be configured to use TLS  
