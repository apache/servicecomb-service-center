ServiceCenter Syncer
-------
[Introduction to English](./README.md)

### 1. 什么是ServiceCenter Syncer  
Syncer是一个多服务中心的同步工具，专为大型微服务架构设计，支持异构服务中心。它坚持以下信念：  
- 对应用程序透明。同步工具启停，不应该对应用程序的原有流程产生影响。  
- 为多服务中心提供对等网络。他们之间是松耦合的，成员可以自由加入与退出。  
- 对异构服务中心提供支持。插件化的形式支持多种服务中心驱动，用户可便捷的接入自定义插件。

##### 名词释义 
- Gossip - Syncer使用来自serf的Gossip协议进行集群间的广播。Gossip协议使用UDP在集群节点间接力式随机传播，它是一种病毒式传播的协议。  
- Serf - Gossip协议的一种具体实现，是一种用于集群成员管理、故障检测和编排的工具，具有分布式、容错性、高可用性的特点。
- Service center - 服务中心中心。可以是单个微服务注册发现中心，也可以是一个微服务注册中心集群。我们定义的服务中心是一个拥有统一实例数据的网络区域。 

### 2. ServiceCenter Syncer架构
Syncer的运行时架构图如下：  
![image](./images/SyncerArchitecture.png?raw=true&v=1)  
如图所示，  

- 在每个服务中心内，均部署了一套服务注册中心（ServiceCenter、Eurake或者其他）集群，该集群管理其所属服务中心的所有微服务实例，并且是彼此隔离的。同时在每个服務中心里各自部署了一个Syncer集群，它负责从注册中心发现实例，并向服务中心注册来自其他服务中心的实例信息。  
- 在多个服务中心间，多个Syncer组成对等网络，维持一个Gossip池。Gossip协议的使用主要带来了以下便捷：
   - Syncer仅需任意一个池成员信息即可加入网络，其他成员信息的发现是自动完成的；  
   - 成员的故障检测是分布式的，是池中多个成员互相协作完成的，这相对于简单的心跳更加准确、完善；  
   - 提供集群消息传递，主要用于实例数据更新事件通知、指定实例跨服务中心查询。  
- Syncer提供RPC服务，用于传输微服务实例信息。当Syncer接收到其他成员的事件通知或跨服务中心查询请求，可通过RPC服务提供的Pull和Push接口进行数据的同步。  

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
$ cd syncer
$ go build
```

##### 3.3 启动ServiceCenter Syncer
###### 命令行参数说明:
- sc-addr: 服务中心地址，即服务注册中心地址。支持集群模式，多个地址间使用英文半角","隔开。    
- bind-addr：Syncer在Gossip协议内P2P地址，用于在P2P网络上和其他Syncer成员之间进行通信。   
- rpc-addr：Syncer的数据通讯地址，用于Syncer间同步微服务实例数据。  
- join-addr：用于指定待加入的P2P网络，P2P中的任何成员的地址， 启动第一个Syncer时不需要该参数。   
- mode: 运行模式配置，当前支持两种模式，单实例模式“single”和集群模式“cluster”, cluster模式下每个syncer拥有三个实例
- cluster-name：当mode为”cluster“时，Syncer集群的名字
- cluster-port： 当mode为“cluster”时，Syncer集群成员之间进行通信的端口
- node：当mode为“cluster”时，syncer集群成员名称。


假设有2个服务中心，每个服务中心都有一个用于微服务发现和注册的服务中心集群，如下所示：   

|     Service center     | Local address |
| :--------------------: | :-----------: |
| http://10.0.0.10:30100 |   10.0.0.10   |
| http://10.0.0.11:30100 |   10.0.0.11   |

#### 单例模式（Single Mode）

分别在两个机器上启动两个syncer来完成这两个服务中心之间的微服务数据同步

- 在10.0.0.10的机器上执行以下命令启动Syncer

```bash
$ ./syncer daemon --sc-addr http://10.0.0.10:30100 --bind-addr 10.0.0.10:30190 --rpc-addr 10.0.0.10:30191
```

- 在10.0.0.11的机器上执行以下命令启动Syncer，并加入10.0.0.10的gossip池

```bash
$ ./syncer daemon --sc-addr http://10.0.0.11:30100 --bind-addr 10.0.0.11:30190 --rpc-addr 10.0.0.11:30191 --join-addr 10.0.0.10:30191
```

#### 集群模式（Cluster Mode）

分别在两个机器上启动两个syncer集群来完成这两个服务中心之间的微服务数据同步

- 在10.0.0.10的机器上启动syncer集群

```bash
$ ./syncer daemon --sc-addr http://10.0.0.10:30100 --bind-addr 10.0.0.10:30190 --rpc-addr 10.0.0.10:30191 --mode cluster --node syncer011 --cluster-port 30201 --join-addr 10.0.0.10:30190
$ ./syncer daemon --sc-addr http://10.0.0.10:30100 --bind-addr 10.0.0.10:30290 --rpc-addr 10.0.0.10:30291 --mode cluster --node syncer012 --cluster-port 30202 --join-addr 10.0.0.10:30190
$ ./syncer daemon --sc-addr http://10.0.0.10:30100 --bind-addr 10.0.0.10:30390 --rpc-addr 10.0.0.10:30391 --mode cluster --node syncer013 --cluster-port 30203 --join-addr 10.0.0.10:30190
```

- 在10.0.0.11机器上启动syncer集群

```bash
$ ./syncer daemon --sc-addr http://10.0.0.11:30100 --bind-addr 10.0.0.11:30190 --rpc-addr 10.0.0.11:30191 --mode cluster --node syncer021 --cluster-port 30201 --join-addr 10.0.0.10:30190
$ ./syncer daemon --sc-addr http://10.0.0.11:30100 --bind-addr 10.0.0.11:30290 --rpc-addr 10.0.0.11:30291 --mode cluster --node syncer022 --cluster-port 30202 --join-addr 10.0.0.10:30190
$ ./syncer daemon --sc-addr http://10.0.0.11:30100 --bind-addr 10.0.0.11:30390 --rpc-addr 10.0.0.11:30391 --mode cluster --node syncer023 --cluster-port 30203 --join-addr 10.0.0.10:30190
```

**结果验证**  
将微服务实例注册到其中一个ServiceCener后30秒，可以从每个ServiceCenter获取有关该实例的信息。

### 4. 特性

Syncer是一个开发中版本，在下面列出已支持的特性，更多开发中的特性请参考[TODO](./TODO-ZH.md)

- 支持多个servicecomb-service-center 服务中心之间进行数据同步
- 在etcd中固化存储微服务实例映射表
- 支持集群模式部署Syncer，每个syncer集群拥有3个实例