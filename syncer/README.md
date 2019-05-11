ServiceCenter Syncer
-------
[中文简介](./README-ZH.md)

### 1. What is ServiceCenter Syncer  
Syncer is a multiple datacenters synchronization tool designed for large microservice architectures,supporting differ-structure datacenters.The project was born out of the belief that:  
- Transparent to the application. The synchronization tool starts and stops and should not affect the original process of the application.  
- Provide peer-to-peer networks for multiple datacenters. They are loosely coupled and members are free to join and quit.  
- Support for differ-structure datacenters. The plug-in form supports multiple datacenters drivers, making it easy for users to use custom plugins.

##### Glossary 
- Gossip - Syncer  uses a gossip protocol from serf, that broadcast messages to the cluster. Gossip uses the UDP protocol for random node-to-node propagation. It is a protocol that spreads like a virus one by one.  
- Serf - A concrete implementation of the Gossip, that  is a tool for cluster membership, failure detection, and orchestration. It is decentralized, fault-tolerant and highly available. 
- Datacenter - It can be a single microservice registry or a cluster of microservice registries. We define a datacenter to be a network area with unified instance data.

### 2. ServiceCenter Syncer Architecture
The aerial view of Syncer looks like this:  
![image](./images/SyncerArchitecture.png?raw=true)  
As shown in the figure,we can see that there are three datacenters, labeled "A" and "B" and "C". 
- Within each datacenters, we have deployed a service registry (ServiceCenter, Eurake, or other) cluster that manages all microservice instances of the datacenter to which it belongs, and the datacenters are isolated from each other. At the same time, we deployed a Syncer in each datacenters, which is responsible for discovering instances from the registry and registering instance information from other datacenters to the current datacenter.
- Between multiple datacenters, multiple Syncers form a peer-to-peer network that maintains a Gossip pool. The use of the Gossip protocol mainly brings the following conveniences:
   - Syncer only needs a single pool member information to join the network, and the discovery of other member information is done automatically;  
   - The fault detection of members is distributed, and Multiple members of the gossip pool are collaboratively completed, which is more accurate and complete than a simple heartbeat;
   - Provides cluster messaging, mainly for instance data update event notification and specifies instance cross-datacenter query.
-  Syncer provides RPC services for transmitting microservice instance information. When the Syner receives event notifications from other members or queries across datacenters, the data can be synchronized through the Pull and Push interfaces provided by the RPC service.  

### 3. Quick Start
##### 3.1 Getting & Running Service-Center
See for details: [ServiceCenter Quick Start](https://github.com/apache/servicecomb-service-center#quick-start)  

##### 3.2 Building Service-Center from source
Requirements: golang version 1.11+  
```bash
# get ServiceCenter source code from github
$ git clone https://github.com/apache/servicecomb-service-center.git $GOPATH/src/github.com/apache/servicecomb-service-center
$cd $GOPATH/src/github.com/apache/servicecomb-service-center

# Compile ServiceCenter with the "go mod" command
$ GO111MODULE=on go mod download
$ GO111MODULE=on go mod vendor
$ go build -o service-center
```

##### 3.3 Running ServiceCenter Syncer
###### Command line parameter description:
- dc-addr: Datacenter address, which is the service registry address. Cluster mode is supported, and multiple addresses are separated by an English ",".   
Example：--dc-addr http://10.0.0.10:30100,http://10.0.0.11:30100
- bind：The management communication address that Syncer listens to itself, used for gossip protocol communication.  
Example：--bind 10.0.0.10:30190
- rpc-addr：The data communication address monitored by Syncer itself, used to synchronize micro service instance data between Syner.  
Example：--rpc-addr 10.0.0.10:30191
- join：The management address of one gossip pool member.  
Example：--join 10.0.0.10:30191 

Now suppose we need to enable the Syner service in 2 ServiceCenter clusters in different datacenters. The following are the details:  
| Datacenter                | Local address |  
| :-----------------------: | :-----------: |  
| http://10.0.0.10:30100    | 10.0.0.10     |   
| http://10.0.0.11:30100    | 10.0.0.11     | 

###### Start the ServiceCenter Syner by executing the following command on the 10.0.0.10 machine：
```bash
$ ./service-center syncer --dc-addr http://10.0.0.10:30100 --bind 10.0.0.10:30190 --rpc-addr 10.0.0.10:30191
```
  
###### Start the ServiceCenter Syncer by executing the following command on the 10.0.0.10 machine and join the 10.0.0.10 gossip pool:
```bash
$ ./service-center syncer --dc-addr http://10.0.0.11:30100 --bind 10.0.0.11:30190 --rpc-addr 10.0.0.11:30191 --join 10.0.0.10:30191
```

###### Result verification  
30 seconds after registering a microservice instance to one of the ServiceCenters, you can get information about the instance from each ServiceCenter.
