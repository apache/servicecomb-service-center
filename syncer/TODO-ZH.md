# 服务中心 Syncer的 待完成任务列表

[Reference to English version](./TODO.md)

## 主体功能

- 完善Service-center 插件功能，Service-center之间进行数据同步时，除Serivce和Instance外的数据进行无损同步
- 增量数据同步
- 异构支持SpringCloud Eureka，Eureka注册的微服务可与Service-center之间进行跨DC数据通信
- 异构支持Istio
- 异构支持K8S etcd
- 异构支持Consule

## 管理功能

- Syncer集群的生命周期管理，增删改查Syncer集群成员
- 跨数据中心微服务黑名单管理
- 跨数据中心微服务白名单管理
- 按需同步管理
- 支持配置文件中读取配置

## 可靠性

- 数据中心内部 Syncer 多实例生命周期管理
- Syncer的微服务实例数据和映射关系表 存储到 etcd

## 安全

- 通信加密
- 权限管理
- 鉴权管理

## 部署

- Docker部署