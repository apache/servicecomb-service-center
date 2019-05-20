ServiceCenter Syncer
-------
[Introduction to English](./README.md)

### 1. 什么是ServiceCenter Syncer  
Syncer是一个多数据中心的同步工具，专为大型微服务架构设计，支持异构数据中心。它坚持以下信念：  
- 对应用程序透明。同步工具启停，不应该对应用程序的原有流程产生影响。  
- 为多数据中心提供对等网络。他们之间是松耦合的，成员可以自由加入与退出。  
- 对异构数据中心提供支持。插件化的形式支持多种数据中心驱动，用户可便捷的接入自定义插件。

##### 名词释义 
- Gossip - Syncer使用来自serf的Gossip协议进行集群间的广播。Gossip协议使用UDP在集群节点间接力式随机传播，它是一种病毒式传播的协议。  
- Serf - Gossip协议的一种具体实现，是一种用于集群成员管理、故障检测和编排的工具，具有分布式、容错性、高可用性的特点。
- Data center - 数据中心。可以是单个微服务注册中心，也可以是一个微服务注册中心集群。我们定义的数据中心是一个拥有统一实例数据的网络区域。 

### 2. ServiceCenter Syncer架构
Syncer的运行时架构图如下：  
![image](./images/SyncerArchitecture.png?raw=true)  
如图所示，我们可以看到被标记为“A”、"B"、“C”的三个数据中心。  

- 在每个数据中心内，均部署了一套服务注册中心（ServiceCenter、Eurake或者其他）集群，该集群管理其所属数据中心的所有微服务实例，并且数据中心是彼此隔离的。同时在每个数据中心里各自部署了一个Syncer集群，它负责从注册中心发现实例，并向数据中心注册来自其他数据中心的实例信息。  
- 在多个数据中心间，多个Syncer组成对等网络，维持一个Gossip池。Gossip协议的使用主要带来了以下便捷：
   - Syncer仅需任意一个池成员信息即可加入网络，其他成员信息的发现是自动完成的；  
   - 成员的故障检测是分布式的，是池中多个成员互相协作完成的，这相对于简单的心跳更加准确、完善；  
   - 提供集群消息传递，主要用于实例数据更新事件通知、指定实例跨数据中心查询。  
- Syncer提供RPC服务，用于传输微服务实例信息。当Syncer接收到其他成员的事件通知或跨数据中心查询请求，可通过RPC服务提供的Pull和Push接口进行数据的同步。  

### 3. 快速入门 
##### 3.1 获取并启动服务中心

以Service-center为例，详细参见：[ServiceCenter 快速入门](https://github.com/apache/servicecomb-service-center#quick-start)  

##### 3.2 从源码构建和运行ServiceCenter Syncer
必要条件：golang 版本 1.11+
```bash
# 从github获取Service-center Syncer最新源码
$ git clone https://github.com/apache/servicecomb-service-center.git $GOPATH/src/github.com/apache/servicecomb-service-center
$cd $GOPATH/src/github.com/apache/servicecomb-service-center

# go mod 编译设置
$ GO111MODULE=on go mod download
$ GO111MODULE=on go mod vendor

# 编译
$ go build -o service-center
```

##### 3.3 启动ServiceCenter Syncer
###### 命令行参数说明:
- dc-addr: 数据中心地址，即服务注册中心地址。支持集群模式，多个地址间使用英文半角","隔开。    
例：--dc-addr http://10.0.0.10:30100, http://10.0.0.11:30100
- bind：Syncer在Gossip协议内P2P地址，用于在P2P网络上和其他Syncer成员之间进行通信。   
例：--bind 10.0.0.10:30190
- rpc-addr：Syncer的数据通讯地址，用于Syncer间同步微服务实例数据。  
例：--rpc-addr 10.0.0.10:30191
- join：用于指定待加入的P2P网络，P2P中的任何成员的地址， 启动第一个Syncer时不需要该参数。   
例：--join 10.0.0.10:30191  

假设有2个数据中心，每个数据中心都有一个用于微服务发现和注册的服务中心集群，如下所示：   

| Datacenter                | Local address |
| :-----------------------: | :-----------: |
| http://10.0.0.10:30100    | 10.0.0.10     |
| http://10.0.0.11:30100    | 10.0.0.11     |   

以下将启动两个Syncer来完成这两个数据中心之间的微服务数据同步:

**在10.0.0.10的机器上执行以下命令启动ServiceCenter Syncer**

```bash
$ ./service-center syncer --dc-addr http://10.0.0.10:30100 --bind 10.0.0.10:30190 --rpc-addr 10.0.0.10:30191
```

**在10.0.0.10的机器上执行以下命令启动ServiceCenter Syncer，并加入10.0.0.10的gossip池**
```bash
$ ./service-center syncer --dc-addr http://10.0.0.11:30100 --bind 10.0.0.11:30190 --rpc-addr 10.0.0.11:30191 --join 10.0.0.10:30191
```

**结果验证**  
将微服务实例注册到其中一个ServiceCener后30秒，可以从每个ServiceCenter获取有关该实例的信息。
