# ServiceComb UX

[English Docs](/ux/README.md)

## 特性

1. 提供统一的微服务管理（[ServiceComb-service-center](https://github.com/apache/servicecomb-service-center) ）、配置管理（[ServiceComb-kie](https://github.com/apache/servicecomb-kie) ）前台展示界面
1. 提供一键启动脚本，方便本地开发调试

## 如何运行
### docker方式:

依赖软件
1. docker-compose v1.29.0+, 安装指南参考 [这里](https://docs.docker.com/compose/install/#install-compose)
1. docker v18.02+, 安装指南参考 [这里](https://docs.docker.com/desktop/#download-and-install)

支持平台
 - OS系统: Windows 10 64-bit Build 18362+, Mac 10.14+, Linux
 - 浏览器: Chrome, Firefox, Safari, Edge
 

运行指令
```bash
docker-compose up -d
```

启动成功后，通过浏览器访问地址http://localhost:4200

![UX](/docs/user-guides/ux.png)

如何停止？
```bash
docker-compose down
```
