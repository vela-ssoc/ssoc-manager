alter table user
    modify domain tinyint(1) default 0 not null comment '帐号域：0-本地帐号 1-sso 帐号';

UPDATE user
SET domain = 0
WHERE domain = 1;

UPDATE user
SET domain = 1
WHERE domain = 2;



