ALTER TABLE `store`
    add `escape` TINYINT(1) NULL AFTER `desc`;

-- 20230906
alter table user
    add access_key VARCHAR(200) null after enable;


alter table user
    add totp_bind TINYINT(1) default 0 not null;

alter table user
    add totp_secret VARCHAR(255) null;

alter table minion_task
    add constraint minion_task_pk
        unique (minion_id, substance_id, name);

-- --------------

alter table minion
    add unstable TINYINT(1) default 0 not null comment '不稳定版本' after broker_name;

alter table minion
    add customized VARCHAR(50) not null comment '定制版' after unstable;

alter table minion_bin
    add customized VARCHAR(255) not null comment '定制版本字段' after name;

alter table minion_bin
    add unstable TINYINT(1) default 0 not null comment '是否是不稳定版或测试版' after customized;

alter table minion_bin
    add caution TEXT null comment '注意事项' after unstable;

alter table minion_bin
    add ability TEXT null comment '功能作用' after caution;

create table minion_customized
(
    id         BIGINT auto_increment comment 'ID',
    name       VARCHAR(10)                        not null comment '定制版名字',
    icon       TEXT                               not null comment '图标',
    updated_at DATETIME default CURRENT_TIMESTAMP not null comment '更新时间',
    created_at DATETIME default CURRENT_TIMESTAMP not null comment '创建时间',
    constraint minion_customized_pk2
        primary key (id),
    constraint minion_customized_pk
        unique (name)
);

alter table third
    add customized VARCHAR(50) not null after size;

create table third_customized
(
    id         bigint auto_increment
        primary key,
    name       varchar(50)                        not null,
    icon       text                               null,
    remark     text                               null,
    created_at datetime default CURRENT_TIMESTAMP not null,
    updated_at datetime default CURRENT_TIMESTAMP not null,
    constraint third_customized_pk2
        unique (name)
);

create table kv_data
(
    bucket     varchar(255)                       not null comment '存储桶',
    `key`      varchar(255)                       not null comment 'key',
    value      mediumblob                         null comment '数据',
    count      bigint   default 0                 not null,
    lifetime   bigint   default 0                 not null comment '生命时长',
    expired_at datetime default CURRENT_TIMESTAMP not null comment '过期时间',
    updated_at datetime default CURRENT_TIMESTAMP not null comment '最近修改时间',
    created_at datetime default CURRENT_TIMESTAMP not null comment '创建时间',
    version    bigint   default 1                 not null,
    primary key (bucket, `key`)
);

create table kv_audit
(
    id         bigint auto_increment
        primary key,
    minion_id  bigint                             not null,
    inet       varchar(50)                        not null,
    bucket     varchar(255)                       not null,
    `key`      varchar(255)                       not null,
    created_at datetime default CURRENT_TIMESTAMP not null,
    updated_at datetime default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP,
    constraint kv_audit_pk_2
        unique (minion_id, bucket, `key`)
)
    comment 'kv审计表';
