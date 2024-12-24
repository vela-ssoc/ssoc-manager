-- MySQL dump 10.14  Distrib 5.5.68-MariaDB, for Linux (x86_64)
--
-- Host: 10.205.144.12    Database: ssoc
-- ------------------------------------------------------
-- Server version	8.0.27

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

USE ssoc_test;

--
-- Table structure for table `alert_server`
--

DROP TABLE IF EXISTS `alert_server`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `alert_server` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `mode` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '发送模式',
  `name` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '名字',
  `url` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '服务器地址',
  `token` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '认证令牌',
  `account` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '咚咚账号',
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=381854756997615617 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `auth_temp`
--

DROP TABLE IF EXISTS `auth_temp`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `auth_temp` (
  `id` bigint NOT NULL COMMENT '用户 ID',
  `uid` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '临时唯一 ID',
  `created_at` datetime NOT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `broker`
--

DROP TABLE IF EXISTS `broker`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `broker` (
  `id` bigint NOT NULL,
  `name` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `servername` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `lan` json DEFAULT NULL,
  `vip` json DEFAULT NULL,
  `secret` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `status` tinyint(1) NOT NULL DEFAULT '0',
  `heartbeat_at` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `bind` varchar(22) COLLATE utf8mb4_unicode_ci NOT NULL,
  `cert_id` bigint DEFAULT NULL,
  `semver` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `broker_bin`
--

DROP TABLE IF EXISTS `broker_bin`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `broker_bin` (
  `id` bigint NOT NULL,
  `file_id` bigint NOT NULL,
  `goos` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL,
  `arch` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `size` int DEFAULT NULL,
  `hash` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `semver` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `weight` bigint DEFAULT NULL,
  `changelog` text COLLATE utf8mb4_unicode_ci,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `broker_stat`
--

DROP TABLE IF EXISTS `broker_stat`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `broker_stat` (
  `id` bigint NOT NULL,
  `name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `mem_used` bigint DEFAULT NULL,
  `mem_total` bigint DEFAULT NULL,
  `cpu_percent` float DEFAULT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `certificate`
--

DROP TABLE IF EXISTS `certificate`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `certificate` (
  `id` bigint NOT NULL COMMENT 'ID',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '名称',
  `certificate` blob NOT NULL COMMENT '证书',
  `private_key` blob NOT NULL COMMENT '私钥',
  `version` int DEFAULT NULL COMMENT '证书版本',
  `iss_country` json DEFAULT NULL COMMENT '颁发者国家',
  `iss_province` json DEFAULT NULL,
  `iss_org` json DEFAULT NULL COMMENT '颁发者组织',
  `iss_cn` text COLLATE utf8mb4_unicode_ci COMMENT 'Common Name',
  `iss_org_unit` json DEFAULT NULL COMMENT '组织单位',
  `sub_country` json DEFAULT NULL COMMENT '主题国家',
  `sub_org` json DEFAULT NULL COMMENT '主题组织',
  `sub_province` json DEFAULT NULL COMMENT '主题省份',
  `sub_cn` text COLLATE utf8mb4_unicode_ci COMMENT '主题 Common Name',
  `dns_names` json DEFAULT NULL COMMENT 'DNS Name',
  `ip_addresses` json DEFAULT NULL COMMENT 'IP',
  `email_addresses` json DEFAULT NULL COMMENT 'Email',
  `uris` json DEFAULT NULL COMMENT 'URIs',
  `not_before` datetime NOT NULL COMMENT '证书生效时间',
  `not_after` datetime NOT NULL COMMENT '证书过期时间',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='证书表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cmdb`
--

DROP TABLE IF EXISTS `cmdb`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cmdb` (
  `id` bigint NOT NULL,
  `inet` varchar(15) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '本系统的IPv4',
  `_id` bigint DEFAULT NULL COMMENT '接口未说明(应该是CMDB的ID)',
  `_org` bigint DEFAULT NULL COMMENT '接口未说明(应该是所属部门ID)',
  `_org_path` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '接口未说明(应该是所属部门全路径, 如: /运维中心/信息安全)',
  `_type` bigint DEFAULT NULL COMMENT '接口未说明',
  `agent_not_check` text COLLATE utf8mb4_unicode_ci COMMENT 'agent巡检白名单',
  `agent_version` bigint DEFAULT NULL COMMENT '客户端版本',
  `appname` varchar(510) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '应用名',
  `area` varchar(510) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '地域',
  `baoleiji_identity` varchar(510) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '堡垒机可登陆账号',
  `business_scope` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '业务作用域',
  `category` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '业务类型',
  `category_branch` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '业务分支',
  `category_zone` bigint DEFAULT NULL COMMENT '分区',
  `ci_type` varchar(510) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '接口未说明',
  `cmc_ip` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '移动IP',
  `cnc_ip` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '网通IP',
  `comment` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '作用',
  `cost_bu` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '成本所属事业部',
  `cpu` varchar(510) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'CPU型号',
  `cpu_count` bigint NOT NULL DEFAULT '0' COMMENT 'CPU数',
  `created_time` varchar(510) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'CMDB创建时间',
  `ctc_ip` varchar(510) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '电信IP',
  `env` varchar(510) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '环境',
  `float_ip` json DEFAULT NULL COMMENT '浮动IP',
  `harddisk` varchar(510) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '硬盘信息',
  `host_ip` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '宿主机IP',
  `hostname` varchar(510) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '主机名',
  `ibu` varchar(510) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '事业部',
  `idc` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'IDC',
  `ipv6` json DEFAULT NULL COMMENT 'IPv6',
  `kernel_version` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '内核版本',
  `minion_not_check` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'salt kit minion巡检白名单',
  `net_open` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '网络开放状态: 公网 仅内网',
  `nic_ip` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '网卡IP',
  `nic_mac` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '网卡MAC',
  `op_duty` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '运维负责人',
  `os_version` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '操作系统版本',
  `private_cloud_type` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '私有云类型',
  `private_ip` json DEFAULT NULL COMMENT '内网IP',
  `rack` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '机架位置',
  `ram_size` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '内存大小',
  `rd_duty` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '开发负责人',
  `security_info` text COLLATE utf8mb4_unicode_ci COMMENT '安全信息',
  `security_risk` bigint DEFAULT NULL COMMENT '安全风险值',
  `server_room` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '机房',
  `server_sn` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '宿主机序列号',
  `ssh_port` bigint DEFAULT NULL COMMENT 'ssh端口',
  `status` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '状态',
  `sys_duty` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '系统负责人',
  `unique` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '接口未说明',
  `uuid` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'UUID',
  `vserver_type` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '虚拟机类型',
  `zabbix_not_check` text COLLATE utf8mb4_unicode_ci COMMENT 'zabbix巡检白名单',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CMDB';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cmdb2`
--

DROP TABLE IF EXISTS `cmdb2`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cmdb2` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `inet` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `app_cluster` longtext COLLATE utf8mb4_unicode_ci,
  `app_duty` longtext COLLATE utf8mb4_unicode_ci,
  `appid` longtext COLLATE utf8mb4_unicode_ci,
  `appname` longtext COLLATE utf8mb4_unicode_ci,
  `auto_renew` longtext COLLATE utf8mb4_unicode_ci,
  `asset_id` longtext COLLATE utf8mb4_unicode_ci,
  `asset_status` longtext COLLATE utf8mb4_unicode_ci,
  `baoleiji_identity` longtext COLLATE utf8mb4_unicode_ci,
  `beetle_service` longtext COLLATE utf8mb4_unicode_ci,
  `billing_type` longtext COLLATE utf8mb4_unicode_ci,
  `brand` longtext COLLATE utf8mb4_unicode_ci,
  `business` longtext COLLATE utf8mb4_unicode_ci,
  `business_env` longtext COLLATE utf8mb4_unicode_ci,
  `charge_mode` longtext COLLATE utf8mb4_unicode_ci,
  `ci_type` longtext COLLATE utf8mb4_unicode_ci,
  `cmc_ip` longtext COLLATE utf8mb4_unicode_ci,
  `cnc_ip` longtext COLLATE utf8mb4_unicode_ci,
  `comment` longtext COLLATE utf8mb4_unicode_ci,
  `cost_dep_cas_id` longtext COLLATE utf8mb4_unicode_ci,
  `cpu` longtext COLLATE utf8mb4_unicode_ci,
  `cpu_count` bigint DEFAULT NULL,
  `ctc_ip` longtext COLLATE utf8mb4_unicode_ci,
  `create_date` longtext COLLATE utf8mb4_unicode_ci,
  `created_at` longtext COLLATE utf8mb4_unicode_ci,
  `created_time` longtext COLLATE utf8mb4_unicode_ci,
  `deleted` longtext COLLATE utf8mb4_unicode_ci,
  `department` longtext COLLATE utf8mb4_unicode_ci,
  `description` longtext COLLATE utf8mb4_unicode_ci,
  `device_spec` longtext COLLATE utf8mb4_unicode_ci,
  `docker_cpu_count` longtext COLLATE utf8mb4_unicode_ci,
  `env` longtext COLLATE utf8mb4_unicode_ci,
  `exp_date` longtext COLLATE utf8mb4_unicode_ci,
  `expired_at` longtext COLLATE utf8mb4_unicode_ci,
  `external_id` longtext COLLATE utf8mb4_unicode_ci,
  `float_ip` longtext COLLATE utf8mb4_unicode_ci,
  `harddisk` longtext COLLATE utf8mb4_unicode_ci,
  `host_ip` longtext COLLATE utf8mb4_unicode_ci,
  `host_type` longtext COLLATE utf8mb4_unicode_ci,
  `hostname` longtext COLLATE utf8mb4_unicode_ci,
  `host_sn` longtext COLLATE utf8mb4_unicode_ci,
  `hyper_threading` longtext COLLATE utf8mb4_unicode_ci,
  `idc` longtext COLLATE utf8mb4_unicode_ci,
  `ilo_ip` longtext COLLATE utf8mb4_unicode_ci,
  `image` longtext COLLATE utf8mb4_unicode_ci,
  `image_version` longtext COLLATE utf8mb4_unicode_ci,
  `instance_id` longtext COLLATE utf8mb4_unicode_ci,
  `ipv6` longtext COLLATE utf8mb4_unicode_ci,
  `imported_at` longtext COLLATE utf8mb4_unicode_ci,
  `in_scaling_group` longtext COLLATE utf8mb4_unicode_ci,
  `instance_type` longtext COLLATE utf8mb4_unicode_ci,
  `internet_max_bandwidth_out` bigint DEFAULT NULL,
  `k8s_cluster` longtext COLLATE utf8mb4_unicode_ci,
  `kernel_version` longtext COLLATE utf8mb4_unicode_ci,
  `logic_cpu_count` bigint DEFAULT NULL,
  `minion_not_check` longtext COLLATE utf8mb4_unicode_ci,
  `name` longtext COLLATE utf8mb4_unicode_ci,
  `namespace` longtext COLLATE utf8mb4_unicode_ci,
  `net_open` longtext COLLATE utf8mb4_unicode_ci,
  `op_duty` longtext COLLATE utf8mb4_unicode_ci,
  `op_duty_backup` longtext COLLATE utf8mb4_unicode_ci,
  `op_duty_main` longtext COLLATE utf8mb4_unicode_ci,
  `op_duty_standby` longtext COLLATE utf8mb4_unicode_ci,
  `os_arch` longtext COLLATE utf8mb4_unicode_ci,
  `os_type` longtext COLLATE utf8mb4_unicode_ci,
  `os_version` longtext COLLATE utf8mb4_unicode_ci,
  `power_states` longtext COLLATE utf8mb4_unicode_ci,
  `private_ip` longtext COLLATE utf8mb4_unicode_ci,
  `public_cloud_id` longtext COLLATE utf8mb4_unicode_ci,
  `public_cloud_idc` longtext COLLATE utf8mb4_unicode_ci,
  `private_cloud_ip` longtext COLLATE utf8mb4_unicode_ci,
  `private_cloud_type` longtext COLLATE utf8mb4_unicode_ci,
  `rack` longtext COLLATE utf8mb4_unicode_ci,
  `raid` longtext COLLATE utf8mb4_unicode_ci,
  `ram` longtext COLLATE utf8mb4_unicode_ci,
  `ram_size` longtext COLLATE utf8mb4_unicode_ci,
  `rd_duty_main` longtext COLLATE utf8mb4_unicode_ci,
  `rd_duty_member` longtext COLLATE utf8mb4_unicode_ci,
  `region` longtext COLLATE utf8mb4_unicode_ci,
  `resource_limits` longtext COLLATE utf8mb4_unicode_ci,
  `resource_requests` longtext COLLATE utf8mb4_unicode_ci,
  `security_info` longtext COLLATE utf8mb4_unicode_ci,
  `server` longtext COLLATE utf8mb4_unicode_ci,
  `server_room` longtext COLLATE utf8mb4_unicode_ci,
  `sn` longtext COLLATE utf8mb4_unicode_ci,
  `shutdown_behavior` longtext COLLATE utf8mb4_unicode_ci,
  `shutdown_mode` longtext COLLATE utf8mb4_unicode_ci,
  `status` longtext COLLATE utf8mb4_unicode_ci,
  `sys_duty` longtext COLLATE utf8mb4_unicode_ci,
  `tags` longtext COLLATE utf8mb4_unicode_ci,
  `throughput` bigint DEFAULT NULL,
  `trade_type` longtext COLLATE utf8mb4_unicode_ci,
  `update_time` longtext COLLATE utf8mb4_unicode_ci,
  `updated_at` longtext COLLATE utf8mb4_unicode_ci,
  `use` longtext COLLATE utf8mb4_unicode_ci,
  `uuid` longtext COLLATE utf8mb4_unicode_ci,
  `vcpu_count` bigint DEFAULT NULL,
  `vmem_size` bigint DEFAULT NULL,
  `vserver_type` longtext COLLATE utf8mb4_unicode_ci,
  `zabbix_not_check` longtext COLLATE utf8mb4_unicode_ci,
  `zone` longtext COLLATE utf8mb4_unicode_ci,
  PRIMARY KEY (`id`),
  UNIQUE KEY `cmdb2_pk` (`inet`)
) ENGINE=InnoDB AUTO_INCREMENT=325995175797248001 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `compound`
--

DROP TABLE IF EXISTS `compound`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `compound` (
  `id` bigint NOT NULL,
  `name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `desc` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `substances` json NOT NULL,
  `exclusion` json DEFAULT NULL,
  `version` bigint NOT NULL DEFAULT '0',
  `created_id` bigint NOT NULL,
  `updated_id` bigint NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `ding`
--

DROP TABLE IF EXISTS `ding`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ding` (
  `id` bigint NOT NULL,
  `code` char(6) COLLATE utf8mb4_unicode_ci NOT NULL,
  `tries` bigint NOT NULL DEFAULT '0',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `ding_id_uindex` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `domain`
--

DROP TABLE IF EXISTS `domain`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `domain` (
  `id` bigint NOT NULL,
  `record` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `type` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL,
  `addr` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `origin` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `isp` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `comment` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id` (`id`),
  UNIQUE KEY `domain_record_type_addr_uindex` (`record`,`type`,`addr`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `effect`
--

DROP TABLE IF EXISTS `effect`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `effect` (
  `id` bigint NOT NULL,
  `name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `submit_id` bigint NOT NULL,
  `tag` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `effect_id` bigint NOT NULL,
  `compound` tinyint(1) NOT NULL DEFAULT '0',
  `version` bigint NOT NULL DEFAULT '0',
  `enable` tinyint(1) NOT NULL DEFAULT '0',
  `exclusion` json DEFAULT NULL,
  `created_id` bigint NOT NULL,
  `updated_id` bigint NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `elastic`
--

DROP TABLE IF EXISTS `elastic`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `elastic` (
  `id` bigint NOT NULL,
  `host` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `username` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `password` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `hosts` json DEFAULT NULL,
  `desc` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `enable` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `email`
--

DROP TABLE IF EXISTS `email`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `email` (
  `id` bigint NOT NULL,
  `host` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `username` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `password` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `enable` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `emc`
--

DROP TABLE IF EXISTS `emc`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `emc` (
  `id` bigint NOT NULL,
  `name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `host` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `account` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `token` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `enable` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `event`
--

DROP TABLE IF EXISTS `event`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `event` (
  `id` bigint NOT NULL,
  `minion_id` bigint NOT NULL,
  `inet` varchar(30) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `subject` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `remote_addr` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `remote_port` int DEFAULT NULL,
  `from_code` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `typeof` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `user` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `auth` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `msg` text COLLATE utf8mb4_unicode_ci,
  `error` text COLLATE utf8mb4_unicode_ci,
  `region` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `level` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `have_read` tinyint NOT NULL DEFAULT '0',
  `send_alert` tinyint(1) NOT NULL DEFAULT '0',
  `secret` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `occur_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY `collect_event_id_uindex` (`id` DESC),
  KEY `event_minion_id_from_code_index` (`minion_id`,`from_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `gridfs_chunk`
--

DROP TABLE IF EXISTS `gridfs_chunk`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `gridfs_chunk` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `file_id` bigint NOT NULL,
  `data` mediumblob NOT NULL,
  `serial` int NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `gridfs_chunk_file_id_serial_uindex` (`file_id`,`serial`)
) ENGINE=InnoDB AUTO_INCREMENT=189057924648405078 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `gridfs_file`
--

DROP TABLE IF EXISTS `gridfs_file`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `gridfs_file` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `size` int DEFAULT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `checksum` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=189057924648325560 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `job`
--

DROP TABLE IF EXISTS `job`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `job` (
  `id` bigint NOT NULL,
  `policy_id` bigint NOT NULL,
  `policy_name` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `policy_desc` text COLLATE utf8mb4_unicode_ci,
  `code_id` bigint NOT NULL,
  `code_name` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `code_icon` blob,
  `code_desc` text COLLATE utf8mb4_unicode_ci,
  `code_hash` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `code_chunk` blob NOT NULL,
  `timeout` int NOT NULL DEFAULT '0',
  `parallel` int NOT NULL DEFAULT '0',
  `tags` json NOT NULL,
  `status` tinyint(1) NOT NULL DEFAULT '0',
  `total` int NOT NULL DEFAULT '0',
  `failed` int NOT NULL DEFAULT '0',
  `success` int NOT NULL DEFAULT '0',
  `args` json DEFAULT NULL,
  `nonce` bigint NOT NULL DEFAULT '0',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `job_code`
--

DROP TABLE IF EXISTS `job_code`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `job_code` (
  `id` bigint NOT NULL,
  `name` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `desc` text COLLATE utf8mb4_unicode_ci,
  `hash` varchar(60) COLLATE utf8mb4_unicode_ci NOT NULL,
  `icon` blob,
  `chunk` blob NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `job_policy`
--

DROP TABLE IF EXISTS `job_policy`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `job_policy` (
  `id` bigint NOT NULL,
  `name` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `desc` text COLLATE utf8mb4_unicode_ci,
  `code_id` bigint NOT NULL,
  `timeout` int NOT NULL DEFAULT '0',
  `parallel` int NOT NULL DEFAULT '0',
  `args` json NOT NULL,
  `created_id` bigint NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `kv_audit`
--

DROP TABLE IF EXISTS `kv_audit`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `kv_audit` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `minion_id` bigint NOT NULL,
  `inet` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `bucket` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `key` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `kv_audit_pk_2` (`minion_id`,`bucket`,`key`)
) ENGINE=InnoDB AUTO_INCREMENT=382943887080882177 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='kv审计表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `kv_data`
--

DROP TABLE IF EXISTS `kv_data`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `kv_data` (
  `bucket` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '存储桶',
  `key` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'key',
  `value` mediumblob COMMENT '数据',
  `count` bigint NOT NULL DEFAULT '0',
  `lifetime` bigint NOT NULL DEFAULT '0' COMMENT '生命时长',
  `expired_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '过期时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '最近修改时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `version` bigint NOT NULL DEFAULT '1',
  PRIMARY KEY (`bucket`,`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `login_lock`
--

DROP TABLE IF EXISTS `login_lock`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `login_lock` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '表 ID',
  `username` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '用户名',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '登录时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=376731444025937921 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='登录错误锁定表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `minion`
--

DROP TABLE IF EXISTS `minion`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `minion` (
  `id` bigint NOT NULL,
  `inet` varchar(15) COLLATE utf8mb4_unicode_ci NOT NULL,
  `inet6` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `mac` varchar(17) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `goos` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `arch` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `edition` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '1-未激活 2-离线 3-在线 4-已删除',
  `uptime` datetime DEFAULT NULL,
  `broker_id` bigint DEFAULT NULL,
  `broker_name` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `unstable` tinyint(1) NOT NULL DEFAULT '0' COMMENT '不稳定版本',
  `customized` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '定制版',
  `org_path` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `identity` text COLLATE utf8mb4_unicode_ci,
  `category` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `op_duty` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `comment` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `ibu` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `idc` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `unload` tinyint(1) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `minion_inet_uindex` (`inet`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `minion_account`
--

DROP TABLE IF EXISTS `minion_account`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `minion_account` (
  `id` bigint NOT NULL,
  `minion_id` bigint NOT NULL,
  `inet` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `login_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `uid` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `gid` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `home_dir` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `description` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `raw` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `minion_account_minion_id_index` (`minion_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `minion_bin`
--

DROP TABLE IF EXISTS `minion_bin`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `minion_bin` (
  `id` bigint NOT NULL,
  `file_id` bigint NOT NULL,
  `goos` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL,
  `arch` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `customized` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '定制版本字段',
  `unstable` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否是不稳定版或测试版',
  `caution` text COLLATE utf8mb4_unicode_ci COMMENT '注意事项',
  `ability` text COLLATE utf8mb4_unicode_ci COMMENT '功能作用',
  `size` int DEFAULT NULL,
  `hash` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `semver` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `weight` bigint DEFAULT NULL,
  `changelog` text COLLATE utf8mb4_unicode_ci,
  `deprecated` tinyint(1) NOT NULL DEFAULT '0',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `minion_customized`
--

DROP TABLE IF EXISTS `minion_customized`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `minion_customized` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `name` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '定制版名字',
  `icon` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '图标',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `minion_customized_pk` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=319491501457592321 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `minion_group`
--

DROP TABLE IF EXISTS `minion_group`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `minion_group` (
  `id` bigint NOT NULL,
  `minion_id` bigint NOT NULL,
  `inet` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `gid` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `description` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `minion_listen`
--

DROP TABLE IF EXISTS `minion_listen`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `minion_listen` (
  `id` bigint NOT NULL,
  `minion_id` bigint NOT NULL,
  `inet` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `record_id` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `pid` int DEFAULT NULL,
  `fd` int DEFAULT NULL,
  `family` int DEFAULT NULL,
  `protocol` int DEFAULT NULL,
  `local_ip` varchar(30) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `local_port` int DEFAULT NULL,
  `path` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `process` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `username` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `minion_listen_id_uindex` (`id` DESC),
  KEY `minion_listen_minion_id_index` (`minion_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `minion_logon`
--

DROP TABLE IF EXISTS `minion_logon`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `minion_logon` (
  `id` bigint NOT NULL,
  `minion_id` bigint NOT NULL,
  `inet` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `user` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `addr` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `msg` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `logon_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `type` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `pid` int DEFAULT NULL,
  `device` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `process` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `ignore` tinyint(1) NOT NULL DEFAULT '0',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `minion_logon_minion_id_index` (`minion_id`),
  KEY `minion_logon_msg_index` (`msg`),
  KEY `minion_logon_logon_at_index` (`logon_at` DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `minion_process`
--

DROP TABLE IF EXISTS `minion_process`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `minion_process` (
  `id` bigint NOT NULL,
  `minion_id` bigint NOT NULL,
  `inet` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `state` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `pid` bigint DEFAULT NULL,
  `ppid` bigint DEFAULT NULL,
  `pgid` bigint DEFAULT NULL,
  `cmdline` text COLLATE utf8mb4_unicode_ci,
  `username` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `cwd` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `executable` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `args` json DEFAULT NULL,
  `user_ticks` bigint DEFAULT NULL,
  `total_pct` float DEFAULT NULL,
  `total_norm_pct` float DEFAULT NULL,
  `system_ticks` bigint DEFAULT NULL,
  `total_ticks` bigint DEFAULT NULL,
  `start_time` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `mem_size` bigint DEFAULT NULL,
  `rss_bytes` bigint DEFAULT NULL,
  `rss_pct` float DEFAULT NULL,
  `share` bigint DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `checksum` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_time` datetime DEFAULT NULL,
  `modified_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `minion_process_minion_id_index` (`minion_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `minion_tag`
--

DROP TABLE IF EXISTS `minion_tag`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `minion_tag` (
  `id` bigint NOT NULL,
  `tag` varchar(15) COLLATE utf8mb4_unicode_ci NOT NULL,
  `minion_id` bigint NOT NULL,
  `kind` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `minion_tag_id_uk` (`tag`,`minion_id`),
  KEY `minion_tag_minion_id_index` (`minion_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `minion_task`
--

DROP TABLE IF EXISTS `minion_task`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `minion_task` (
  `id` bigint NOT NULL,
  `minion_id` bigint NOT NULL,
  `substance_id` bigint NOT NULL,
  `inet` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `dialect` tinyint(1) NOT NULL DEFAULT '0',
  `status` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `hash` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `link` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `from` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `failed` tinyint(1) NOT NULL DEFAULT '0',
  `cause` text COLLATE utf8mb4_unicode_ci,
  `runners` json DEFAULT NULL,
  `uptime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `minion_task_pk` (`minion_id`,`substance_id`,`name`),
  KEY `minion_task_minion_id_index` (`minion_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `notifier`
--

DROP TABLE IF EXISTS `notifier`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `notifier` (
  `id` bigint NOT NULL,
  `name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `events` json NOT NULL,
  `risks` json DEFAULT NULL,
  `ways` json DEFAULT NULL,
  `dong` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `email` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `mobile` varchar(15) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `event_code` text COLLATE utf8mb4_unicode_ci,
  `risk_code` text COLLATE utf8mb4_unicode_ci,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `oplog`
--

DROP TABLE IF EXISTS `oplog`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `oplog` (
  `id` bigint NOT NULL,
  `user_id` bigint DEFAULT NULL,
  `username` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `nickname` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `client_addr` varchar(25) COLLATE utf8mb4_unicode_ci NOT NULL,
  `direct_addr` varchar(25) COLLATE utf8mb4_unicode_ci NOT NULL,
  `method` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL,
  `path` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `query` text COLLATE utf8mb4_unicode_ci,
  `length` bigint NOT NULL DEFAULT '0',
  `content` blob,
  `cause` text COLLATE utf8mb4_unicode_ci,
  `request_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `elapsed` bigint NOT NULL DEFAULT '0',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `failed` tinyint(1) NOT NULL DEFAULT '0',
  UNIQUE KEY `id` (`id` DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户操作记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pass_dns`
--

DROP TABLE IF EXISTS `pass_dns`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `pass_dns` (
  `id` bigint NOT NULL,
  `domain` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `kind` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `before_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id` (`id`),
  UNIQUE KEY `pass_dns_kind_pk` (`domain`,`kind`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pass_ip`
--

DROP TABLE IF EXISTS `pass_ip`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `pass_ip` (
  `id` bigint NOT NULL,
  `ip` varchar(129) COLLATE utf8mb4_unicode_ci NOT NULL,
  `kind` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `before_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id` (`id`),
  UNIQUE KEY `pass_ip_kind_pk` (`ip`,`kind`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `risk`
--

DROP TABLE IF EXISTS `risk`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `risk` (
  `id` bigint NOT NULL,
  `minion_id` bigint NOT NULL,
  `inet` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `risk_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `level` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `payload` text COLLATE utf8mb4_unicode_ci,
  `subject` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `local_ip` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `local_port` int DEFAULT NULL,
  `remote_ip` varchar(40) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `remote_port` int DEFAULT NULL,
  `from_code` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `region` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `reference` text COLLATE utf8mb4_unicode_ci,
  `send_alert` tinyint(1) DEFAULT NULL,
  `have_read` tinyint(1) DEFAULT '0',
  `occur_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `secret` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` tinyint(1) DEFAULT NULL,
  `template` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `metadata` json DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `risk_dns`
--

DROP TABLE IF EXISTS `risk_dns`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `risk_dns` (
  `id` bigint NOT NULL,
  `domain` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `kind` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `origin` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `before_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id` (`id`),
  UNIQUE KEY `risk_dns_domain_kin_pk` (`domain`,`kind`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `risk_file`
--

DROP TABLE IF EXISTS `risk_file`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `risk_file` (
  `id` bigint NOT NULL,
  `checksum` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `algorithm` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL,
  `kind` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `origin` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `desc` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `before_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id` (`id`),
  UNIQUE KEY `risk_file_checksum` (`checksum`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `risk_ip`
--

DROP TABLE IF EXISTS `risk_ip`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `risk_ip` (
  `id` bigint NOT NULL,
  `ip` varchar(39) COLLATE utf8mb4_unicode_ci NOT NULL,
  `kind` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `origin` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `before_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id` (`id`),
  UNIQUE KEY `risk_ip_kind_pk` (`ip`,`kind`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sbom_component`
--

DROP TABLE IF EXISTS `sbom_component`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sbom_component` (
  `id` bigint NOT NULL,
  `minion_id` bigint NOT NULL,
  `inet` varchar(15) COLLATE utf8mb4_unicode_ci NOT NULL,
  `project_id` bigint NOT NULL,
  `filepath` varchar(510) COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `version` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `sha1` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `language` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `licenses` json DEFAULT NULL,
  `purl` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `critical_num` int NOT NULL DEFAULT '0',
  `critical_score` double NOT NULL DEFAULT '0',
  `high_num` int NOT NULL DEFAULT '0',
  `high_score` double NOT NULL DEFAULT '0',
  `medium_num` int NOT NULL DEFAULT '0',
  `medium_score` double NOT NULL DEFAULT '0',
  `low_num` int NOT NULL DEFAULT '0',
  `low_score` double NOT NULL DEFAULT '0',
  `total_num` int NOT NULL DEFAULT '0',
  `total_score` double NOT NULL DEFAULT '0',
  `status` tinyint(1) NOT NULL DEFAULT '0',
  `nonce` bigint NOT NULL DEFAULT '0',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sbom_minion`
--

DROP TABLE IF EXISTS `sbom_minion`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sbom_minion` (
  `id` bigint NOT NULL,
  `inet` varchar(15) COLLATE utf8mb4_unicode_ci NOT NULL,
  `critical_num` int NOT NULL DEFAULT '0',
  `critical_score` double NOT NULL DEFAULT '0',
  `high_num` int NOT NULL DEFAULT '0',
  `high_score` double NOT NULL DEFAULT '0',
  `medium_num` int NOT NULL DEFAULT '0',
  `medium_score` double NOT NULL DEFAULT '0',
  `low_num` int NOT NULL DEFAULT '0',
  `low_score` double NOT NULL DEFAULT '0',
  `total_num` int NOT NULL DEFAULT '0',
  `total_score` double NOT NULL DEFAULT '0',
  `nonce` bigint NOT NULL DEFAULT '0',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sbom_project`
--

DROP TABLE IF EXISTS `sbom_project`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sbom_project` (
  `id` bigint NOT NULL,
  `minion_id` bigint NOT NULL,
  `inet` varchar(15) COLLATE utf8mb4_unicode_ci NOT NULL,
  `filepath` varchar(510) COLLATE utf8mb4_unicode_ci NOT NULL,
  `size` bigint NOT NULL DEFAULT '0',
  `sha1` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `component_num` int NOT NULL DEFAULT '0',
  `pid` int NOT NULL DEFAULT '0',
  `exe` varchar(510) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `username` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `modify_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `critical_num` int NOT NULL DEFAULT '0',
  `critical_score` double NOT NULL DEFAULT '0',
  `high_num` int NOT NULL DEFAULT '0',
  `high_score` double NOT NULL DEFAULT '0',
  `medium_num` int NOT NULL DEFAULT '0',
  `medium_score` double NOT NULL DEFAULT '0',
  `low_num` int NOT NULL DEFAULT '0',
  `low_score` double NOT NULL DEFAULT '0',
  `total_num` int NOT NULL DEFAULT '0',
  `total_score` double NOT NULL DEFAULT '0',
  `nonce` bigint NOT NULL DEFAULT '0',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY `sbom_project_minion_id_filepath_uindex` (`minion_id`,`filepath`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sbom_vuln`
--

DROP TABLE IF EXISTS `sbom_vuln`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sbom_vuln` (
  `id` bigint NOT NULL,
  `vuln_id` varchar(40) COLLATE utf8mb4_unicode_ci NOT NULL,
  `purl` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `title` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` text COLLATE utf8mb4_unicode_ci,
  `score` double NOT NULL DEFAULT '0',
  `level` tinyint(1) NOT NULL DEFAULT '0',
  `vector` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `cve` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `cwe` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `reference` text COLLATE utf8mb4_unicode_ci,
  `references` json DEFAULT NULL,
  `nonce` bigint NOT NULL DEFAULT '0',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `sbom_vuln_vuln_id_purl_uindex` (`vuln_id`,`purl`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `siem_server`
--

DROP TABLE IF EXISTS `siem_server`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `siem_server` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `name` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '名字',
  `url` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '服务器地址',
  `token` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '认证令牌',
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=381854847300980737 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `startup`
--

DROP TABLE IF EXISTS `startup`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `startup` (
  `id` bigint NOT NULL,
  `node` json NOT NULL,
  `logger` json NOT NULL,
  `console` json NOT NULL,
  `extends` json NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `failed` tinyint(1) DEFAULT NULL,
  `reason` text COLLATE utf8mb4_unicode_ci,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `store`
--

DROP TABLE IF EXISTS `store`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `store` (
  `id` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `value` blob,
  `desc` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `version` bigint NOT NULL DEFAULT '0',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `escape` tinyint DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `store_id_uindex` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `substance`
--

DROP TABLE IF EXISTS `substance`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `substance` (
  `id` bigint NOT NULL,
  `name` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `icon` blob,
  `hash` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `desc` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `chunk` mediumblob NOT NULL,
  `links` json DEFAULT NULL,
  `minion_id` bigint DEFAULT NULL,
  `version` bigint NOT NULL DEFAULT '0',
  `created_id` bigint NOT NULL,
  `updated_id` bigint NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `substance_task`
--

DROP TABLE IF EXISTS `substance_task`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `substance_task` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '数据库 ID',
  `task_id` bigint NOT NULL,
  `minion_id` bigint NOT NULL COMMENT '节点 ID',
  `inet` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点 IP',
  `broker_id` bigint NOT NULL COMMENT '节点所在的 broker_id',
  `broker_name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点所在的 broker 名字',
  `failed` tinyint(1) NOT NULL COMMENT '是否下发失败',
  `reason` text COLLATE utf8mb4_unicode_ci COMMENT '如果失败，此处填写失败原因',
  `executed` tinyint(1) NOT NULL COMMENT '是否下发完毕',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '任务创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '任务更新时间',
  PRIMARY KEY (`id`),
  KEY `substance_task_broker_id_index` (`broker_id`),
  KEY `substance_task_task_id_index` (`task_id`),
  KEY `substance_task_broker_id_task_id_minion_id_index` (`broker_id`,`task_id`,`minion_id`)
) ENGINE=InnoDB AUTO_INCREMENT=379338913495568385 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sys_info`
--

DROP TABLE IF EXISTS `sys_info`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sys_info` (
  `id` bigint NOT NULL,
  `release` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `cpu_core` tinyint DEFAULT '0',
  `mem_total` bigint DEFAULT '0',
  `mem_free` bigint DEFAULT '0',
  `swap_total` bigint DEFAULT '0',
  `swap_free` bigint DEFAULT '0',
  `host_id` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `family` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `uptime` bigint DEFAULT NULL,
  `agent_total` bigint DEFAULT NULL,
  `agent_alloc` bigint DEFAULT NULL,
  `cpu_model` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `boot_at` bigint DEFAULT NULL,
  `virtual_sys` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `virtual_role` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `proc_number` bigint DEFAULT NULL,
  `hostname` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `kernel_version` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `third`
--

DROP TABLE IF EXISTS `third`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `third` (
  `id` bigint NOT NULL,
  `file_id` bigint NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `hash` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `path` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `desc` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `size` int NOT NULL,
  `customized` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `created_id` bigint NOT NULL,
  `updated_id` bigint NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `extension` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `third_file_id_uindex` (`file_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `third_customized`
--

DROP TABLE IF EXISTS `third_customized`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `third_customized` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `icon` text COLLATE utf8mb4_unicode_ci,
  `remark` text COLLATE utf8mb4_unicode_ci,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `third_customized_pk2` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=305051708736851969 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `third_tag`
--

DROP TABLE IF EXISTS `third_tag`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `third_tag` (
  `id` bigint NOT NULL,
  `tag` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `third_id` bigint NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user`
--

DROP TABLE IF EXISTS `user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `user` (
  `id` bigint NOT NULL,
  `username` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '用户名',
  `nickname` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '用户昵称',
  `dong` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '联系人咚咚号',
  `domain` tinyint(1) NOT NULL DEFAULT '0' COMMENT '帐号域：0-本地帐号 1-集团员工帐号 2-证券员工帐号。默认 0',
  `password` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '密码',
  `enable` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否禁用',
  `access_key` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `token` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'token',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间',
  `issue_at` datetime DEFAULT NULL COMMENT '签发时间',
  `session_at` datetime DEFAULT NULL COMMENT '最后一次访问时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '逻辑删除',
  `totp_bind` tinyint(1) NOT NULL DEFAULT '0',
  `totp_secret` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `user_token_index` (`token`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `vip`
--

DROP TABLE IF EXISTS `vip`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `vip` (
  `id` bigint NOT NULL,
  `virtual_ip` varchar(39) COLLATE utf8mb4_unicode_ci NOT NULL,
  `virtual_port` int NOT NULL DEFAULT '0',
  `virtual_addr` varchar(45) COLLATE utf8mb4_unicode_ci NOT NULL,
  `enable` tinyint(1) NOT NULL DEFAULT '0',
  `idc` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `member_ip` varchar(30) COLLATE utf8mb4_unicode_ci NOT NULL,
  `member_port` int NOT NULL DEFAULT '0',
  `status` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `priority` int NOT NULL DEFAULT '0',
  `biz_branch` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `biz_dept` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `biz_type` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id` (`id`),
  UNIQUE KEY `vip_virtual_ip_virtual_port_member_ip_member_port_uindex` (`virtual_ip`,`virtual_port`,`member_ip`,`member_port`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2024-12-23 17:22:41
