# SSOC 环境搭建及部署流程

## 环境搭建

### 数据库

> ssoc 支持 MySQL 和 OpenGauss 数据库，安装时根据情况选择其一即可。

- MySQL 8.1.0 及以上，部署完毕后创建 `ssoc` 数据库即可。

```sql
CREATE DATABASE IF NOT EXISTS ssoc;
```

- [OpenGauss 6.0.0 TLS](https://docs.opengauss.org/zh/docs/6.0.0/docs/ReleaseNotes/%E7%89%88%E6%9C%AC%E4%BB%8B%E7%BB%8D.html)
  及以上，部署完毕创建 `ssoc` 库时一定要注意兼容模式 `DBCOMPATIBILITY = 'PG'` 和字符编码 `ENCODING = 'UTF8'`。

```sql
-- 此为样例 SQL，如果 DBA 有分配角色权限等需求，请自行在创建 SQL 中添加。
CREATE
DATABASE ssoc WITH ENCODING = 'UTF8' DBCOMPATIBILITY = 'PG';
```

验证方式（OpenGauss）：

`ssoc` 库创建成功之后，查询 [PG_DATABASE](https://docs.opengauss.org/zh/docs/6.0.0/docs/DatabaseReference/PG_DATABASE.html) 表。

```sql
SELECT * FROM PG_DATABASE;
```

看看 `datcompatibility` 字段是否为 `PG`，`encoding` 是否为 `7`。

```shell
# 1. 在 gsql 命令交互中输入 \l 即可查看数据库字符编码。

# 2. 在 shell 命令行中输入下面命令：
gsql -l 
```

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

[libpcap-devel](https://www.tcpdump.org/) 编译 BPF 需要。

> Ubuntu 下可能叫：libpcap-dev

### agent

- 与 broker 连通
