# Service Center [![Build Status](https://travis-ci.org/ServiceComb/service-center.svg?branch=master)](https://travis-ci.org/ServiceComb/service-center)

A standalone Service Center to allow services to register their instance information and to discover providers of a given service


## Quick Start

### Getting Service Center

The easiest way to get Service Center is to use one of the pre-built release binaries which are available for OSX, Linux, Windows, and Docker. Instructions for using these binaries are on the [GitHub releases page][github-release].

[github-release]: https://github.com/servicecomb/service-center/releases/

### Running Service Center

First, you need to run a etcd(version: 1.3.x) as a database service，and then modify the etcd IP and port in the Service Center configuration file (./conf/app.conf : manager_cluster).

```sh
./bin/start.sh
```
This will bring up Service Center listening on port 30100 for service communication.

## Documentation

Project documentation is available on the ServiceComb website.


## Building

You don’t need to build from source to use Service Center (binaries on the [GitHub releases page][github-release]), but if you want to try out the latest and greatest, Service Center can be easily built.  You can refer to this compilation script([.travis.yml][travis.yml]).

[github-release]: https://github.com/servicecomb/service-center/releases/
[travis.yml]: https://github.com/ServiceComb/service-center/blob/master/.travis.yml

## Automated Testing

      
## Contact

Mailing list: 

Planning/Roadmap: milestones, roadmap

Bugs: issues

## Contributing

See CONTRIBUTING for details on submitting patches and the contribution workflow.

## Reporting Issues

See reporting bugs for details about reporting any issues.
