ALTER TABLE `store`
    add `escape` TINYINT(1) NULL AFTER `desc`;

-- 20230906
alter table user add access_key VARCHAR(200) null after enable;