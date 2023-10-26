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
