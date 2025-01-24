# 安装

## 环境搭建

### 数据库

普通版本：MySQL 8.1.0 及以上。

信创版本：OpenGauss 5.0.3 LTS 及以上，__强烈推荐__ 用最新 TLS 版。

### Elasticsearch

Elasticsearch 7.x 及以上。

## manager

### 网络

- 与数据库连通

- 与 ES 连通（前端界面查看日志时需要）

- 与咚咚接口连通（告警、扫码登录需要）

### 依赖软件

- [graphviz](https://graphviz.org/) 性能分析火焰图需要（如果未用到该功能可忽略）。

### 安装步骤

#### 部署程序

1. 创建安装目录：`mkdir -p /vdb/ssoc/manager`
2. 解压 [ssoc-manager.zip](ssoc-manager.zip) 到安装目录
3. 修改 ${安装目录}/resources/config/manager.yaml 中相关配置参数
4. 创建软链接：`ln -s ssoc-manager-yyyyMMdd ssoc-manager`

#### 安装服务

1. 将 ssoc-manager.service 拷贝到 /etc/systemd/system 目录
2. 刷新服务：`systemctl daemon-reload`
3. 运行服务：`systemctl start ssoc-manager.service`
4. 开机自启：`systemctl enable ssoc-manager.service`

## broker

### 网络打通

- 与 manager 连通

- 与数据库连通

- 与 ES 连通（对 agent 上报的日志存储）

- 与咚咚接口连通（咚咚告警）

### 依赖软件

[libpcap-devel](https://www.tcpdump.org/) 需要编译 agent 的 BPF 策略。

### agent

- 与 broker 连通
