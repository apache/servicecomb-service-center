ServiceCenter Syncer
-------
[中文简介](./README-ZH.md)

### 1. What is ServiceCenter Syncer  
Syncer is a multiple datacenters synchronization tool designed for large microservice architectures,supporting differ-structure datacenters. The project adheres to the following beliefst:  
- Transparent to the application. Regardless of whether the tool is running or not, it should not affect the original microservices.  
- Provide peer-to-peer networks for multiple data-centers. Syncers are loosely coupled and members can  join and quit freely.  
- Support for different data-centers. Plug-in form, users can easily implement plug-ins for different data-centers.

##### Glossary 
- Gossip - Syncer  uses a gossip protocol by serf for inter-cluster broadcasts. The Gossip protocol uses UDP to indirectly force random propagation between cluster nodes, which is a viral propagation protocol.  
- Serf - An implementation of the Gossip protocol, a tool for cluster member management, fault detection, and orchestration, which is distributed, fault tolerant and highly available. 
- Data center - It can be a single microservice discovery and registration center or a cluster of microservice discovery and registration centers. We define a data center to be a network area with unified instances data.

### 2. ServiceCenter Syncer Architecture
Syncer's runtime architecture diagram is as follows,
![image](./images/SyncerArchitecture.png?raw=true)  
There are three data-centers, labeled "A" and "B" and "C". 

- Within each data center, a service registry (ServiceCenter, Eurake, or other) cluster is deployed that manages all microservice instances of the data center to which it belongs, and the data centers are isolated from each other. At the same time, a Syner cluster is deployed in each data center, which is responsible for discovering instances from the registry and registering instance information from other data centers to its own data centers.
- Between multiple data centers, multiple Syncers form a peer-to-peer network that maintains a Gossip pool. The use of the Gossip protocol mainly brings the following conveniences, 
   - Syncer only needs any pool member information to join the network, and the discovery of other member information is done automatically. 
   - The fault detection of members is distributed, and multiple members in the pool cooperate with each other, which is more accurate and perfect than simple heartbeat.
   - Provides cluster messaging, mainly for event notification of instances data.
-  Syncer provides RPC service for transmitting microservice instances information. When a Syncer receives event notifications from other members or queries across data centers, the data can be synchronized through the Pull and Push interfaces provided by the RPC service.  

### 3. Quick Start
##### 3.1 Getting & Running Data center

Take Service-center as an example, reference to [ServiceCenter Quick Start](https://github.com/apache/servicecomb-service-center#quick-start)  

##### 3.2 Building Service-Center Syncer from Source
Requirements, Golang version 1.11+  
```bash
# Get Service-center Syncer source code from github
$ git clone https://github.com/apache/servicecomb-service-center.git $GOPATH/src/github.com/apache/servicecomb-service-center
$cd $GOPATH/src/github.com/apache/servicecomb-service-center

# Set "go mod" compile env
$ GO111MODULE=on go mod download
$ GO111MODULE=on go mod vendor

# Build it
$ go build -o service-center
```

##### 3.3 Running ServiceCenter Syncer
###### Parameter Description
- dc-addr 

  Data center address, which is the service registry address. Cluster mode is supported, and multiple addresses are separated by commas.   
  Example `--dc-addr http://10.0.0.10:30100,http://10.0.0.11:30100`

- bind

  P2P address of Syncer for communication with other Syner members in P2P networks by Gossip.   
  Example `--bind 10.0.0.10:30190`

- rpc-addr

  Address of Syncer for data transmission, used to synchronize microservices data  between Syners.  
  Example `--rpc-addr 10.0.0.10:30191`

- join

  The address of any member of the P2P network, to enable itself join the specified P2P network, ignore this parameter when starting the first syncer on a P2P network.   
  Example `--join 10.0.0.10:30191 `

Suppose there are 2 Data centers, each of them with a Service-center cluster for microservices discovery and registry, as following,   

| Datacenter                | Local address |
| :-----------------------: | :-----------: |
| http://10.0.0.10:30100    | 10.0.0.10     |
| http://10.0.0.11:30100    | 10.0.0.11     |   

Start Service-center Syncer to enable communication between 2 data centers,

**Start the ServiceCenter Syner by executing the following command on the 10.0.0.10 machine**

```bash
$ ./service-center syncer --dc-addr http://10.0.0.10:30100 --bind 10.0.0.10:30190 --rpc-addr 10.0.0.10:30191
```

**Start the ServiceCenter Syncer by executing the following command on the 10.0.0.10 machine and join the 10.0.0.10 gossip pool**
```bash
$ ./service-center syncer --dc-addr http://10.0.0.11:30100 --bind 10.0.0.11:30190 --rpc-addr 10.0.0.11:30191 --join 10.0.0.10:30191
```

**Verification**  
30 seconds after registering a microservice to one of the Service-centers,  the information about it can be get from the other one.
