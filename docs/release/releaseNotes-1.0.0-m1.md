# Release Notes

### Apache ServiceComb Service-Center (incubating) version 1.0.0-m1
#### API Changes :
 * Added new API to get All Schema List.
 * Add Service statistics in the Governance API.
 * Add Self-microservice information in the Governance API.

#### New Features/Improvements:
* Support discovery of SC instances by Consumer micro-service.
* Event driven implementation for dependency rules.
* Make compact interval configurable and avoid defragmentation of the database when compacted.
* Update the default quota’s limit of service/instance count.
* Update black-list rule controls in discovery.

#### Metrics : 
* Added support for Prometheus metrics exposure.
* Added templates for Grafana Dashboard.
            
#### Optimization:
* Optimized Restful clients and plugins loader.
* Optimized Service-Count calculation rule.
* Use CDN for resolving all the dependencies of frontend.

#### Bug-Fix:
* Fix panic issue while deleting instance and invalid metrics request.
* Fix modify schema response issue and heart-beat failure when etcd has no leader.
* Fix batch delete api to exempt from unregistering service-center microservice.
* Fix watcher wrong event sequence when SC modify resource concurrently
* Fix discovery of default APP services in Shared service mode 
