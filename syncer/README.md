ServiceCenter Syncer
-------
[中文简介](./README-ZH.md)

### 1. What is ServiceCenter Syncer  
Syncer is a multiple servicecenters synchronization tool designed for large microservice architectures,supporting differ-structure servicecenters. The project adheres to the following beliefst:  
- Transparent to the application. Regardless of whether the tool is running or not, it should not affect the original microservices.  
- Provide peer-to-peer networks for multiple service-centers. Syncers are loosely coupled and members can  join and quit freely.  
- Support for different service-centers. Plug-in form, users can easily implement plug-ins for different service-centers.

##### Glossary 
- Gossip - Syncer  uses a gossip protocol by serf for inter-cluster broadcasts. The Gossip protocol uses UDP to indirectly force random propagation between cluster nodes, which is a viral propagation protocol.  
- Serf - An implementation of the Gossip protocol, a tool for cluster member management, fault detection, and orchestration, which is distributed, fault tolerant and highly available. 
- service center - It can be a single microservice discovery and registration center or a cluster of microservice discovery and registration centers. We define a service center to be a network area with unified instances data.

### 2. ServiceCenter Syncer Architecture
Syncer's runtime architecture diagram is as follows,
![image](./images/SyncerArchitecture.png?raw=true&v=1)  
There are three service-centers, 

- Within each service center, a service registry (ServiceCenter, Eurake, or other) cluster is deployed that manages all microservice instances of the service center to which it belongs, and the service centers are isolated from each other. At the same time, a Syner cluster is deployed in each service center, which is responsible for discovering instances from the registry and registering instance information from other service centers to its own service centers.
- Between multiple service centers, multiple Syncers form a peer-to-peer network that maintains a Gossip pool. The use of the Gossip protocol mainly brings the following conveniences, 
   - Syncer only needs any pool member information to join the network, and the discovery of other member information is done automatically. 
   - The fault detection of members is distributed, and multiple members in the pool cooperate with each other, which is more accurate and perfect than simple heartbeat.
   - Provides cluster messaging, mainly for event notification of instances data.
-  Syncer provides RPC service for transmitting microservice instances information. When a Syncer receives event notifications from other members or queries across service centers, the data can be synchronized through the Pull and Push interfaces provided by the RPC service.  

### 3. Quick Start
##### 3.1 Getting & Running Service center

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
$ cd syncer
$ go build
```

##### 3.3 Running ServiceCenter Syncer
###### Parameter Description
- sc-addr 

  Service center address, which is the service registry address. Cluster mode is supported, and multiple addresses are separated by commas.   
  
- bind-addr

  P2P address of Syncer for communication with other Syner members in P2P networks by Gossip.   
  
- rpc-addr

  Address of Syncer for data transmission, used to synchronize microservices data  between Syncers.  
  
- join-addr

  The address of any member of the P2P network, to enable itself join the specified P2P network, ignore this parameter when starting the first syncer on a P2P network.   
  
- mode

  Runtime mode of Syncer, specify 'cluster' for cluster mode which enables each Syncer to have three instance, default to 'single' mode.
  
- cluster-name

  Name of the Syncer cluster when mode is set to be 'cluster'.
  
- cluster-port

  The port that the Syner cluster members communicate with when mode is set to be "cluster".
  
- node

  Member name of Syncer cluster when mode is set to be "cluster".
  
  


Suppose there are 2 Service centers, each of them with a Service-center cluster for microservices discovery and registry, as following,   

|     Servicecenter      | Local address |
| :--------------------: | :-----------: |
| http://10.0.0.10:30100 |   10.0.0.10   |
| http://10.0.0.11:30100 |   10.0.0.11   |

#### Single Mode

Start Service-center Syncer to enable communication between 2 service centers

- Start Sycner  on host 10.0.0.10 

```bash
$ ./syncer daemon --sc-addr http://10.0.0.10:30100 --bind-addr 10.0.0.10:30190 --rpc-addr 10.0.0.10:30191
```

- Start Syncer on host 10.0.0.11 and join into 10.0.0.10 gossip pool

```bash
$ ./syncer daemon --sc-addr http://10.0.0.11:30100 --bind-addr 10.0.0.11:30190 --rpc-addr 10.0.0.11:30191 --join-addr 10.0.0.10:30191
```

#### Cluster Mode

Start 2 Syncer clusters on there host to synchronize microservice data between the 2 service centers

- Start Syncer cluster on host 10.0.0.10

```bash
$ ./syncer daemon --sc-addr http://10.0.0.10:30100 --bind-addr 10.0.0.10:30190 --rpc-addr 10.0.0.10:30191 --mode cluster --node syncer011 --cluster-port 30201 --join-addr 10.0.0.10:30190
$ ./syncer daemon --sc-addr http://10.0.0.10:30100 --bind-addr 10.0.0.10:30290 --rpc-addr 10.0.0.10:30291 --mode cluster --node syncer012 --cluster-port 30202 --join-addr 10.0.0.10:30190
$ ./syncer daemon --sc-addr http://10.0.0.10:30100 --bind-addr 10.0.0.10:30390 --rpc-addr 10.0.0.10:30391 --mode cluster --node syncer013 --cluster-port 30203 --join-addr 10.0.0.10:30190
```

- Start Syncer cluster on host 10.0.0.11

```bash
$ ./syncer daemon --sc-addr http://10.0.0.11:30100 --bind-addr 10.0.0.11:30190 --rpc-addr 10.0.0.11:30191 --mode cluster --node syncer021 --cluster-port 30201 --join-addr 10.0.0.10:30190
$ ./syncer daemon --sc-addr http://10.0.0.11:30100 --bind-addr 10.0.0.11:30290 --rpc-addr 10.0.0.11:30291 --mode cluster --node syncer022 --cluster-port 30202 --join-addr 10.0.0.10:30190
$ ./syncer daemon --sc-addr http://10.0.0.11:30100 --bind-addr 10.0.0.11:30390 --rpc-addr 10.0.0.11:30391 --mode cluster --node syncer023 --cluster-port 30203 --join-addr 10.0.0.10:30190
```

**Verification**  
30 seconds after registering a microservice to one of the Service-centers,  the information about it can be get from the other one.

### 4. Features

Syncer is in developing progress, reference to [TODO](./TODO.md) to get more developing features. Supported features are listed as follows,

- Data synchronization among multiple servicecomb-service-centers
- Solidify the mapping table of micro-service instances into etcd
- Support Syncer cluster mode, each Syncer has 3 instances

